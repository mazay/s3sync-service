package main

import (
	"flag"
	"sync"

	"github.com/aws/aws-sdk-go/service/s3"
)

var wg = sync.WaitGroup{}

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
}

func main() {
	var config Config
	var configpath string

	// Read command line args
	flag.StringVar(&configpath, "c", "config.yml", "Path to the config.yml")
	flag.Parse()

	// Read config file
	readFile(&config, configpath)

	// init logger
	initLogger(config)

	// Init upload worker
	if config.UploadWorkers == 0 {
		config.UploadWorkers = 10
	}

	uploadCh := make(chan UploadCFG, config.UploadQueueBuffer)
	for x := 0; x < config.UploadWorkers; x++ {
		go uploadWorker(uploadCh)
	}

	// Init checksum checker workers
	checksumCh := make(chan ChecksumCFG)
	for x := 0; x < 20; x++ {
		go checksumWorker(checksumCh, uploadCh)
	}

	// Start separate thread for each site
	wg.Add(len(config.Sites))
	for _, site := range config.Sites {
		// Set default value for StorageClass
		if site.StorageClass == "" {
			site.StorageClass = "STANDARD"
		}
		go syncSite(site, uploadCh, checksumCh)
	}
	wg.Wait()
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
		filename := compareChecksum(cfg.filename, cfg.checksumRemote)
		if len(filename) > 0 {
			// Add file to the upload queue
			uploadCh <- cfg.UploadCFG
		}
	}
}
