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
	"testing"
)

type ReadConfigTest struct {
	filePath string
	expected string
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
