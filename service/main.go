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

package service

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"k8s.io/client-go/kubernetes"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	version    string
	status     string
	configpath string
	configmap  string
	config     *Config
	clientset  kubernetes.Interface

	startupTime = time.Now()

	osExit = os.Exit
	wg     = sync.WaitGroup{}

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

	errorsMetric = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
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
	file   string
	site   *Site
	action string
}

// ChecksumCFG - structure for the checksum comparison queue
type ChecksumCFG struct {
	UploadCFG      UploadCFG
	filename       string
	checksumRemote string
	site           *Site
}

func isInK8s() bool {
	_, present := os.LookupEnv("KUBERNETES_SERVICE_HOST")
	return present
}

// Start function starts the s3sync-service daemon
func Start() {
	var (
		httpPort    int
		metricsPort int
		metricsPath string
	)

	// Read command line args
	flag.StringVar(&configpath, "config", "config.yml", "Path to the config.yml")
	flag.StringVar(&configmap, "configmap", "", "K8s configmap in the format namespace/configmap, if set config is ignored and s3sync-service will read and watch for changes in the specified configmap")
	flag.IntVar(&httpPort, "http-port", 8090, "Port for internal HTTP server, 0 to disable")
	flag.IntVar(&metricsPort, "metrics-port", 9350, "Prometheus exporter port, 0 to disable the exporter")
	flag.StringVar(&metricsPath, "metrics-path", "/metrics", "Prometheus exporter path")
	flag.Parse()

	status = "STARTING"

	if configmap != "" {
		err := cmValidate(configmap)
		if err != nil {
			logger.Error(err)
			osExit(7)
		}
	}

	if isInK8s() && configmap != "" {
		clientset = k8sClientset()
	}

	// Read the config
	_empty, config := getConfig()
	if _empty {
		logger.Error("something went wrong - got empty configuration structure")
		osExit(5)
	}

	// init logger
	initLogger(config)

	// Init channels
	mainStopperChan := make(chan os.Signal, 1)
	siteStopperChan := make(chan bool)
	checksumStopperChan := make(chan bool)
	uploadStopperChan := make(chan bool)
	reloaderChan := make(chan bool)
	uploadCh := make(chan UploadCFG, config.UploadQueueBuffer)
	checksumCh := make(chan ChecksumCFG)

	// Start http server
	if httpPort != 0 {
		go httpServer(httpPort, reloaderChan)
	}

	// Start prometheus exporter
	if metricsPort != 0 {
		go prometheusExporter(metricsPort, metricsPath)
	}

	signal.Notify(mainStopperChan, syscall.SIGINT, syscall.SIGTERM)

	// Start main worker
	wg.Add(1)
	go func() {
		logger.Infoln("starting up")
		if inK8s && configmap != "" {
			go k8sWatchCm(clientset, configmap, reloaderChan)
		}
		// Start upload workers
		uploadWorker(config, uploadCh, uploadStopperChan)
		// Start checksum checker workers
		checksumWorker(config, checksumCh, uploadCh, checksumStopperChan)
		// Start syncing sites
		syncSites(config, uploadCh, checksumCh, siteStopperChan)
		status = "RUNNING"
		logger.Infoln("all systems go")
		// Wait for events
		for {
			select {
			case <-mainStopperChan:
				logger.Infoln("shutting down gracefully")
				status = "STOPPING"
				stopWorkers(config, siteStopperChan, uploadStopperChan, checksumStopperChan)
				wg.Done()
				return
			case force := <-reloaderChan:
				oldConfig := config.DeepCopy()
				_empty, config := getConfig()
				if !_empty && reflect.DeepEqual(config, oldConfig) && !force {
					logger.Infoln("no config changes detected, reload cancelled")
				} else {
					status = "RELOADING"
					// Reset metrics
					for _, site := range config.Sites {
						site.setDefaults(config)
						sizeMetric.WithLabelValues(site.LocalPath, site.Bucket, site.BucketPath, site.Name).Set(0)
						objectsMetric.WithLabelValues(site.LocalPath, site.Bucket, site.BucketPath, site.Name).Set(0)
						errorsMetric.WithLabelValues(site.LocalPath, site.Bucket, site.BucketPath, site.Name, "local").Set(0)
						errorsMetric.WithLabelValues(site.LocalPath, site.Bucket, site.BucketPath, site.Name, "cloud").Set(0)
						errorsMetric.WithLabelValues(site.LocalPath, site.Bucket, site.BucketPath, site.Name, "watcher").Set(0)
					}
					logger.Infoln("reloading configuration")
					if !_empty {
						// Switch logging level (if needed), can't be switched to lower verbosity
						setLogLevel(config.LogLevel)
						stopWorkers(oldConfig, siteStopperChan, uploadStopperChan, checksumStopperChan)
						logger.Debugln("reading config file")
						// Start upload workers
						uploadWorker(config, uploadCh, uploadStopperChan)
						// Start checksum checker workers
						checksumWorker(config, checksumCh, uploadCh, checksumStopperChan)
						// Start syncing sites
						syncSites(config, uploadCh, checksumCh, siteStopperChan)
						status = "RUNNING"
					}
				}
			}
		}
	}()

	wg.Wait()
}

