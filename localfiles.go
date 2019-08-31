package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

// FilePathWalkDir walks throught the directory and all subdirectories returning list of files
func FilePathWalkDir(root string, skipList []string) ([]string, error) {
	var files []string
	var excluded = false

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			for _, exclusion := range skipList {
				re := regexp.MustCompile(exclusion)
				if re.FindAll([]byte(path), -1) != nil {
					excluded = true
				} else {
					excluded = false
				}
			}
			if excluded {
				fmt.Printf("skipping without errors: %+v \n", path)
			} else {
				files = append(files, path)
			}
		}
		return nil
	})
	return files, err
}
