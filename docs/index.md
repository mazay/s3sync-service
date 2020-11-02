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

The `s3sync-service` is a lightweight application designed to sync multiple local path with configured [S3](https://aws.amazon.com/s3/) locations (known as _sites_) in realtime - your data getting synced upon change!

You can either use one [S3 bucket](https://docs.aws.amazon.com/AmazonS3/latest/dev/UsingBucket.html) and multiple bucket paths to store your data, or different S3 buckets under one AWS account or even different S3 buckets under different AWS accounts hosted in different [AWS regions](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-regions-availability-zones.html).

There are a variety of options for running the `s3sync-service`:

- [here](running-locally.md) is described how to run it locally
- [this page](running-in-docker.md) has guidelines for running it in [Docker](https://www.docker.com/)
- and [our favourite recipe](running-on-k8s.md) for running it on [Kubernetes](https://kubernetes.io/)

Feel free to check [how it works](how-it-works.md) in more details or familiarise yourself with available [configuration](configuration.md) options.

Found a bug or missing some feature? Please [file an issue](https://github.com/mazay/s3sync-service/issues/new/choose) - we appreciate your feedback.
