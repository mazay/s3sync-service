package main

import (
	"fmt"
	"os"
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
					// Convert filepath to string
					filepath := fmt.Sprint(event.Path)
					if fmt.Sprint(event.Op) == "REMOVE" {
						s3Key := generateS3Key(site.BucketPath, site.LocalPath, filepath)
						deleteFile(s3Service, site.Bucket, s3Key)
					} else if fmt.Sprint(event.Op) == "RENAME" || fmt.Sprint(event.Op) == "MOVE" {
						// Upload the new object with new name/path
						uploadCh <- UploadCFG{s3Service, filepath, timeout, site}
						// remove the old object
						oldFilepath := fmt.Sprint(event.OldPath)
						removedS3Key := generateS3Key(site.BucketPath, site.LocalPath, oldFilepath)
						deleteFile(s3Service, site.Bucket, removedS3Key)
					} else {
						excluded := checkIfExcluded(filepath, site.Exclusions)
						if excluded {
							logger.Printf("skipping without errors: %+v \n", filepath)
						} else {
							fileWatcher(s3Service, timeout, site, uploadCh, event, filepath)
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

// fileWatcher is watching for file mtime and adds the file into the upload queue if it's > 30 seconds old
// Workaround for - https://github.com/radovskyb/watcher/issues/66
func fileWatcher(s3Service *s3.S3, timeout time.Duration, site Site, uploadCh chan<- UploadCFG, event watcher.Event, filepath string) {
	for {
		select {
		case <-time.After(time.Second):
			file, _ := os.Stat(filepath)
			mtime := file.ModTime()
			if time.Now().Sub(mtime).Seconds() >= 30 {
				logger.Printf("there were no writes to the file for 30 seconds, adding to the upload queue: %s", filepath)
				uploadCh <- UploadCFG{s3Service, filepath, timeout, site}
				return
			}
		}
	}
}
