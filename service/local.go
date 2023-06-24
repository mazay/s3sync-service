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
	"crypto/md5"
	"encoding/hex"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// IsExcluded check if a path is excluded
func IsExcluded(path string, exclusions []string, inclusions []string) bool {
	var excluded bool

	for _, inclusion := range inclusions {
		re := regexp.MustCompile(inclusion)
		if re.FindAll([]byte(path), -1) != nil {
			excluded = false
		}
	}

	for _, exclusion := range exclusions {
		re := regexp.MustCompile(exclusion)
		if re.FindAll([]byte(path), -1) != nil {
			excluded = true
		}
	}

	return excluded
}

// FilePathWalkDir walks through the directory and all subdirectories returning list of files for upload and list of files to be deleted from S3
func FilePathWalkDir(site *Site, awsItems map[string]string, uploadCh chan<- UploadCFG, checksumCh chan<- ChecksumCFG) {
	err := filepath.WalkDir(site.LocalPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Update errors metric
			errorsMetric.WithLabelValues(site.LocalPath, site.Bucket, site.BucketPath, site.Name, "local").Inc()
			logger.Error(err)
			return err
		}

		if !d.IsDir() {
			// A shorthand workaround for skipping symlinks from processing
			if d.Type() == fs.ModeSymlink {
				logger.Infof("%s is a symlink, skipping", path)
			} else {
				excluded := IsExcluded(path, site.Exclusions, site.Inclusions)
				s3Key := generateS3Key(site.BucketPath, site.LocalPath, path)
				if excluded {
					logger.Debugf("skipping without errors: %+v", path)
					// Delete the excluded object if already in the cloud
					if awsItems[s3Key] != "" && site.RetireDeleted {
						uploadCh <- UploadCFG{s3Key, site, "delete"}
					}
				} else {
					checksumCh <- ChecksumCFG{UploadCFG{path, site, "upload"}, path, awsItems[s3Key], site}
				}
			}
		}
		return nil
	})

	// Check for deleted files
	if site.RetireDeleted {
		for key := range awsItems {
			// Generate localPath by removing BucketPath from the key value and addint LocalPath
			localPath := filepath.Join(site.LocalPath, strings.Replace(key, site.BucketPath, "", 1))
			// Send s3 key for deleteion if generated localPath does not exist
			if _, err := os.Stat(localPath); os.IsNotExist(err) {
				uploadCh <- UploadCFG{key, site, "delete"}
			}
		}
	}

	if err != nil {
		// Update errors metric
		errorsMetric.WithLabelValues(site.LocalPath, site.Bucket, site.BucketPath, site.Name, "local").Inc()
		logger.Error(err)
	}
}

// CompareChecksum compares local file checksum with S3 ETag value
func CompareChecksum(filename string, checksumRemote string, site *Site) string {
	var sumOfSums []byte
	var parts int
	var finalSum []byte
	chunkSize := int64(5 * 1024 * 1024)

	logger.Debugf("%s: comparing checksums", filename)

	if checksumRemote == "" {
		logger.Infof("%s: remote checksum is unavailable, probably dealing with a new file", filename)
		return filename
	}

	file, err := os.Open(filename)
	if err != nil {
		// Update errors metric
		errorsMetric.WithLabelValues(site.LocalPath, site.Bucket, site.BucketPath, site.Name, "local").Inc()
		logger.Error(err)
		return ""
	}
	defer file.Close()

	dataSize, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		// Update errors metric
		errorsMetric.WithLabelValues(site.LocalPath, site.Bucket, site.BucketPath, site.Name, "local").Inc()
		logger.Error(err)
		return ""
	}

	for start := int64(0); start < dataSize; start += chunkSize {
		length := chunkSize
		if start+chunkSize > dataSize {
			length = dataSize - start
		}
		sum, err := chunkMd5Sum(file, start, length)
		if err != nil {
			// Update errors metric
			errorsMetric.WithLabelValues(site.LocalPath, site.Bucket, site.BucketPath, site.Name, "local").Inc()
			logger.Error(err)
			return ""
		}
		sumOfSums = append(sumOfSums, sum...)
		parts++
	}

	if parts == 1 {
		finalSum = sumOfSums
	} else {
		h := md5.New()
		_, err := h.Write(sumOfSums)
		if err != nil {
			// Update errors metric
			errorsMetric.WithLabelValues(site.LocalPath, site.Bucket, site.BucketPath, site.Name, "local").Inc()
			logger.Error(err)
			return ""
		}
		finalSum = h.Sum(nil)
	}

	sumHex := hex.EncodeToString(finalSum)

	if parts > 1 {
		sumHex += "-" + strconv.Itoa(parts)
	}

	if sumHex != checksumRemote {
		logger.Infof("%s: checksums missmatch, local checksum is %s, remote - %s", filename, sumHex, checksumRemote)
		return filename
	}
	logger.Debugf("%s: checksums matched", filename)
	return ""
}

func chunkMd5Sum(file io.ReadSeeker, start int64, length int64) ([]byte, error) {
	if _, err := file.Seek(start, io.SeekStart); err != nil {
		return nil, err
	}

	h := md5.New()
	if _, err := io.CopyN(h, file, length); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}
