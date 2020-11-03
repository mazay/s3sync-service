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
	"context"
	"strings"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func k8sClientset() *kubernetes.Clientset {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		logger.Panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger.Panic(err.Error())
	}

	return clientset
}

func k8sWatchCm(configmap string, reloaderChan chan<- bool) {
	clientset := k8sClientset()
	ctx := context.Background()
	cm := strings.Split(configmap, "/")
	namespace := cm[0]
	configmapName := cm[1]

	watcher, err := clientset.CoreV1().ConfigMaps(namespace).Watch(ctx, metav1.ListOptions{
		FieldSelector: fields.Set{"metadata.name": configmapName}.AsSelector().String(),
		Watch:         true,
	})

	if err != nil {
		logger.Fatalln(err.Error())
	}

	for event := range watcher.ResultChan() {
		logger.Infof("configmap %s was %s", configmap, event.Type)
		reloaderChan <- true
	}
}

func k8sGetCm(configmap string) string {
	var configMap map[string]string

	clientset := k8sClientset()
	ctx := context.Background()
	cm := strings.Split(configmap, "/")
	namespace := cm[0]
	configmapName := cm[1]

	cmObj, err := clientset.CoreV1().ConfigMaps(namespace).Get(ctx, configmapName,
		metav1.GetOptions{})

	if err != nil {
		logger.Fatalln(err.Error())
	} else {
		configMap = cmObj.Data
	}

	return configMap["config.yml"]
}
