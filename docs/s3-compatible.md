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

As of version 0.2.0 s3sync-service supports other S3 compatible services. So far it's been successfully tested with [Backblaze](https://www.backblaze.com/).

Multiple storage providers can be used at the same time, eg. _site_ `A` syncs with S3 bucket while _site_ `B` with Backblaze B2.

# Backblaze site configuration

Please use the [Backblaze official documentation](https://help.backblaze.com/hc/en-us/articles/360047425453) to gather:

 - S3 endpoint
 - Application Key
 - Application Key ID

Those parameters are needed in order to configure your _site_ to use the Backblaze B2 bucket. The `access_key` and `secret_access_key` can be set globally or via environment variables depending on your needs and use case, however, keep in mind that when you sync with different storage providers one of them would have to override global credentials as in the below example:

```yaml
- name: backblaze
  local_path: /local/path
  bucket: backblaze-nucket-name
  endpoint: s3.eu-central-003.backblazeb2.com # your S3 endpoint
  access_key: app_key_id # your Application Key ID
  secret_access_key: app_key # your Application Key
```

**All the rest [site configuration options](configuration.md) are still compatible with Backblaze B2.**
