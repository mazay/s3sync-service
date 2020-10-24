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
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	mv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func k8sWatchPVCs(namespace string) {
	clientset := k8sClientset()

	if namespace == "" {
		namespace = v1.NamespaceAll
	}

	logger.Infoln("starting to watch for PVCs")

	watchlist := cache.NewListWatchFromClient(
		clientset.CoreV1().RESTClient(),
		"persistentvolumeclaims",
		namespace,
		fields.Everything(),
	)

	_, controller := cache.NewInformer(
		watchlist,
		&v1.PersistentVolumeClaim{},
		time.Second*0,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				mObj := obj.(mv1.Object)
				logger.Infof("pvc added: %s/%s", mObj.GetNamespace(), mObj.GetName())
			},
			DeleteFunc: func(obj interface{}) {
				mObj := obj.(mv1.Object)
				logger.Infof("pvc deleted: %s/%s", mObj.GetNamespace(), mObj.GetName())
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				mObj := newObj.(mv1.Object)
				logger.Infof("pvc changed: %s/%s", mObj.GetNamespace(), mObj.GetName())
			},
		},
	)

	stop := make(chan struct{})
	go controller.Run(stop)
	for {
		time.Sleep(time.Second)
	}
}
