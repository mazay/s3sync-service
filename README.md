# s3sync-service

## Description

The tool is aimed to sync data into S3 storage service in background.

## Configuration

Exaple configuration:
```yaml
config:
- local_path: /local/path1
  bucket: backup-bucket-path1
  exclusions:
    - .ds_store
    - .appledouble
- local_path: /local/path2
  bucket: generic-backup-bucket
  bucket_path: /path2
  bucket_region: us-east-1
  exclusions:
    - thumbs.db
- local_path: /local/path3
  bucket: generic-backup-bucket
  bucket_path: /path3
  bucket_region: us-east-1
  exclusions:
    - thumbs.db
```

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| local_path | Local file systme path to be synced with S3 | n/a | yes |
| bucket | S3 bucet name | n/a | yes |
| bucket_path | S3 path prefix | n/a | no |
| bucket_region | S3 bucket region if [environment variables](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html) or [aws cli configuration](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html#cli-quick-configuration) should be overriden | no |
| storage_class | [S3 storage class](https://docs.aws.amazon.com/AmazonS3/latest/dev/storage-class-intro.html#sc-compare) | `STANDARD` | no |
| exclusions | List of regexp exclude filters | n/a | no |
