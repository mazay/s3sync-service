package main

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/service/s3"
)

func checkIfExcluded(path string, exclusions []string) bool {
	excluded := false

	for _, exclusion := range exclusions {
		re := regexp.MustCompile(exclusion)
		if re.FindAll([]byte(path), -1) != nil {
			excluded = true
		}
	}

	return excluded
}

// FilePathWalkDir walks throught the directory and all subdirectories returning list of files for upload and list of files to be deleted from S3
func FilePathWalkDir(site Site, awsItems map[string]string, s3Service *s3.S3, uploadCh chan<- UploadCFG, checksumCh chan<- ChecksumCFG) {
	err := filepath.Walk(site.LocalPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logger.Error(err)
		}

		if !info.IsDir() {
			excluded := checkIfExcluded(path, site.Exclusions)
			s3Key := generateS3Key(site.BucketPath, site.LocalPath, path)
			if excluded {
				logger.Debugf("skipping without errors: %+v", path)
				// Delete the excluded object if already in the cloud
				if awsItems[s3Key] != "" {
					uploadCh <- UploadCFG{s3Service, s3Key, site, "delete"}
				}
			} else {
				checksumRemote, _ := awsItems[s3Key]
				checksumCh <- ChecksumCFG{UploadCFG{s3Service, path, site, "upload"}, path, checksumRemote}
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
				uploadCh <- UploadCFG{s3Service, key, site, "delete"}
			}
		}
	}

	if err != nil {
		logger.Error(err)
	}

	return
}

func compareChecksum(filename string, checksumRemote string) string {
	var sumOfSums []byte
	var parts int
	var finalSum []byte
	chunkSize := int64(5 * 1024 * 1024)

	logger.Debugf("%s: comparing checksums", filename)

	if checksumRemote == "" {
		return filename
	}

	file, err := os.Open(filename)
	if err != nil {
		logger.Error(err)
		return ""
	}
	defer file.Close()

	dataSize, err := file.Seek(0, io.SeekEnd)
	if err != nil {
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
		logger.Debugf("%s: checksums do not match, local checksum is %s, remote - %s", filename, sumHex, checksumRemote)
		return filename
	}
	logger.Debugf("%s: checksums matched", filename)
	return ""
}

func chunkMd5Sum(file io.ReadSeeker, start int64, length int64) ([]byte, error) {
	file.Seek(start, io.SeekStart)
	h := md5.New()
	if _, err := io.CopyN(h, file, length); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}
