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

# Release Notes

## 0.0.7

- Specified `kubeVersion: ">=1.13.10"`
- Use `.Release.Name` for naming all resources created by the chart
- Fixed container level `securityContext`
- Introduced `.Values.httpServer.enable` and `.Values.httpServer.port` to replace `.Values.httpServerPort`
- Added container health checks using embedded HTTP server resource at `/info`

## 0.0.6

First working version, minimal supported application version is `0.1.0`.
