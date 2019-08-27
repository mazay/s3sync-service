package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

// Site is a option for backing up data to S3
type Site struct {
	LocalPath    string `yaml:"local_path"`
	Bucket       string `yaml:"bucket"`
	BucketPath   string `yaml:"bucket_path"`
	BucketRegion string `yaml:"bucket_region"`
	StorageClass string `yaml:"storage_class"`
}

// Config structure - contains lst of Site options
type Config struct {
	Config []Site `yaml:",flow"`
}

func configProcessError(err error) {
	fmt.Println(err)
	os.Exit(2)
}

func readFile(config *Config, configpath string) {
	f, err := os.Open(configpath)
	if err != nil {
		configProcessError(err)
	}

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&config)
	if err != nil {
		configProcessError(err)
	}
}
