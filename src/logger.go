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

func initLogger(config *Config) {
	logLevels := map[string]logrus.Level{
		"trace": logrus.TraceLevel,
		"debug": logrus.DebugLevel,
		"info":  logrus.InfoLevel,
		"warn":  logrus.WarnLevel,
		"error": logrus.ErrorLevel,
		"fatal": logrus.FatalLevel,
		"panic": logrus.PanicLevel,
	}

	// set default loglevel
	if config.LogLevel == "" {
		config.LogLevel = "info"
	}

	if LogLevel, ok := logLevels[config.LogLevel]; ok {
		log.SetFormatter(&logrus.JSONFormatter{
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime: "@timestamp",
				logrus.FieldKeyMsg:  "message",
			},
		})

		log.SetOutput(os.Stdout)

		log.SetLevel(LogLevel)

		if config.LogLevel == "trace" || config.LogLevel == "debug" {
			log.SetReportCaller(true)
		}
	} else {
		LoggerInitError(fmt.Errorf("Log level definition not found for '%s'", config.LogLevel))
	}
}
