package main

import (
	"reflect"
	"testing"
)

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
		}
	}
}

type generateS3KeyTest struct {
	bucketPath string
	localPath  string
	filePath   string
	result     string
}

var generateS3KeyTestData = []generateS3KeyTest{
	{"", "./", "./test_data/test.file", "test_data/test.file"},
	{"foo", "./", "./test_data/test.file", "foo/test_data/test.file"},
}

func TestGenerateS3Key(t *testing.T) {
	for _, testSet := range generateS3KeyTestData {
		result := generateS3Key(testSet.bucketPath, testSet.localPath, testSet.filePath)
		if result != testSet.result {
			t.Error(
				"bucketPath:", testSet.bucketPath,
				"root:", testSet.localPath,
				"path:", testSet.filePath,
				"expected", testSet.result,
				"got", result,
			)
		}
	}
}
