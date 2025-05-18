package service

import "os"

func secureFileClose(file *os.File) {
	if err := file.Close(); err != nil {
		logger.Error(err)
	}
}

func secureRemove(path string) {
	if err := os.Remove(path); err != nil {
		logger.Error(err)
	}
}

func secureRemoveAll(path string) {
	if err := os.RemoveAll(path); err != nil {
		logger.Error(err)
	}
}
