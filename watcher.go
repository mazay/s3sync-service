package main

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/radovskyb/watcher"
)

func watch(s3Service *s3.S3, timeout time.Duration, site Site, uploadCh chan<- UploadCFG) {
	w := watcher.New()
	w.FilterOps(watcher.Create, watcher.Write, watcher.Remove, watcher.Rename, watcher.Move)

	go func() {
		for {
			select {
			case event := <-w.Event:
				if !event.IsDir() {
					logger.Println(event)
					filepath := fmt.Sprint(event.Path)
					if fmt.Sprint(event.Op) == "REMOVE" {
						s3Key := generateS3Key(site.BucketPath, site.LocalPath, filepath)
						deleteFile(s3Service, site.Bucket, s3Key)
					} else {
						excluded := checkIfExcluded(filepath, site.Exclusions)
						if excluded {
							logger.Printf("skipping without errors: %+v \n", filepath)
						} else {
							uploadCh <- UploadCFG{s3Service, filepath, timeout, site}
						}
					}
				}
			case err := <-w.Error:
				logger.Fatal(err)
			case <-w.Closed:
				return
			}
		}
	}()

	if err := w.AddRecursive(site.LocalPath); err != nil {
		logger.Fatal(err)
	}

	// Start the watching process - it'll check for changes every 100ms.
	if err := w.Start(time.Millisecond * 100); err != nil {
		logger.Fatal(err)
	}
}
