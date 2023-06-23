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

	"github.com/bxcodec/faker/v3"
)

type generateS3KeyTest struct {
	bucketPath string
	localPath  string
	filePath   string
	expected   string
}

type getObjectSizeTest struct {
	bucket   string
	expected int64
}

func (s *Site) SeBucket(bucket string) {
	s.Bucket = bucket
}

func FakeConfig() (Config, error) {
	a := Config{}
	err := faker.FakeData(&a)
	if err != nil {
		return a, err
	}
	return a, nil
}

func TestGetS3Session(t *testing.T) {
	for x := 0; x < 10; x++ {
		config, err := FakeConfig()
		if err == nil {
			for _, site := range config.Sites {
				site.getS3Session()
				responseType := reflect.TypeOf(site.client).String()
				if responseType != "*s3.Client" {
					t.Error("Expected type *s3.Client, got", responseType)
				}
			}
		} else {
			t.Error(err)
		}
	}
}

func TestGenerateS3Key(t *testing.T) {
	var generateS3KeyTestData = []generateS3KeyTest{
		{"", "./", "./test_data/test.file", "test_data/test.file"},
		{"foo", "./", "./test_data/test.file", "foo/test_data/test.file"},
	}

	for _, testSet := range generateS3KeyTestData {
		result := generateS3Key(testSet.bucketPath, testSet.localPath, testSet.filePath)
		if result != testSet.expected {
			t.Error(
				"bucketPath:", testSet.bucketPath,
				"root:", testSet.localPath,
				"path:", testSet.filePath,
				"expected", testSet.expected,
				"got", result,
			)
		}
	}
}

func TestGetObjectSize(t *testing.T) {
	var site Site
	var getObjectSizeTestData = []getObjectSizeTest{
		{"fake-s3-bucket", 0},
	}

	for _, testSet := range getObjectSizeTestData {
		site.SeBucket(testSet.bucket)
		site.BucketRegion = "us-east-1"
		site.getS3Session()
		size := site.getObjectSize("some-key")
		if size != testSet.expected {
			t.Error(
				"Expected", testSet.expected,
				"got", size,
			)
		}
	}
}
