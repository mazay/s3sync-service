package main

import (
	"reflect"
	"testing"

	"github.com/bxcodec/faker"
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
				s3Service := getS3Service(site)
				responseType := reflect.TypeOf(s3Service).String()
				if responseType != "*s3.S3" {
					t.Error("Expected type *s3.S3, got", responseType)
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
		s3Service := getS3Service(site)
		size := getObjectSize(s3Service, site, "some-key")
		if size != testSet.expected {
			t.Error(
				"Expected", testSet.expected,
				"got", size,
			)
		}
	}
}
