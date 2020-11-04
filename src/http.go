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
	"encoding/json"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// InfoResponse defines info response data
type InfoResponse struct {
	VERSION         string
	STARTUPTIME     time.Time
	STATUS          string
	SITES           int
	UPLOADWORKERS   int
	CHECKSUMWORKERS int
	LOGLEVEL        string
}

// ReloadResponse defines info response data
type ReloadResponse struct {
	VERSION     string
	STARTUPTIME time.Time
	STATUS      string
}

// ReloadHandler defines reload handler
type ReloadHandler struct {
	Chan chan<- bool
}

// Reload handler method
func (rh *ReloadHandler) handler(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Server", "s3sync-service")
	res.Header().Set("Content-Type", "application/json")

	info := ReloadResponse{
		version,
		startupTime,
		status,
	}

	js, err := json.Marshal(info)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	rh.Chan <- true

	res.WriteHeader(http.StatusOK)
	res.Write(js)
}

// Prometheus exporter http server
func prometheusExporter(metricsPort string, metricsPath string) {
	http.Handle(metricsPath, promhttp.Handler())
	http.ListenAndServe(":"+metricsPort, nil)
}

// API http server
func httpServer(httpPort string, reloaderChan chan<- bool) {
	reloadHandler := ReloadHandler{Chan: reloaderChan}
	http.HandleFunc("/info", info)
	http.HandleFunc("/reload", reloadHandler.handler)
	http.ListenAndServe(":"+httpPort, nil)
}

// Info resource
func info(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Server", "s3sync-service")
	res.Header().Set("Content-Type", "application/json")

	info := InfoResponse{
		version,
		startupTime,
		status,
		len(config.Sites),
		config.UploadWorkers,
		config.ChecksumWorkers,
		config.LogLevel,
	}

	js, err := json.Marshal(info)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusOK)
	res.Write(js)
}
