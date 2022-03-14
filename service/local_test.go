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
	"testing"
)

type exclusionsTest struct {
	path       string
	exclusions []string
	inclusions []string
	expected   bool
}

type checksumComparisonTest struct {
	filename string
	checksum string
	site     Site
	expected string
}

func TestIsExcluded(t *testing.T) {
	var exclusionsTestData = []exclusionsTest{
		{"/etc/init.d", []string{"etc", "init.d"}, []string{}, true},
		{"/some/file.txt", []string{".*"}, []string{"txt"}, false},
		{"/var/log/messages", []string{}, []string{}, false},
		{"", []string{"etc"}, []string{}, false},
	}

	for _, testSet := range exclusionsTestData {
		result := IsExcluded(testSet.path, testSet.exclusions, testSet.inclusions)
		if result != testSet.expected {
			t.Error(
				"For path", testSet.path,
				"with exclusions", testSet.exclusions,
				"with inclusions", testSet.inclusions,
				"expected", testSet.expected,
				"got", result,
			)
		}
	}
}

func TestCompareChecksum(t *testing.T) {
	var checksumTestData = []checksumComparisonTest{
		{"../test_data/test.file", "18deb71f004bfe392e9e7b84ea67a7a3", Site{}, ""},
		{"../test_data/test.file", "123", Site{}, "../test_data/test.file"},
		{"../test_data/test.file", "", Site{}, "../test_data/test.file"},
		{"../test_data/nonexistent.file", "123", Site{}, ""},
		{"", "", Site{}, ""},
	}

	for _, testSet := range checksumTestData {
		result := CompareChecksum(testSet.filename, testSet.checksum, testSet.site)
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
