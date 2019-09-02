package main

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

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
				logger.Printf("skipping without errors: %+v \n", path)
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
	if checksumRemote == "" {
		return filename
	}

	contents, err := ioutil.ReadFile(filename)
	if err == nil {
		sum := md5.Sum(contents)
		sumString := fmt.Sprintf("%x", sum)
		// checksums don't match, mark for upload
		if sumString != checksumRemote {
			return filename
		}
		// Files matched
		return ""
	}
	return filename
}
