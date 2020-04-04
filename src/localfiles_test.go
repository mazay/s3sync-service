package main

import (
	"testing"
)

type exclusionsTest struct {
	path       string
	exclusions []string
	expected   bool
}

type checksumComparisonTest struct {
	filename string
	checksum string
	site     Site
	expected string
}

func TestCheckIfExcluded(t *testing.T) {
	var exclusionsTestData = []exclusionsTest{
		{"/etc/init.d", []string{"etc", "init.d"}, true},
		{"/var/log/messages", []string{}, false},
		{"", []string{"etc"}, false},
	}

	for _, testSet := range exclusionsTestData {
		result := checkIfExcluded(testSet.path, testSet.exclusions)
		if result != testSet.expected {
			t.Error(
				"For path", testSet.path,
				"with exclusions", testSet.exclusions,
				"expected", testSet.expected,
				"got", result,
			)
		}
	}
}

func TestCompareChecksum(t *testing.T) {
	var checksumTestData = []checksumComparisonTest{
		{"../test_data/test.file", "1bc6a9a8be0cc1d8e1f0b734c8911e6c", Site{}, ""},
		{"../test_data/test.file", "123", Site{}, "../test_data/test.file"},
		{"../test_data/test.file", "", Site{}, "../test_data/test.file"},
		{"../test_data/nonexistent.file", "123", Site{}, ""},
		{"", "", Site{}, ""},
	}

	for _, testSet := range checksumTestData {
		result := compareChecksum(testSet.filename, testSet.checksum, testSet.site)
		if result != testSet.expected {
			t.Error(
				"For file", testSet.filename,
				"with checksum", testSet.checksum,
				"expected", testSet.expected,
				"got", result,
			)
		}
	}
}
