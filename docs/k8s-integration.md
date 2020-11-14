<!--
s3sync-service - Realtime S3 synchronisation tool
Copyright (c) 2020  Yevgeniy Valeyev

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
 -->

# Intro

As of version `0.1.0` the `s3sync-service` is shipped with some k8s integration, which allows you to configure the application for reading k8s `configmap` instead of a configuration file. This approach bring some benefits such as:

* no need to mount the `configmap` to a pod
* the application will watch for the `configmap` changes and perform reload if needed

This approach requires RBAC resources allowing `read/watch/list` of the `configmap` and valid `configmap` data structure. Which will be handled automagically if you use [the helm chart](helm-charts.md), otherwise please make sure the data contains `config.yml` and valid [configuration](configuration.md) underneath.

# Manual way

The `configmap` data should be similar to the following:
```yaml
apiVersion: v1
kind: ConfigMap
data:
  config.yml: |-
    aws_region: us-east-1
    sites:
    - name: my-data1
      bucket: my-data-bucket1
      local_path: /my-data1
    - name: my-data2
      bucket: my-data-bucket2
      local_path: /my-data2
      retire_deleted: true
      storage_class: STANDARD_IA
```

And following are the RBAC resources:
```yaml
---
# Source: s3sync-service/templates/serviceaccount.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: s3sync-service-kube-system-serviceaccount
  namespace: kube-system
---
# Source: s3sync-service/templates/clusterrole.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: s3sync-service-kube-system-clusterrole
rules:
  - apiGroups:
      - ""
    resources:
      - configmaps
    verbs:
      - get
      - list
      - watch
---
# Source: s3sync-service/templates/clusterrolebinding.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: s3sync-service-kube-system-clusterrolebinding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: s3sync-service-kube-system-clusterrole
subjects:
  - kind: ServiceAccount
    name: s3sync-service-kube-system-serviceaccount
    namespace: kube-system
---
# Source: s3sync-service/templates/role.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: s3sync-service-kube-system-role
  namespace: kube-system
rules:
  - apiGroups:
      - ""
    resources:
      - configmaps
    resourceNames:
      - "s3sync-service-kube-system-configmap"
    verbs:
      - get
      - watch
---
# Source: s3sync-service/templates/rolebinding.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: s3sync-service-kube-system-rolebinding
  namespace: kube-system
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: s3sync-service-kube-system-role
subjects:
  - kind: ServiceAccount
    name: s3sync-service-kube-system-serviceaccount
    namespace: kube-system
```

Feel free to check out the [example deployment manifest](https://raw.githubusercontent.com/mazay/s3sync-service/master/deploy/all-in-one.yaml) and [the guide](running-on-k8s.md#deployment-manifest).
