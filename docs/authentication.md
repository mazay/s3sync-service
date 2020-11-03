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

# Authentication

The application support the following ways of providing AWS credentials and region, they all can coexist but have different level of priority.

1. [Site AWS credentials (region)](configuration.md#site-configuration-options) - overrides any of the below for specific site
1. [Global AWS credentials (region)](configuration.md#global-configuration-options) - used for all sites, which don't have credentials set
1. [Environment variables](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html) / [EC2 instance profile](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_use_switch-role-ec2_instance-profiles.html) - used if none of the above is set
