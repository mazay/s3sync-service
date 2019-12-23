package main

import (
	"testing"
	"time"
)

var siteTestData = []Site{
	{
		"test_site",
		"./test_data",
		"mock-s3-bucket",
		"",
		"",
		"",
		"",
		"",
		true,
		[]string{},
		time.Duration(1000),
	},
	{
		"test_site",
		"./test_data",
		"mock-s3-bucket",
		"",
		"",
		"",
		"AKIAI44QH8DHBEXAMPLE",
		"je7MtGbClwBF/2Zp9Utk/h3yCo8nvbEXAMPLEKEY",
		true,
		[]string{},
		time.Duration(1000),
	},
}

func TestGetS3Session(t *testing.T) {
	for _, site := range siteTestData {
		getS3Service(site)
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
