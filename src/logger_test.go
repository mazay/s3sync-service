package main

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
