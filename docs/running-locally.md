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

# Running binary

1. [Download binary](https://github.com/mazay/s3sync-service/releases/latest) for your OS/architecture
1. Create configuration file, eg. - `/path/to/config/config.yml`.
1. Provide AWS credentials via [environment variables](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html) ([EC2 instance profile](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_use_switch-role-ec2_instance-profiles.html) should also do the trick), alternatively credentials could be provided in the config file, and start the application:

```bash
export AWS_ACCESS_KEY_ID=AKIAI44QH8DHBEXAMPLE
export AWS_SECRET_ACCESS_KEY=je7MtGbClwBF/2Zp9Utk/h3yCo8nvbEXAMPLEKEY
export AWS_DEFAULT_REGION=us-east-1

./s3sync-service -config /path/to/config/config.yml
```

# Configuration

Example configuration:

```yaml
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
- local_path: /local/path3
  bucket: backup-bucket-path3
  bucket_path: path3
  exclusions:
    - "[Tt]humbs.db"
```

Please check [this page](configuration.md) for more details on configuration options.
