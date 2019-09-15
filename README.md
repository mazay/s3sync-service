# s3sync-service

## Description

The `s3sync-service` tool is asynchronously syncing data to S3 storage service for multiple _sites_ (path + bucket combination).

On start, the `s3sync-service` compares local directory contents with S3 (using checksums<->ETag and also validates StorageClass) - copies new files and removes files deleted locally from S3 storage (if `retire_deleted` is set to `true`). Once the initial sync is over the `s3sync-service` start watching the specified local directories and subdirectories for changes in order to perform real-time sync to S3.

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
"./s3sync-service -c /opt/s3sync-service/config.yml"
```

## Configuration

Exaple configuration:

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
  bucket: generic-backup-bucket
  bucket_path: /path2
  exclusions:
    - "[Tt]humbs.db"
- local_path: /local/path3
  bucket: generic-backup-bucket
  bucket_path: /path3
  exclusions:
    - "[Tt]humbs.db"
```

### Generic configuration options

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| loglevel | Logging level, valid options are - `trace`, `debug`, `info`, `warn`, `error`, `fatal`, `panic`. With log level set to `trace` logger will output everything, with `debug` everything apart from `trace` and so on. | `info` | no |
| upload_queue_buffer | Number of elements in the upload queue waiting for processing, might improve performance, however, increases memory usage | `0` | no |
| upload_workers | Number of upload workers for the service | `10` | no |
| watch_interval | Interval for file system watcher in milliseconds | `1000` | no |

### Sites configuration options

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| local_path | Local file system path to be synced with S3, using relative path is known to cause some issues. | n/a | yes |
| bucket | S3 bucet name | n/a | yes |
| bucket_path | S3 path prefix | n/a | no |
| bucket_region | S3 bucket region | n/a | yes |
| retire_deleted | Remove files from S3 which do not exist locally | `false` | no |
| storage_class | [S3 storage class](https://docs.aws.amazon.com/AmazonS3/latest/dev/storage-class-intro.html#sc-compare) | `STANDARD` | no |
| access_key | AWS Access Key | n/a | no |
| secret_access_key | AWS Secret Access Key | n/a | no |
| watch_interval | Interval for file system watcher in milliseconds, overrides global setting | `global.watch_interval` | no |
| exclusions | List of regexp filters for exclusions | n/a | no |
