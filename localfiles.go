package main

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/aws/aws-sdk-go/service/s3"
)

func checkIfInList(item string, list []string) bool {
	for _, i := range list {
		if i == item {
			return true
		}
	}

	return false
}

func checkIfExcluded(path string, exclusions []string) bool {
	var excluded bool
	excluded = false

	for _, exclusion := range exclusions {
		re := regexp.MustCompile(exclusion)
		if re.FindAll([]byte(path), -1) != nil {
			excluded = true
		}
	}

	return excluded
}

// FilePathWalkDir walks throught the directory and all subdirectories returning list of files for upload and list of files to be deleted from S3
func FilePathWalkDir(site Site, awsItems map[string]string, s3Service *s3.S3, checksumCh chan<- ChecksumCFG) ([]string, error) {
	var deleteKeys []string
	var localS3Keys []string

	err := filepath.Walk(site.LocalPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			excluded := checkIfExcluded(path, site.Exclusions)
			if excluded {
				logger.Debug("skipping without errors: %+v", path)
			} else {
				s3Key := generateS3Key(site.BucketPath, site.LocalPath, path)
				localS3Keys = append(localS3Keys, s3Key)
				checksumRemote, _ := awsItems[s3Key]
				checksumCh <- ChecksumCFG{UploadCFG{s3Service, path, site}, path, checksumRemote}
			}
		}
		return nil
	})

	// Generate a list of deleted files
	for key := range awsItems {
		if !checkIfInList(key, localS3Keys) {
			deleteKeys = append(deleteKeys, key)
		}
	}

	return deleteKeys, err
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
