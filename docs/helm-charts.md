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

# Using the Helm Repository

The following command can be used to add the repository:
```bash
helm repo add s3sync-service https://charts.s3sync-service.org
```

Please check [this page](running-on-k8s.md#helm) for examples.

# Configuration

| Parameter | Description | Required | Default |
|-----------|-------------|----------|---------|
| **image.repository** | The image repository URL, useful if you want to use your registry | no | `zmazay/s3sync-service` |
| **image.pullPolicy** | The image pull policy for the deployment | no | `IfNotPresent` |
| **image.tag** | The `s3sync-service` tag, useful if you want to override default tag for the chart release | no | `depends on the chart release, >=0.1.0` |
| **createRbac** | Set to `false` if you don't want to use [watch for configmap changes](how-it-works.md#application-reload) or willing to create RBAC manually | no | `true` |
| **serviceAccountName** | Ability to use manually created `ServiceAccount` | no | `""` |
| **imagePullSecrets** | The registry secret if private one is being used | no | `[]` |
| **podAnnotations** | A map of extra pod annotation to be added | no | `{}` |
| **podSecurityContext** | Defines the pod securityContext | no | `{}` |
| **securityContext** | Security context to be set on the container | no | `{}` |
| **resources** | Resources requests/limits for the container | no | `{}` |
| **nodeSelector** | Node labels for pod assignment | no | `{}` |
| **tolerations** | Node tolerations for pod assignment | no | `[]` |
| **affinity** | Node affinity for pod assignment | no | `{}` |
| **httpServerPort** | Listen port for the `s3sync-service` HTTP server | no | `8090` |
| **prometheusExporter.enable** | Enable the embedded Prometheus exporter | no | `true` |
| **prometheusExporter.port** | Listen port for the the embedded Prometheus exporter | no | `9350` |
| **prometheusExporter.path** | Metrics path for the the embedded Prometheus exporter | no | `/metrics` |
| **configmap.name** | Name of a configmap if created manually | no | `""` |
| **configmap.watch** | Enable configmap watch feature, requires RBAC | no | `true` |
| **config.access_key** | AWS Access Key, could be provided here or in secret, also both could be omitted - (authentication)[authentication.md] | no | `""` |
| **config.secret_access_key** | AWS Secret Access Key, could be provided here or in secret, also both could be omitted - (authentication)[authentication.md] | no | `""` |
| **config.aws_region** | Global AWS region setting | no | `us-east-1` |
| **config.loglevel** | Logging level | no | `info` |
| **config.upload_queue_buffer** | Number of elements in the upload queue waiting for processing | no | `0` |
| **config.upload_workers** | Number of upload workers | no | `10` |
| **config.checksum_workers** | Number of checksum workers | no | `5` |
| **config.watch_interval** | Interval for file system watcher | no | `1s` |
| **config.s3_ops_retries** | Number of retries for upload and delete operations | no | `1s` |
| **config.sites** | List of the site configurations, check [this](configuration.md) for available options | no | `1s` |
| **secret.customName** | Name of a secret object if managed separately | no | `""` |
| **secret.AWS_ACCESS_KEY_ID** | AWS Access Key, will be used to create secret object if provided | no | `""` |
| **secret.AWS_SECRET_ACCESS_KEY** | AWS Secret Access Key, will be used to create secret object if provided | no | `""` |
| **volumes** | A map of volumes (PVCs) to be attached to the container and used for syncing the data | **yes** | `{}` |
| **volumeMounts** | A map of volumeMounts for the listed above volumes | **yes** | `{}` |
