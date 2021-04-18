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

func TestValidLoggerInit(t *testing.T) {
	var config Config
	var LogLevels = []string{
		"",
		"trace",
		"debug",
		"info",
		"warn",
		"error",
		"fatal",
		"panic",
	}

	for _, item := range LogLevels {
		config.LogLevel = item
		initLogger(&config)
	}
}

func TestInvalidLoggerInit(t *testing.T) {
	var config Config
	var oldOsExit = osExit
	var got int

	config.LogLevel = "invalid_loglevel"

	defer func() { osExit = oldOsExit }()

	myExit := func(code int) {
		got = code
	}

	osExit = myExit

	initLogger(&config)

	if exp := 3; got != exp {
		t.Errorf("Expected exit code: %d, got: %d", exp, got)
	}
}
