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
	"reflect"
	"testing"
)

type SetDefaultsTest struct {
	cfgData  string
	expected bool
}

func TestGetConfig(t *testing.T) {
	configpath = "../test_data/valid_config.yml"
	empty, _ := getConfig()

	if !empty {
		t.Error("Expected to get config data, got empty struct")
	}
}

func TestReadConfigFile(t *testing.T) {
	config := readConfigFile("../test_data/valid_config.yml")
	if config == nil {
		t.Error("Expected successful config read for ../test_data/valid_config.yml")
	}
}

func TestReadNonexistentConfigFile(t *testing.T) {
	oldOsExit := osExit
	defer func() { osExit = oldOsExit }()

	var got int
	myExit := func(code int) {
		got = code
	}

	osExit = myExit

	readConfigFile("../test_data/nonexistent_config.yml")

	if exp := 2; got != exp {
		t.Errorf("Expected exit code: %d, got: %d", exp, got)
	}
}

func TestSetDefaults(t *testing.T) {
	var SetDefaultsData = []SetDefaultsTest{
		// Valid set with all fields set - no overrides with defaults
		{
			`
aws_region: us-east-1
loglevel: error
upload_workers: 1
checksum_workers: 2
watch_interval: 60000ms
s3_ops_retries: 3
sites:
- name: test-s3sync-service
  local_path: /some/local/path
  bucket: test-s3sync-service
  bucket_path: some_path
  bucket_region: eu-central-1
  storage_class: STANDARD_IA
  access_key: AKIAI44QH8DHBEXAMPLE
  secret_access_key: je7MtGbClwBF/2Zp9Utk/h3yCo8nvbEXAMPLEKEY
  watch_interval: 5m
  s3_ops_retries: 5
`,
			true,
		},
		// Valid set with not all fields set - some fields are filled in by default values
		{
			`
aws_region: us-east-1
sites:
- local_path: /some/local/path
  bucket: test-s3sync-service
`,
			false,
		},
	}

	for _, testSet := range SetDefaultsData {
		config := readConfigString(testSet.cfgData)
		for _, site := range config.Sites {
			siteOrig := site
			site.setDefaults(config)
			result := reflect.DeepEqual(site, siteOrig)
			if result != testSet.expected {
				t.Error(
					"For cfgData", testSet.cfgData,
					"expected", testSet.expected,
					"got", result,
				)
			}
		}
	}
}
