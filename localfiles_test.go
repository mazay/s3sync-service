package main

import (
	"testing"
)

type exclusionsTest struct {
	path       string
	exclusions []string
	result     bool
}

var exclusionsTestData = []exclusionsTest{
	{"/etc/init.d", []string{"etc", "init.d"}, true},
	{"/var/log/messages", []string{}, false},
	{"", []string{"etc"}, false},
}

func TestCheckIfExcluded(t *testing.T) {
	for _, testSet := range exclusionsTestData {
		result := checkIfExcluded(testSet.path, testSet.exclusions)
		if result != testSet.result {
			t.Error(
				"For path", testSet.path,
				"with exclusions", testSet.exclusions,
				"expected", testSet.result,
				"got", result,
			)
		}
	}
}

type checksumComparisonTest struct {
	filename string
	checksum string
	result   string
}

var checksumTestData = []checksumComparisonTest{
	{"test_data/test.file", "1bc6a9a8be0cc1d8e1f0b734c8911e6c", ""},
	{"test_data/test.file", "123", "test_data/test.file"},
	{"test_data/test.file", "", "test_data/test.file"},
	{"test_data/nonexistent.file", "123", ""},
	{"", "", ""},
}

func TestCompareChecksum(t *testing.T) {
	for _, testSet := range checksumTestData {
		result := compareChecksum(testSet.filename, testSet.checksum)
		if result != testSet.result {
			t.Error(
				"For file", testSet.filename,
				"with checksum", testSet.checksum,
				"expected", testSet.result,
				"got", result,
			)
		}
	}
}
