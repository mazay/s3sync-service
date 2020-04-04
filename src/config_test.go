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
