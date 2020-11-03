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

# Running on k8s

Please check an example deployment manifest below.

- Pay attention to the image tag - it's not recommended to use neither `devel` nor `latest` as those are not providing any stability and rebuilt upon changes in `devel` or `master` branches respectively.
- Please note that you have to take care of creation of `persistentVolume`'s and `persistentVolumeClaim`'s.

It is advised to use `configmap` argument when run on k8s, with this set up `s3sync-service` will ignore the `config` setting and read directly from the specified configmap. If the configmap gets changed during the runtime - `s3sync-service` will perform reload in order to apply the changes.

---

The `resources` allocation is the tricky part and you should play around with your setup in order to figure out the right values, the provided example works fine with syncing 5 sites, with total about 25000 of files in size of around 200GB.

---

```yaml
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: s3sync-service
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: s3sync-service
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
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: s3sync-service
  namespace: kube-system
rules:
  - apiGroups:
      - ""
    resources:
      - configmaps
    resourceNames:
      - "s3sync-service"
    verbs:
      - get
      - watch
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: s3sync-service
  namespace: kube-system
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: s3sync-service
subjects:
  - kind: ServiceAccount
    name: s3sync-service
    namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: s3sync-service
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: s3sync-service
subjects:
  - kind: ServiceAccount
    name: s3sync-service
    namespace: kube-system
---
apiVersion: v1
kind: ConfigMap
metadata:
  labels:
    app: s3sync-service
  name: s3sync-service
  namespace: kube-system
data:
  config.yml: |-
    aws_region: eu-central-1
    upload_workers: 10
    sites:
    - local_path: /local/path1
      bucket: backup-bucket-path1
      bucket_region: us-east-1
      storage_class: STANDARD_IA
      access_key: AKIAI44QH8DHBEXAMPLE
      secret_access_key: je7MtGbClwBF/2Zp9Utk/h3yCo8nvbEXAMPLEKEY
      exclusions:
        - .[Dd][Ss]_[Ss]tore
        - .[Aa]pple[Dd]ouble
    - local_path: /local/path2
      bucket: backup-bucket-path2
      bucket_path: path2
      exclusions:
        - "[Tt]humbs.db"
---
apiVersion: v1
kind: Secret
metadata:
  name: s3sync-service
  namespace: kube-system
data:
  AWS_ACCESS_KEY_ID: QUtJQUk0NFFIOERIQkVYQU1QTEUK
  AWS_SECRET_ACCESS_KEY: amU3TXRHYkNsd0JGLzJacDlVdGsvaDN5Q284bnZiRVhBTVBMRUtFWQo=
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: s3sync-service
  namespace: kube-system
spec:
  replicas: 1
  selector:
    matchLabels:
      app: s3sync-service
  template:
    metadata:
      labels:
        app: s3sync-service
      annotations:
        prometheus.io/path: "/metrics"
        prometheus.io/port: "9350"
        prometheus.io/scrape: "true"
    spec:
      serviceAccountName: s3sync-service
      containers:
      - image: zmazay/s3sync-service:devel
        name: s3sync-service
        command:
        - "./s3sync-service"
        - "-configmap=kube-system/s3sync-service"
        env:
        - name: AWS_ACCESS_KEY_ID
          valueFrom:
            secretKeyRef:
              name: s3sync-service
              key: AWS_ACCESS_KEY_ID
        - name: AWS_SECRET_ACCESS_KEY
          valueFrom:
            secretKeyRef:
              name: s3sync-service
              key: AWS_SECRET_ACCESS_KEY
        resources:
          limits:
            cpu: 400m
            memory: 512Mi
          requests:
            cpu: 200m
            memory: 384Mi
        ports:
        - containerPort: 8090
          name: http
          protocol: TCP
        volumeMounts:
        - name: config-volume
          mountPath: /opt/s3sync-service
        - name: local-path1
          mountPath: /local/path1
          readOnly: true
        - name: local-path2
          mountPath: /local/path2
          readOnly: true
      terminationGracePeriodSeconds: 300
      volumes:
      - name: local-path1
        persistentVolumeClaim:
          claimName: local-path1
      - name: local-path2
        persistentVolumeClaim:
          claimName: /local/path2
```

Please check [this page](configuration.md) for more details on configuration options.
