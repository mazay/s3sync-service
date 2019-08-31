package main

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
)

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

// FilePathWalkDir walks throught the directory and all subdirectories returning list of files
func FilePathWalkDir(root string, exclusions []string, awsItems map[string]string, bucketPath string) ([]string, error) {
	var files []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			excluded := checkIfExcluded(path, exclusions)
			if excluded {
				fmt.Printf("skipping without errors: %+v \n", path)
			} else {
				relativeRoot, _ := filepath.Rel(bucketPath, root)
				relativePath, _ := filepath.Rel(relativeRoot, path)
				checksumRemote, _ := awsItems[relativePath]
				filename := compareChecksum(path, checksumRemote)
				if len(filename) > 0 {
					files = append(files, path)
				}
			}
		}
		return nil
	})

	return files, err
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
