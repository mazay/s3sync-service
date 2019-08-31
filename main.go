package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"time"
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
				fmt.Printf("skipping without errors: %+v \n", info.Name())
			} else {
				files = append(files, path)
			}
		}
		return nil
	})
	return files, err
}

func main() {
	var config Config
	var configpath string
	var timeout time.Duration

	flag.StringVar(&configpath, "c", "config.yml", "Path to the config.yml")
	flag.DurationVar(&timeout, "t", 0, "Upload timeout.")
	flag.Parse()

	readFile(&config, configpath)

	for _, site := range config.Config {
		files, err := FilePathWalkDir(site.LocalPath, site.Exclusions)
		if err != nil {
			log.Fatal(err)
		}

		for _, f := range files {
			syncFile(f, timeout, site)
		}
	}
}
