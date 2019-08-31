package main

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
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
func FilePathWalkDir(root string, exclusions []string, awsItems map[string]string, bucketPath string) ([]string, []string, error) {
	var uploadFiles []string
	var deleteKeys []string
	var localS3Keys []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			excluded := checkIfExcluded(path, exclusions)
			if excluded {
				logger.Printf("skipping without errors: %+v \n", path)
			} else {
				s3Key := generateS3Key(bucketPath, root, path)
				localS3Keys = append(localS3Keys, s3Key)
				checksumRemote, _ := awsItems[s3Key]
				filename := compareChecksum(path, checksumRemote)
				if len(filename) > 0 {
					uploadFiles = append(uploadFiles, path)
				}
			}
		}
		return nil
	})

	for key := range awsItems {
		if !checkIfInList(key, localS3Keys) {
			deleteKeys = append(deleteKeys, key)
		}
	}

	return uploadFiles, deleteKeys, err
}

func compareChecksum(filename string, checksumRemote string) string {
	if checksumRemote == "" {
		return filename
	}

	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		return ""
	}
	sum := md5.Sum(contents)
	sumString := fmt.Sprintf("%x", sum)
	if sumString != checksumRemote {
		return filename
	}
	return ""
}
