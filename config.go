package main

import (
	"os"

	"gopkg.in/yaml.v2"
)

// Site is a option for backing up data to S3
type Site struct {
	LocalPath       string   `yaml:"local_path"`
	Bucket          string   `yaml:"bucket"`
	BucketPath      string   `yaml:"bucket_path"`
	BucketRegion    string   `yaml:"bucket_region"`
	StorageClass    string   `yaml:"storage_class"`
	AccessKey       string   `yaml:"access_key"`
	SecretAccessKey string   `yaml:"secret_access_key"`
	RetireDeleted   bool     `yaml:"retire_deleted"`
	Exclusions      []string `yaml:",flow"`
}

// Config structure - contains lst of Site options
type Config struct {
	UploadWorkers int    `yaml:"upload_workers"`
	Sites         []Site `yaml:",flow"`
}

func configProcessError(err error) {
	logger.Println(err)
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
