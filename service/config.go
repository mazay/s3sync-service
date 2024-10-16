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

// DeppCopy the config data
func (cfg *Config) DeepCopy() *Config {
	copy := &Config{
		AccessKey:         cfg.AccessKey,
		SecretAccessKey:   cfg.SecretAccessKey,
		AwsRegion:         cfg.AwsRegion,
		LogLevel:          cfg.LogLevel,
		UploadQueueBuffer: cfg.UploadQueueBuffer,
		UploadWorkers:     cfg.UploadWorkers,
		ChecksumWorkers:   cfg.ChecksumWorkers,
		WatchInterval:     cfg.WatchInterval,
		S3OpsRetries:      cfg.S3OpsRetries,
	}

	for _, site := range cfg.Sites {
		siteCopy := &Site{
			Name:            site.Name,
			LocalPath:       site.LocalPath,
			Bucket:          site.Bucket,
			Endpoint:        site.Endpoint,
			BucketPath:      site.BucketPath,
			BucketRegion:    site.BucketRegion,
			StorageClass:    site.StorageClass,
			AccessKey:       site.AccessKey,
			SecretAccessKey: site.SecretAccessKey,
			RetireDeleted:   site.RetireDeleted,
			WatchInterval:   site.WatchInterval,
			S3OpsRetries:    site.S3OpsRetries,
			client:          site.client,
		}
		siteCopy.Exclusions = append(siteCopy.Exclusions, site.Exclusions...)
		siteCopy.Inclusions = append(siteCopy.Inclusions, site.Inclusions...)

		copy.Sites = append(copy.Sites, *siteCopy)
	}

	return copy
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
	// Set site S3OpsRetries
	if len(site.Inclusions) == 0 {
		site.Inclusions = []string{".*"}
	}
}

func configProcessError(err error) {
	logger.Error(err)
	osExit(2)
}

func getConfig() (bool, *Config) {
	var (
		emptyConfig *Config
		empty       = true
	)

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
	var cfg *Config

	f, err := os.Open(configpath)
	if err != nil {
		configProcessError(err)
	}

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	f.Close()
	if err != nil {
		configProcessError(err)
	} else {
		cfg.setDefaults()
	}

	return cfg
}

func readConfigString(cfgData string) *Config {
	var cfg *Config

	err := yaml.Unmarshal([]byte(cfgData), &cfg)
	if err != nil {
		configProcessError(err)
	} else {
		cfg.setDefaults()
	}

	return cfg
}
