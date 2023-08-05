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

![Version: 0.4.2](https://img.shields.io/badge/Version-0.4.2-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.4.2](https://img.shields.io/badge/AppVersion-0.4.2-informational?style=flat-square)

# Using the Helm Repository

The following command can be used to add the repository:
```bash
helm repo add s3sync-service https://charts.s3sync-service.org
```

Please check [this page](running-on-k8s.md#helm) for examples.

# Configuration

## Requirements

Kubernetes: `>=1.13.10-0`

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | [affinity](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#affinity-and-anti-affinity) settings |
| config.access_key | string | `""` | [global](configuration.md#global-configuration-options) AWS access key ID settings |
| config.aws_region | string | `"us-east-1"` | [global](configuration.md#global-configuration-options) AWS Region settings |
| config.checksum_workers | int | `5` | number of the checksum workers |
| config.loglevel | string | `"info"` | logging level |
| config.s3_ops_retries | int | `5` | [global](configuration.md#global-configuration-options) S3 retries settings |
| config.secret_access_key | string | `""` | [global](configuration.md#global-configuration-options) AWS secret access key settings |
| config.sites | object | `{}` | list of site configuration options, check the [documentation](configuration.md#site-configuration-options) for details |
| config.upload_queue_buffer | int | `0` | the upload queue buffer, check the [documentation](configuration.md#global-configuration-options) for details |
| config.upload_workers | int | `10` | number of the upload workers |
| config.watch_interval | string | `"1s"` | [global](configuration.md#global-configuration-options) watch interval settings |
| configmap.watch | bool | `true` | enable the [configmap watch](k8s-integration.md) feature |
| createRbac | bool | `true` | set to false if you not planning on using configmap watch functionality or want to create RBAC objects manually |
| httpServer.enable | bool | `true` | enable the s3sync-service [http service](http-server.md) |
| httpServer.port | int | `8090` | listen port for the s3sync http service |
| image.pullPolicy | string | `"IfNotPresent"` | image pull policy |
| image.repository | string | `"quay.io/s3sync-service/s3sync-service"` | docker repository, uses `quay.io` mirror by default |
| image.tag | string | `""` | overrides the image tag whose default is the chart appVersion |
| imagePullSecrets | list | `[]` | might be useful when using private registry |
| nodeSelector | object | `{}` | [nodeSelector](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#nodeselector) for the pod |
| podAnnotations | object | `{}` | extra pod annotations |
| podSecurityContext | object | `{}` | the [pod security context](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/#set-the-security-context-for-a-pod) |
| prometheusExporter.enable | bool | `true` | enable built-in prometheus exporter |
| prometheusExporter.path | string | `"/metrics"` | netrics path for the prometheus exporter |
| prometheusExporter.port | int | `9350` | listen port for the built-in prometheus exporter |
| resources | object | `{}` | container [resources allocation](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/) |
| secret.AWS_ACCESS_KEY_ID | string | `""` | AWS access Key ID, omit if you want to create the secret separately |
| secret.AWS_SECRET_ACCESS_KEY | string | `""` | AWS secret access key, omit if you want to create the secret separately |
| secret.name | string | `""` | k8s secret name containing `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`, this needed only if you want to create the secret separately |
| securityContext | object | `{}` | the [container security context](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/#set-the-security-context-for-a-container) |
| serviceAccountName | string | `""` | ServiceAccount name if was created manually |
| tolerations | list | `[]` | pod [tolerations](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/) |
| volumeMounts | object | `{}` | the [volumeMounts](https://kubernetes.io/docs/concepts/storage/volumes/#background) definitions |
| volumes | object | `{}` | the pod [volumes](https://kubernetes.io/docs/concepts/storage/volumes/) definitions |

----------------------------------------------
Autogenerated from chart metadata using [helm-docs v1.11.0](https://github.com/norwoodj/helm-docs/releases/v1.11.0)
