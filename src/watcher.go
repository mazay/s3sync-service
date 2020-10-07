package main

import (
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/radovskyb/watcher"
)

func watch(s3Service *s3.S3, site Site, uploadCh chan<- UploadCFG) {
	logger.Printf("starting the FS watcher for site %s", site.Name)

	w := watcher.New()
	w.FilterOps(watcher.Create, watcher.Write, watcher.Remove, watcher.Rename, watcher.Move)

	go func() {
		for {
			select {
			case event := <-w.Event:
				if !event.IsDir() {
					logger.Debugln(event)
					// Convert filepath to string
					filepath := fmt.Sprint(event.Path)
					excluded := checkIfExcluded(filepath, site.Exclusions)
					if excluded {
						logger.Debugf("skipping without errors: %+v", filepath)
					} else {
						if fmt.Sprint(event.Op) == "REMOVE" {
							if site.RetireDeleted {
								s3Key := generateS3Key(site.BucketPath, site.LocalPath, filepath)
								uploadCh <- UploadCFG{s3Service, s3Key, site, "delete"}
							}
						} else if fmt.Sprint(event.Op) == "RENAME" || fmt.Sprint(event.Op) == "MOVE" {
							// Upload the new object with new name/path
							uploadCh <- UploadCFG{s3Service, filepath, site, "upload"}
							// remove the old object
							oldFilepath := fmt.Sprint(event.OldPath)
							removedS3Key := generateS3Key(site.BucketPath, site.LocalPath, oldFilepath)
							uploadCh <- UploadCFG{s3Service, removedS3Key, site, "delete"}
						} else {
							fileWatcher(s3Service, site, uploadCh, event, filepath)
						}
					}
				}
			case err := <-w.Error:
				// Update errors metric
				errorsMetric.WithLabelValues(site.LocalPath, site.Bucket, site.BucketPath, site.Name, "watcher").Inc()
				logger.Errorln(err)
			case <-w.Closed:
				return
			}
		}
	}()

	if err := w.AddRecursive(site.LocalPath); err != nil {
		// Update errors metric
		errorsMetric.WithLabelValues(site.LocalPath, site.Bucket, site.BucketPath, site.Name, "watcher").Inc()
		logger.Errorln(err)
	}

	// Start the watching process - it'll check for changes every Xms.
	if err := w.Start(time.Millisecond * site.WatchInterval); err != nil {
		// Update errors metric
		errorsMetric.WithLabelValues(site.LocalPath, site.Bucket, site.BucketPath, site.Name, "watcher").Inc()
		logger.Errorln(err)
	}
}

// fileWatcher is watching for file mtime and adds the file into the upload queue if it's > 30 seconds old
// Workaround for - https://github.com/radovskyb/watcher/issues/66
func fileWatcher(s3Service *s3.S3, site Site, uploadCh chan<- UploadCFG, event watcher.Event, filepath string) {
	for {
		select {
		case <-time.After(time.Second):
			file, _ := os.Stat(filepath)
			mtime := file.ModTime()
			if time.Now().Sub(mtime).Seconds() >= 30 {
				logger.Debugf("there were no writes to the file for 30 seconds, adding to the upload queue: %s", filepath)
				uploadCh <- UploadCFG{s3Service, filepath, site, "upload"}
				return
			}
		}
	}
}
