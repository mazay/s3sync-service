package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"time"
)

// FilePathWalkDir walks throught the directory and all subdirectories returning list of files
func FilePathWalkDir(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
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
		// Set default value for StorageClass, available values are here
		// https://docs.aws.amazon.com/AmazonS3/latest/dev/storage-class-intro.html#sc-compare
		if site.StorageClass == "" {
			site.StorageClass = "STANDARD"
		}

		files, err := FilePathWalkDir(site.LocalPath)
		if err != nil {
			log.Fatal(err)
		}

		for _, f := range files {
			syncFile(f, timeout, site)
		}
	}
}
