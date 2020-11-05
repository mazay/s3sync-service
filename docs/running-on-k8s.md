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

The `s3sync-service` was designed with containerisation in mind thus it's really easy to run in Docker or Kubernetes. To simplify things we have prepared [Helm](https://helm.sh/) charts and all in one deployment manifest, however you still have to create PVs/PVCs in order to mount the data volumes on the `s3sync-service` pod.

# Helm

In order to start with helm approach your would need to add the repository with:
```bash
helm repo add s3sync-service https://charts.s3sync-service.org
```

Now let's create a `values.yaml` file for our deployment with the following contents:
```yaml
# You can omit this part if you choose to create the secret manually
# please note the keys we expect in the secret
# or use some other authentication methods
secret:
  AWS_ACCESS_KEY_ID: "AKIAI44QH8DHBEXAMPLE"
  AWS_SECRET_ACCESS_KEY: "je7MtGbClwBF/2Zp9Utk/h3yCo8nvbEXAMPLEKEY"

# PVCs local-path1 and local-path2 should be created separately
# as well at corresponding PVs
volumes:
  - name: local-path1
    persistentVolumeClaim:
      claimName: local-path1
  - name: local-path2
    persistentVolumeClaim:
      claimName: local-path2

volumeMounts:
  - name: local-path1
    mountPath: /local/path1
    readOnly: true
  - name: local-path2
    mountPath: /local/path2
    readOnly: true

# The sites configuration for syncing the mentioned above PVs/PVCs,
# please note that there's no limit on the number of sites so feel free to create as many as you meed
config.sites:
  - local_path: /local/path1
    bucket: backup-bucket-path1
    bucket_region: us-east-1
    storage_class: STANDARD_IA
    exclusions:
      - .[Dd][Ss]_[Ss]tore
      - .[Aa]pple[Dd]ouble
  - local_path: /local/path2
    bucket: backup-bucket-path2
    bucket_path: path2
    exclusions:
      - "[Tt]humbs.db"
```

That would be the minimal set of variables required to start the application.
Now we ready to deploy it:
```bash
helm install s3sync-service -f values.yaml --name s3sync-service
```

At this point our application should be up and running, you can check it with the following command or by examining the pod logs:
```bash
helm status s3sync-service
```

Please check [the Helm Charts](helm-charts.md) documentation on available configuration options.

# Deployment manifest

As an alternative to the helm deployments there's a generated [all-in-one.yaml](https://raw.githubusercontent.com/mazay/s3sync-service/master/deploy/all-in-one.yaml) deployment manifest, however, this approach is a bit more challenging as it requires some modifications in the manifest to be made prior the deployment.

The first part to be adjusted is the secret:
```yaml
...
data:
  token: |-
    AWS_ACCESS_KEY_ID: QUtJQUk0NFFIOERIQkVYQU1QTEU=
    AWS_SECRET_ACCESS_KEY: amU3TXRHYkNsd0JGLzJacDlVdGsvaDN5Q284bnZiRVhBTVBMRUtFWQ==
...
```

This is `base64` encoded secret key for accessing your S3 bucket(s), you can either update these values with the actual keys or you can create them using some more secure tool, for example [sealed-secrets-controller](https://github.com/bitnami-labs/sealed-secrets). If you choose the second option - the secret object can be removed and you just need to pass the right secret name in the following line:
```yaml
...
envFrom:
- secretRef:
    name: s3sync-service-kube-system-secret
...
```

Next you need to create PVs/PVCs and add something like that to the manifest:
```yaml
...
        envFrom:
        - secretRef:
            name: s3sync-service-kube-system-secret
        volumes:
        - name: local-path1
          persistentVolumeClaim:
            claimName: local-path1
        - name: local-path2
          persistentVolumeClaim:
            claimName: local-path2
...
      terminationGracePeriodSeconds: 300
      volumeMounts:
      - name: local-path1
        mountPath: /local/path1
        readOnly: true
      - name: local-path2
        mountPath: /local/path2
        readOnly: true
...
```

And finally add your sites to the configmap:
```yaml
...
data:
  config.yml: |-
    aws_region: us-east-1
    loglevel: info
    upload_queue_buffer: 0
    upload_workers: 10
    checksum_workers: 5
    watch_interval: 1s
    s3_ops_retries: 5
    sites:
    - local_path: /local/path1
      bucket: backup-bucket-path1
      bucket_region: us-east-1
      storage_class: STANDARD_IA
      exclusions:
        - .[Dd][Ss]_[Ss]tore
        - .[Aa]pple[Dd]ouble
    - local_path: /local/path2
      bucket: backup-bucket-path2
      bucket_path: path2
      exclusions:
        - "[Tt]humbs.db"
```

And deploy your resources:
```bash
kubectl apply -f all-in-one.yaml
```

At this point the app should be running on your k8s cluster, feel free to go ahead and check it's logs.

---
