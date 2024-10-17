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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bxcodec/faker/v3"
)

func TestInfoHandler(t *testing.T) {
	// generate fake config
	err := faker.FakeData(&config)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("GET", "/info", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(infoHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Error(
			"handler returned wrong status code: got", status,
			"epected", http.StatusOK)
	}

	expected, err := json.Marshal(InfoResponse{
		version,
		startupTime,
		status,
		len(config.Sites),
		config.UploadWorkers,
		config.ChecksumWorkers,
		config.LogLevel,
	})
	if err != nil {
		t.Error(err.Error())
	}

	if rr.Body.String() != string(expected) {
		t.Error(
			"handler returned unexpected body: got", rr.Body.String(),
			"expected", string(expected),
		)
	}
}

func TestReloadHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/reload", nil)
	if err != nil {
		t.Fatal(err)
	}

	dummyChan := make(chan bool, 1)
	defer close(dummyChan)

	reloadHandler := ReloadHandler{Chan: dummyChan}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(reloadHandler.handler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Error(
			"handler returned wrong status code: got", status,
			"epected", http.StatusOK)
	}

	expected, err := json.Marshal(ReloadResponse{
		version,
		startupTime,
		status,
	})
	if err != nil {
		t.Error(err.Error())
	}

	if rr.Body.String() != string(expected) {
		t.Error(
			"handler returned unexpected body: got", rr.Body.String(),
			"expected", string(expected),
		)
	}
}

func TestPrometheusExporter(t *testing.T) {
	go prometheusExporter(9350, "/metrics")

	// Give http server some time to start
	time.Sleep(time.Second * 5)

	resp, err := http.Get("http://127.0.0.1:9350/metrics")
	if err != nil {
		t.Error(err.Error())
	}

	if resp.StatusCode != 200 {
		t.Error(
			"server returned unexpected status: got", resp.Status,
			"expected 200",
		)
	}
}

func TestHttpServerr(t *testing.T) {
	tstReloaderChan := make(chan bool)

	go httpServer(8090, tstReloaderChan)

	// Give http server some time to start
	time.Sleep(time.Second * 5)

	resp, err := http.Get("http://127.0.0.1:8090/info")
	if err != nil {
		t.Error(err.Error())
	}

	if resp.StatusCode != 200 {
		t.Error(
			"server returned unexpected status: got", resp.Status,
			"expected 200",
		)
	}
}
