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
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWatch(t *testing.T) {
	// get os specific tmp path
	tmpPath, exists := os.LookupEnv("TMPDIR")
	if !exists {
		tmpPath = os.TempDir()
	}

	// Create a test directory
	tmpPath = filepath.Join(tmpPath, "s3sync-test")
	err := os.MkdirAll(tmpPath, 0o755)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpPath)

	// Create a mock site
	site := &Site{
		Name:          "test",
		LocalPath:     tmpPath,
		WatchInterval: time.Millisecond * 100,
		RetireDeleted: true,
	}

	// Create a mock upload channel
	uploadCh := make(chan UploadCFG)

	// Create a mock site stopper channel
	siteStopperChan := make(chan bool)

	// Start the watcher
	wg.Add(1)
	go watch(site, uploadCh, siteStopperChan)

	// Wait for the watcher to start
	time.Sleep(time.Second * 1)

	// Create a test file
	testFile, err := os.CreateTemp(site.LocalPath, "testfile")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(testFile.Name())

	// Wait for the watcher to detect the file
	time.Sleep(time.Second * 35)

	// Check that the file was scheduled for upload
	select {
	case upload := <-uploadCh:
		if upload.file != testFile.Name() {
			t.Errorf("Expected upload of %s, got %s", testFile.Name(), upload.file)
		}
	default:
		t.Error("Expected upload, got none")
	}

	// Delete the test file
	err = os.Remove(testFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	// Wait for the watcher to detect the delete
	time.Sleep(time.Second * 5)

	// Check that the file was scheduled for delete
	select {
	case upload := <-uploadCh:
		if upload.action != "delete" {
			t.Errorf("Expected delete, got %s", upload.action)
		}
	default:
		t.Error("Expected delete, got none")
	}

	// Stop the watcher
	siteStopperChan <- true
}
