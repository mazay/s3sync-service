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
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func prometheusExporter(metricsPort string, metricsPath string) {
	http.Handle(metricsPath, promhttp.Handler())
	http.ListenAndServe(":"+metricsPort, nil)
}

func httpServer(httpPort string, reloaderChan chan<- bool) {
	http.HandleFunc("/reload", func(res http.ResponseWriter, req *http.Request) {
		// Trigger config reload
		reloaderChan <- true
		// Response with confirmation
		res.WriteHeader(http.StatusOK)
		fmt.Fprint(res, `{"status":"SUCCESS","message":"check logs for details"}`)
	})
	http.ListenAndServe(":"+httpPort, nil)
}
