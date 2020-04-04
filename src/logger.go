package main

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

// LoggerInitError is a custom error representation
func LoggerInitError(err error) {
	logger.Error(err)
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
		logger.SetFormatter(&logrus.JSONFormatter{})
		logger.SetOutput(os.Stdout)

		logger.SetLevel(LogLevel)

		if config.LogLevel == "trace" || config.LogLevel == "debug" {
			logger.SetReportCaller(true)
		}
	} else {
		LoggerInitError(fmt.Errorf("Log level definition not found for '%s'", config.LogLevel))
	}
}
