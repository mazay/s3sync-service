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
	"reflect"
	"testing"
)

func TestK8sGetCm(t *testing.T) {
	config := `
aws_region: us-east-1
sites:
- local_path: /some/local/path
bucket: mock-bucket
`

	k8sClient := K8sClient{}
	k8sClient.mockClient(config)

	tests := []struct {
		cm   string
		want string
	}{
		{"test/mock-configmap", "match"},
		{"test/mock-configmap-does-not-exist", "fail"},
	}
	for _, tt := range tests {
		data := k8sClient.k8sGetCm(tt.cm)
		if !reflect.DeepEqual(config, data) && tt.want != "fail" {
			t.Error(
				"Expected:", config,
				"got:", data,
			)
		}
	}
}
