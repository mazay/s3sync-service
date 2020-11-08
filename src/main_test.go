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
	"os"
	"testing"
)

func TestIsInK8s(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"inK8s", true},
		{"notInK8s", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "inK8s" {
				os.Setenv("KUBERNETES_SERVICE_HOST", "127.0.0.1")
			} else if tt.name == "notInK8s" {
				os.Unsetenv("KUBERNETES_SERVICE_HOST")
			}
			if got := isInK8s(); got != tt.want {
				t.Errorf("isInK8s() = %v, want %v", got, tt.want)
			}
		})
	}
}
