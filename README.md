# s3sync-service

## Description

The tool is aimed to sync data into S3 storage service for multiple _sites_ (path + bucket combination).

## Configuration

Exaple configuration:
```yaml
upload_workers: 10
sites:
- local_path: /local/path1
  bucket: backup-bucket-path1
  bucket_region: us-east-1
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
| upload_workers | Number of upload workers for the service | 10 | no |

### Sites configuration options

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| local_path | Local file systme path to be synced with S3 | n/a | yes |
| bucket | S3 bucet name | n/a | yes |
| bucket_path | S3 path prefix | n/a | no |
| bucket_region | S3 bucket region if [environment variables](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html) or [aws cli configuration](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html#cli-quick-configuration) should be overriden | AWS CLI configuration | no |
| retire_deleted | Remove files from S3 which do not exist locally | `false` | no |
| storage_class | [S3 storage class](https://docs.aws.amazon.com/AmazonS3/latest/dev/storage-class-intro.html#sc-compare) | `STANDARD` | no |
| access_key | AWS Access Key | n/a | no |
| secret_access_key | AWS Secret Access Key | n/a | no |
| exclusions | List of regexp filters for exclusions | n/a | no |
