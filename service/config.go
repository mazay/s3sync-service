//
// s3sync-service - Realtime S3 synchronisation tool
// Copyright (c) 2020  Yevgeniy Valeyev
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
//

package service

import (
	"os"
	"runtime"
	"strings"
	"time"

	yaml "gopkg.in/yaml.v3"
)

// Config structure - contains lst of Site options
type Config struct {
	AccessKey         string        `yaml:"access_key"`
	SecretAccessKey   string        `yaml:"secret_access_key"`
	AwsRegion         string        `yaml:"aws_region"`
	LogLevel          string        `yaml:"loglevel"`
	UploadQueueBuffer int           `yaml:"upload_queue_buffer"`
	UploadWorkers     int           `yaml:"upload_workers"`
	ChecksumWorkers   int           `yaml:"checksum_workers"`
	WatchInterval     time.Duration `yaml:"watch_interval"`
	S3OpsRetries      int           `yaml:"s3_ops_retries"`
	Sites             []Site        `yaml:",flow"`
}

// Site is a option for backing up data to S3
type Site struct {
	Name            string        `yaml:"name"`
	LocalPath       string        `yaml:"local_path"`
	Bucket          string        `yaml:"bucket"`
	Endpoint        string        `yaml:"endpoint"`
	BucketPath      string        `yaml:"bucket_path"`
	BucketRegion    string        `yaml:"bucket_region"`
	StorageClass    string        `yaml:"storage_class"`
	AccessKey       string        `yaml:"access_key"`
	SecretAccessKey string        `yaml:"secret_access_key"`
	RetireDeleted   bool          `yaml:"retire_deleted"`
	Exclusions      []string      `yaml:",flow"`
	WatchInterval   time.Duration `yaml:"watch_interval"`
	S3OpsRetries    int           `yaml:"s3_ops_retries"`
}

func (cfg *Config) setDefaults() {
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}

	if cfg.UploadWorkers == 0 {
		cfg.UploadWorkers = 10
	}

	if cfg.ChecksumWorkers == 0 {
		// If in k8s then run 2 workers
		// otherwise run 2 workers per core
		if inK8s {
			cfg.ChecksumWorkers = 2
		} else {
			cfg.ChecksumWorkers = runtime.NumCPU() * 2
		}
	}

	if cfg.WatchInterval == 0 {
		cfg.WatchInterval = time.Millisecond * 1000
	}

	if cfg.S3OpsRetries == 0 {
		cfg.S3OpsRetries = 5
	}
}

func (site *Site) setDefaults(cfg *Config) {
	// Remove leading slash from the BucketPath
	site.BucketPath = strings.TrimPrefix(site.BucketPath, "/")
	// Set site name
	if site.Name == "" {
		site.Name = site.Bucket + "/" + site.BucketPath
		site.Name = strings.TrimSuffix(site.Name, "/")
	}
	// Set site AccessKey
	if site.AccessKey == "" {
		site.AccessKey = cfg.AccessKey
	}
	// Set site SecretAccessKey
	if site.SecretAccessKey == "" {
		site.SecretAccessKey = cfg.SecretAccessKey
	}
	// Set site BucketRegion
	if site.BucketRegion == "" {
		site.BucketRegion = cfg.AwsRegion
	}
	// Set default value for StorageClass
	if site.StorageClass == "" {
		site.StorageClass = "STANDARD"
	}
	// Set site WatchInterval
	if site.WatchInterval == 0 {
		site.WatchInterval = cfg.WatchInterval
	}
	// Set site S3OpsRetries
	if site.S3OpsRetries == 0 {
		site.S3OpsRetries = cfg.S3OpsRetries
	}
}

func configProcessError(err error) {
	logger.Error(err)
	osExit(2)
}

func getConfig() (bool, *Config) {
	var emptyConfig *Config

	empty := true

	if inK8s && configmap != "" {
		cfg := k8sGetCm(clientset, configmap)
		config = readConfigString(cfg)
	} else {
		config = readConfigFile(configpath)
	}

	if config != emptyConfig {
		empty = false
	}

	return empty, config
}

func readConfigFile(configpath string) *Config {
	f, err := os.Open(configpath)
	if err != nil {
		configProcessError(err)
	}

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&config)
	f.Close()
	if err != nil {
		configProcessError(err)
	} else {
		config.setDefaults()
	}

	return config
}

func readConfigString(cfgData string) *Config {
	err := yaml.Unmarshal([]byte(cfgData), &config)
	if err != nil {
		configProcessError(err)
	} else {
		config.setDefaults()
	}

	return config
}
