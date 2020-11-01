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

# s3sync-service

[![Build Status](https://teamcity.yottacloud.org:30000/app/rest/builds/buildType:(id:S3syncService_UnitTesting)/statusIcon)](https://teamcity.yottacloud.org:30000/viewType.html?buildTypeId=S3syncService_UnitTesting&guest=1) [![Go Report Card](https://goreportcard.com/badge/github.com/mazay/s3sync-service)](https://goreportcard.com/report/github.com/mazay/s3sync-service) ![Docker Image Size (latest by date)](https://img.shields.io/docker/image-size/zmazay/s3sync-service)

## Description

The `s3sync-service` tool is asynchronously syncing data to S3 storage service for multiple _sites_ (path + bucket combination).

On start, the `s3sync-service` launches pool of generic upload workers, checksum workers and an FS watcher for each _site_. Once all of the above launched it starts comparing local directory contents with S3 (using checksums<->ETag and also validates StorageClass) which might take quite a while depending on the size of your data directory, disk speed, and available CPU resources.  All the new files or removed files  (if `retire_deleted` is set to `true`) are put into the upload queue for processing. The FS watchers, upload and checksum workers remain running while the main process is working, which makes sure that your data is synced to S3 upon change.

## Running the s3sync-service

1. Create directory with [configuration file](#Configuration), eg. - `/path/to/config/config.yml`.
2. Run docker container with providing AWS credentials via environment variables (IAM role should also do the trick), alternatively credentials could be provided in the [config file](#Configuration), mount directory containing the config file and all of the backup directories listed in the config file:

```bash
docker run --rm -ti \
-e "AWS_ACCESS_KEY_ID=AKIAI44QH8DHBEXAMPLE" \
-e "AWS_SECRET_ACCESS_KEY=je7MtGbClwBF/2Zp9Utk/h3yCo8nvbEXAMPLEKEY" \
-e "AWS_DEFAULT_REGION=us-east-1" \
-v "/path/to/config:/opt/s3sync-service" \
-v "/backup/path:/backup" \
zmazay/s3sync-service \
./s3sync-service -config /opt/s3sync-service/config.yml
```

## Configuration

Example configuration:

```yaml
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
- local_path: /local/path3
  bucket: backup-bucket-path3
  bucket_path: path3
  exclusions:
    - "[Tt]humbs.db"
```

### Command line args

```bash
Usage of ./s3sync-service:
  -config string
    	Path to the config.yml (default "config.yml")
  -http-port string
    	Port for internal HTTP server (default "8090")
  -metrics-path string
    	Prometheus exporter path (default "/metrics")
  -metrics-port string
    	Prometheus exporter port, 0 to disable the exporter (default "9350")
```

### Generic configuration options

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| access_key | Global AWS Access Key | n/a | no |
| secret_access_key | Global AWS Secret Access Key | n/a | no |
| aws_region | AWS region | n/a | no |
| loglevel | Logging level, valid options are - `trace`, `debug`, `info`, `warn`, `error`, `fatal`, `panic`. With log level set to `trace` logger will output everything, with `debug` everything apart from `trace` and so on. | `info` | no |
| upload_queue_buffer | Number of elements in the upload queue waiting for processing, might improve performance, however, increases memory usage | `0` | no |
| checksum_workers | Number of checksum workers for the service | `CPU*2` | no |
| upload_workers | Number of upload workers for the service | `10` | no |
| watch_interval | Interval for file system watcher in format of number and a unit suffix. Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h". | `1000ms` | no |
| s3_ops_retries | Number of retries for upload and delete operations | `2 in k8s, CPU cores * 2 otherwise` | no |

### Site configuration options

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| name | Human friendly site name | `bucket/bucket_path` | no |
| local_path | **Absolute** path on local file system to be synced with S3 | n/a | yes |
| bucket | S3 bucket name | n/a | yes |
| bucket_path | S3 path prefix | n/a | no |
| bucket_region | S3 bucket region | `global.aws_region` | no |
| retire_deleted | Remove files from S3 which do not exist locally | `false` | no |
| storage_class | [S3 storage class](https://docs.aws.amazon.com/AmazonS3/latest/dev/storage-class-intro.html#sc-compare) | `STANDARD` | no |
| access_key | Site AWS Access Key | `global.access_key` | no |
| secret_access_key | Site AWS Secret Access Key | `global.secret_access_key` | no |
| watch_interval | Interval for file system watcher in format of number and a unit suffix. Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h". | `global.watch_interval` | no |
| exclusions | List of regex filters for exclusions | n/a | no |
| s3_ops_retries | Number of retries for upload and delete operations | `global.s3_ops_retries` | no |

### Prometheus metrics

In addition to the default Go metrics, `s3sync-service` exports some custom metrics on default path (`/metrics`) and port (`9350`), check the [command line arguments](#command-line-args) for customisation.

All the custom metrics are exported separately for the configured sites (has `site="my-site"` in labels).

| Metric name | Description | Metric type |
|-------------|-------------|-------------|
| s3sync_data_total_size | Total size of the synced objects | Gauge |
| s3sync_data_objects_count | Total number of the synced objects | Gauge |
| s3sync_errors_count | Number of errors, could be used for alerting | Counter |

### Gotchas

1. Same bucket can be used for multiple sites (local directories) only in case both use some `bucket_path`, otherwise, site using bucket root will delete the data from the prefix used by another site. Setting `retire_deleted` to `false` for the site using bucket root should fix this issue.
1. AWS credentials and region have the following priority:
    1. Site AWS credentials (region)
    1. Global AWS credentials (region)
    1. Environment variables
