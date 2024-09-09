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
	"context"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

func k8sClientset() kubernetes.Interface {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		logger.Panic(err.Error())
		osExit(6)
	}

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		logger.Panic(err.Error())
		osExit(6)
	}

	return clientset
}

func k8sWatchCm(clientset kubernetes.Interface, configmap string, reloaderChan chan<- bool) {
	cm := strings.Split(configmap, "/")
	namespace := cm[0]
	configmapName := cm[1]

	logger.Infoln("starting to watch for configmap changes")

	watchlist := cache.NewListWatchFromClient(
		clientset.CoreV1().RESTClient(),
		"configmaps",
		namespace,
		fields.OneTermEqualSelector("metadata.name", configmapName),
	)

	opts := cache.InformerOptions{
		ListerWatcher: watchlist,
		ObjectType:    &v1.ConfigMap{},
		Handler: cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				logger.Infof("configmap %s added, triggering reload", configmap)
				reloaderChan <- false
			},
			DeleteFunc: func(obj interface{}) {
				logger.Errorf("configmap %s deleted", configmap)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				logger.Infof("configmap %s updated, triggering reload", configmap)
				reloaderChan <- false
			},
		},
		ResyncPeriod: time.Minute,
	}

	_, controller := cache.NewInformerWithOptions(opts)

	stop := make(chan struct{})
	go controller.Run(stop)
	for {
		time.Sleep(time.Second)
	}
}

func k8sGetCm(clientset kubernetes.Interface, configmap string) string {
	var configMap map[string]string

	ctx := context.Background()
	cm := strings.Split(configmap, "/")
	namespace := cm[0]
	configmapName := cm[1]

	cmObj, err := clientset.CoreV1().ConfigMaps(namespace).Get(ctx, configmapName,
		metav1.GetOptions{})

	if err != nil {
		logger.Errorln(err.Error())
	} else {
		configMap = cmObj.Data
	}

	return configMap["config.yml"]
}
