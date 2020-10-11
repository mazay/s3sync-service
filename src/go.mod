//
// s3sync-service - Realtime S3 synchronisation tool
// Copyright (c) 2020  Yevgeniy Valeyev
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
//

module github.com/mazay/s3sync-service

go 1.15

require (
	github.com/aws/aws-sdk-go v1.35.5
	github.com/bxcodec/faker v2.0.1+incompatible
	github.com/prometheus/client_golang v1.7.1
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.14.0 // indirect
	github.com/prometheus/procfs v0.2.0 // indirect
	github.com/radovskyb/watcher v1.0.7
	github.com/sirupsen/logrus v1.7.0
	gopkg.in/yaml.v2 v2.3.0
)
