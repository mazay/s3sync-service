package main

import (
	"os"

	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

func initLogger(config Config) {
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

	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetOutput(os.Stdout)

	logger.SetLevel(logLevels[config.LogLevel])

	if config.LogLevel == "trace" || config.LogLevel == "debug" {
		logger.SetReportCaller(true)
	}
}
