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
	"fmt"
	"io/fs"
	"os"
	"time"

	"github.com/radovskyb/watcher"
)

func watch(site *Site, uploadCh chan<- UploadCFG,
	siteStopperChan <-chan bool) {
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
					excluded := IsExcluded(filepath, site.Exclusions, site.Inclusions)
					if excluded {
						logger.Debugf("skipping without errors: %+v", filepath)
					} else {
						if fmt.Sprint(event.Op) == "REMOVE" {
							if site.RetireDeleted {
								s3Key := generateS3Key(site.BucketPath, site.LocalPath, filepath)
								uploadCh <- UploadCFG{s3Key, site, "delete"}
							}
						} else if fmt.Sprint(event.Op) == "RENAME" || fmt.Sprint(event.Op) == "MOVE" {
							// Upload the new object with new name/path
							uploadCh <- UploadCFG{filepath, site, "upload"}
							// remove the old object
							oldFilepath := fmt.Sprint(event.OldPath)
							removedS3Key := generateS3Key(site.BucketPath, site.LocalPath, oldFilepath)
							uploadCh <- UploadCFG{removedS3Key, site, "delete"}
						} else {
							// A shorthand workaround for skipping symlinks from processing
							if event.FileInfo.Mode().Type() == fs.ModeSymlink {
								logger.Infof("%s is a symlink, skipping", event.Path)
							} else {
								fileWatcher(site, uploadCh, filepath)
							}
						}
					}
				}
			case err := <-w.Error:
				// Update errors metric
				errorsMetric.WithLabelValues(site.LocalPath, site.Bucket, site.BucketPath, site.Name, "watcher").Inc()
				logger.Errorln(err)
			case <-w.Closed:
				return
			case <-siteStopperChan:
				wg.Done()
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
	if err := w.Start(site.WatchInterval); err != nil {
		// Update errors metric
		errorsMetric.WithLabelValues(site.LocalPath, site.Bucket, site.BucketPath, site.Name, "watcher").Inc()
		logger.Errorln(err)
	}
}

// fileWatcher is watching for file mtime and adds the file into the upload queue if it's > 30 seconds old
// Workaround for - https://github.com/radovskyb/watcher/issues/66
func fileWatcher(site *Site, uploadCh chan<- UploadCFG, filepath string) {
	for {
		file, _ := os.Stat(filepath)
		mtime := file.ModTime()
		if time.Since(mtime).Seconds() >= 30 {
			logger.Debugf("there were no writes to the file for 30 seconds, adding to the upload queue: %s", filepath)
			uploadCh <- UploadCFG{filepath, site, "upload"}
			return
		}

		logger.Debugf("looks like the file %s is still being written into, waiting", filepath)
		time.Sleep(time.Second * 10)
	}
}