func stopWorkers(config *Config, siteStopperChan chan<- bool,
	uploadStopperChan chan<- bool, checksumStopperChan chan<- bool,
) {
	logger.Debugln("sending stop signal to all site watchers")
	for range config.Sites {
		siteStopperChan <- true
	}
	logger.Debugln("sending stop signal to all upload workers")
	for x := 0; x < config.UploadWorkers; x++ {
		uploadStopperChan <- true
	}
	logger.Debugln("sending stop signal to all checksum workers")
	for x := 0; x < config.ChecksumWorkers; x++ {
		checksumStopperChan <- true
	}
}

func uploadWorker(config *Config, uploadCh <-chan UploadCFG,
	uploadStopperChan <-chan bool,
) {
	logger.Infof("starting %s upload workers", strconv.Itoa(config.UploadWorkers))
	for x := 0; x < config.UploadWorkers; x++ {
		wg.Add(1)
		go func() {
			for {
				select {
				case cfg := <-uploadCh:
					switch cfg.action {
					case "upload":
						cfg.site.uploadFile(cfg.file)
					case "delete":
						cfg.site.deleteFile(cfg.file)
					default:
						logger.Errorf("programming error, unknown action: %s", cfg.action)
					}
				case <-uploadStopperChan:
					wg.Done()
					return
				}
			}
		}()
	}
}

func checksumWorker(config *Config, checksumCh <-chan ChecksumCFG,
	uploadCh chan<- UploadCFG, checksumStopperChan <-chan bool,
) {
	logger.Infof("starting %s checksum workers", strconv.Itoa(config.ChecksumWorkers))
	for x := 0; x < config.ChecksumWorkers; x++ {
		wg.Add(1)
		go func() {
			for {
				select {
				case cfg := <-checksumCh:
					filename := CompareChecksum(cfg.filename, cfg.checksumRemote, cfg.site)
					if len(filename) > 0 {
						// Add file to the upload queue
						uploadCh <- cfg.UploadCFG
					}
				case <-checksumStopperChan:
					wg.Done()
					return
				}
			}
		}()
	}
}

func syncSites(config *Config, uploadCh chan<- UploadCFG,
	checksumCh chan<- ChecksumCFG, siteStopperChan <-chan bool,
) {
	// Start separate goroutine for each site
	for _, site := range config.Sites {
		go func(site Site) {
			site.setDefaults(config)
			wg.Add(1)
			// Initi S3 session
			site.getS3Session()
			// Watch directory for realtime sync
			go watch(&site, uploadCh, siteStopperChan)
			// Fetch S3 objects
			awsItems, err := site.getAwsS3ItemMap()
			if err != nil {
				logger.Errorln(err)
				wg.Done()
				osExit(4)
			} else {
				// Compare S3 objects with local
				FilePathWalkDir(&site, awsItems, uploadCh, checksumCh)
				logger.Infof("finished initial sync for site %s", site.Name)
			}
		}(site)
	}
}

// cmValidate checks if provided configmap argument matches the pattern
func cmValidate(cm string) error {
	cmSplit := strings.Split(cm, "/")
	if len(cmSplit) != 2 {
		return fmt.Errorf("unexpected format for the configmap argument, expected format is '<namespace>/<configmap-name>', got '%s'", cm)
	}

	return nil
}
