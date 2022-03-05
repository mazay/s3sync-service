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

# HTTP server

The HTTP server is started by default and listens for port `8090`, which can be changed via [the command line arguments](configuration.md#command-line-args).

| Resource | Response code | Description |
|----------|:-------------:|-------------|
| `/reload` | `200` | Triggers [application reload](how-it-works.md#application-reload), there is an optional URL parameter - `force` |
| `/info` | `200` | Returns some basic info on the application, such as running version, startup time, etc. Used for k8s health checks. |
