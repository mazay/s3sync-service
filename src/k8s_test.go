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
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8smock "k8s.io/client-go/kubernetes/fake"
)

func TestK8sGetCm(t *testing.T) {
	config := `
aws_region: us-east-1
sites:
- local_path: /some/local/path
  bucket: mock-bucket
`

	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "mock-configmap"},
		Data:       map[string]string{"config.yml": config},
	}
	clientset := k8smock.NewSimpleClientset(cm)

	tests := []struct {
		cm   string
		want string
	}{
		{"test/mock-configmap", "match"},
		{"test/mock-configmap-does-not-exist", "fail"},
	}
	for _, tt := range tests {
		data := k8sGetCm(clientset, tt.cm)
		if !reflect.DeepEqual(config, data) && tt.want != "fail" {
			t.Error(
				"Expected:", config,
				"got:", data,
			)
		}
	}
}
