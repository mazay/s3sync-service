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
	"errors"
	"fmt"
	"os"
	"testing"
)

func TestIsInK8s(t *testing.T) {
	var (
		err   error
		tests = []struct {
			name string
			want bool
		}{
			{"inK8s", true},
			{"notInK8s", false},
		}
	)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.name {
			case "inK8s":
				err = os.Setenv("KUBERNETES_SERVICE_HOST", "127.0.0.1")
				if err != nil {
					t.Fatal(err)
				}
			case "notInK8s":
				err = os.Unsetenv("KUBERNETES_SERVICE_HOST")
				if err != nil {
					t.Fatal(err)
				}
			}
			if got := isInK8s(); got != tt.want {
				t.Errorf("isInK8s() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCmValidate(t *testing.T) {
	tests := []struct {
		cm     string
		result bool
	}{
		{"configmap-name", false},
		{"namespace/configmap-name", true},
	}
	for _, tt := range tests {
		t.Run(tt.cm, func(t *testing.T) {
			if err := cmValidate(tt.cm); errors.Is(err, nil) != tt.result {
				fmt.Print(errors.Is(err, nil))
				t.Errorf("cmValidate unexpectedly failed with %s", err)
			}
		})
	}
}
