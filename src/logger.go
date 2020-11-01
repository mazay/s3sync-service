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
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

var (
	log    = logrus.New()
	logger = log.WithFields(logrus.Fields{"app": "s3sync-service", "version": version})
)

// LoggerInitError is a custom error representation
func LoggerInitError(err error) {
	log.Error(err)
	osExit(3)
}

func setLogLevel(level string) {
	logLevels := map[string]logrus.Level{
		"trace": logrus.TraceLevel,
		"debug": logrus.DebugLevel,
		"info":  logrus.InfoLevel,
		"warn":  logrus.WarnLevel,
		"error": logrus.ErrorLevel,
		"fatal": logrus.FatalLevel,
		"panic": logrus.PanicLevel,
	}

	if level == "" {
		level = "info"
	}

	if LogLevel, ok := logLevels[level]; ok {
		log.SetLevel(LogLevel)

		if level == "trace" || level == "debug" {
			log.SetReportCaller(true)
		}
	} else {
		LoggerInitError(fmt.Errorf("Log level definition not found for '%s'", level))
	}
}

func initLogger(config *Config) {
	log.SetFormatter(&logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime: "@timestamp",
			logrus.FieldKeyMsg:  "message",
		},
	})

	log.SetOutput(os.Stdout)

	setLogLevel(config.LogLevel)
}
