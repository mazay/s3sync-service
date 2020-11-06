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

The `s3sync-service` tool is asynchronously syncing data to S3 storage service for multiple _sites_ (path + bucket combination).


# Quickstart

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

# Configuration

Example configuration, check [this](src/example_config.yml) for more details:

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

For more details please refer to [the documentation](https://docs.s3sync-service.org/).

# Contributing

If you feel like contributing to the project - there are [various ways](CONTRIBUTING.md) of doing so.
