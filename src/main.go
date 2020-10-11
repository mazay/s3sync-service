//
// s3sync-service - Realtime S3 synchronisation tool
// Copyright (c) 2020  Yevgeniy Valeyev
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
//

package main

import (
	"flag"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	version string

	osExit = os.Exit

	wg = sync.WaitGroup{}

	inK8s = isInK8s()

	sizeMetric = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "s3sync",
			Subsystem: "data",
			Name:      "total_size",
			Help:      "Total size of the data in S3",
		},
		[]string{"local_path", "bucket", "bucket_path", "site"},
	)

	objectsMetric = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "s3sync",
			Subsystem: "data",
			Name:      "objects_count",
			Help:      "Number of objects in S3",
		},
		[]string{"local_path", "bucket", "bucket_path", "site"},
	)

	errorsMetric = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "s3sync",
			Subsystem: "errors",
			Name:      "count",
			Help:      "Number of errors",
		},
		[]string{"local_path", "bucket", "bucket_path", "site", "scope"},
	)
)

// UploadCFG - structure for the upload queue
type UploadCFG struct {
	s3Service *s3.S3
	file      string
	site      Site
	action    string
}

// ChecksumCFG - structure for the checksum comparison queue
type ChecksumCFG struct {
	UploadCFG      UploadCFG
	filename       string
	checksumRemote string
	site           Site
}

func isInK8s() bool {
	_, present := os.LookupEnv("KUBERNETES_SERVICE_HOST")
	if present {
		logger.Debugln("Look mom, I'm in k8s!")
	}
	return present
}

func main() {
	var configpath string
	var metricsPort string
	var metricsPath string

	// Read command line args
	flag.StringVar(&configpath, "config", "config.yml", "Path to the config.yml")
	flag.StringVar(&metricsPort, "metrics-port", "9350", "Prometheus exporter port, 0 to disable the exporter")
	flag.StringVar(&metricsPath, "metrics-path", "/metrics", "Prometheus exporter path")
	flag.Parse()

	// Read config file
	config := readConfigFile(configpath)

	// init logger
	initLogger(config)

	// Start prometheus exporter
	if metricsPort != "0" {
		go prometheusExporter(metricsPort, metricsPath)
	}

	// Set global WatchInterval
	if config.WatchInterval == 0 {
		config.WatchInterval = 1000
	}

	// Set global S3OpsRetries
	if config.S3OpsRetries == 0 {
		config.S3OpsRetries = 5
	}

	// Init upload worker
	if config.UploadWorkers == 0 {
		config.UploadWorkers = 10
	}

	uploadCh := make(chan UploadCFG, config.UploadQueueBuffer)
	logger.Infof("starting %s upload workers", strconv.Itoa(config.UploadWorkers))
	for x := 0; x < config.UploadWorkers; x++ {
		go uploadWorker(uploadCh)
	}

	// Init checksum checker workers
	if config.ChecksumWorkers == 0 {
		// If in k8s then run 2 workers
		// otherwise run 2 workers per core
		if inK8s {
			config.ChecksumWorkers = 2
		} else {
			config.ChecksumWorkers = runtime.NumCPU() * 2
		}
	}

	checksumCh := make(chan ChecksumCFG)
	logger.Infof("starting %s checksum workers", strconv.Itoa(config.ChecksumWorkers))
	for x := 0; x < config.ChecksumWorkers; x++ {
		go checksumWorker(checksumCh, uploadCh)
	}

	// Start separate thread for each site
	wg.Add(len(config.Sites))
	for _, site := range config.Sites {
		// Remove leading slash from the BucketPath
		site.BucketPath = strings.TrimLeft(site.BucketPath, "/")
		// Set site name
		if site.Name == "" {
			site.Name = site.Bucket + "/" + site.BucketPath
		}
		// Set site AccessKey
		if site.AccessKey == "" {
			site.AccessKey = config.AccessKey
		}
		// Set site SecretAccessKey
		if site.SecretAccessKey == "" {
			site.SecretAccessKey = config.SecretAccessKey
		}
		// Set site BucketRegion
		if site.BucketRegion == "" {
			site.BucketRegion = config.AwsRegion
		}
		// Set default value for StorageClass
		if site.StorageClass == "" {
			site.StorageClass = "STANDARD"
		}
		// Set site WatchInterval
		if site.WatchInterval == 0 {
			site.WatchInterval = config.WatchInterval
		}
		// Set site S3OpsRetries
		if site.S3OpsRetries == 0 {
			site.S3OpsRetries = config.S3OpsRetries
		}
		go syncSite(site, uploadCh, checksumCh)
	}
	wg.Wait()
}

func prometheusExporter(metricsPort string, metricsPath string) {
	http.Handle(metricsPath, promhttp.Handler())
	http.ListenAndServe(":"+metricsPort, nil)
}

func uploadWorker(uploadCh <-chan UploadCFG) {
	for cfg := range uploadCh {
		if cfg.action == "upload" {
			uploadFile(cfg.s3Service, cfg.file, cfg.site)
		} else if cfg.action == "delete" {
			deleteFile(cfg.s3Service, cfg.file, cfg.site)
		} else {
			logger.Errorf("programming error, unknown action: %s", cfg.action)
		}
	}
}

func checksumWorker(checksumCh <-chan ChecksumCFG, uploadCh chan<- UploadCFG) {
	for cfg := range checksumCh {
		filename := compareChecksum(cfg.filename, cfg.checksumRemote, cfg.site)
		if len(filename) > 0 {
			// Add file to the upload queue
			uploadCh <- cfg.UploadCFG
		}
	}
}
