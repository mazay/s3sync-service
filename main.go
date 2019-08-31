package main

import (
	"flag"
	"log"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
)

var wg = sync.WaitGroup{}
var logger = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)

// UploadCFG describes structure of upload configuration for the uploadWorker
type UploadCFG struct {
	s3Service *s3.S3
	file      string
	timeout   time.Duration
	site      Site
}

func main() {
	var config Config
	var configpath string
	var timeout time.Duration

	// Read command line args
	flag.StringVar(&configpath, "c", "config.yml", "Path to the config.yml")
	flag.DurationVar(&timeout, "t", 0, "Upload timeout.")
	flag.Parse()

	// Read config file
	readFile(&config, configpath)

	// Init upload worker
	if config.UploadWorkers == 0 {
		config.UploadWorkers = 10
	}

	uploadCh := make(chan UploadCFG, 500)
	for x := 0; x < config.UploadWorkers; x++ {
		go uploadWorker(uploadCh)
	}

	// Start separate thread for each site
	wg.Add(len(config.Sites))
	for _, site := range config.Sites {
		go syncSite(timeout, site, uploadCh)
	}
	wg.Wait()
}

func uploadWorker(uploadCh <-chan UploadCFG) {
	for cfg := range uploadCh {
		uploadConfig := cfg
		uploadFile(uploadConfig.s3Service, uploadConfig.file, uploadConfig.timeout, uploadConfig.site)
	}
}
