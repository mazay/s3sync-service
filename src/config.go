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

package main

import (
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

// Site is a option for backing up data to S3
type Site struct {
	Name            string        `yaml:"name"`
	LocalPath       string        `yaml:"local_path"`
	Bucket          string        `yaml:"bucket"`
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

func configProcessError(err error) {
	logger.Error(err)
	osExit(2)
}

func readConfigFile(configpath string) *Config {
	var config *Config

	f, err := os.Open(configpath)
	if err != nil {
		configProcessError(err)
	}

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&config)
	if err != nil {
		configProcessError(err)
	}

	return config
}
