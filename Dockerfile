#
# s3sync-service - Realtime S3 synchronisation tool
# Copyright (c) 2020  Yevgeniy Valeyev
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with this program.  If not, see <http://www.gnu.org/licenses/>.
#

FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.15.2-alpine AS builder
ARG RELEASE_VERSION=devel
ARG TARGETOS
ARG TARGETARCH
ENV GOOS=${TARGETOS}
ENV GOARCH=${TARGETARCH}
WORKDIR /go/src/s3sync-service
RUN apk add git curl
COPY src/*.go ./
COPY src/go.mod ./
RUN go mod vendor
RUN go build -ldflags "-X main.version=${RELEASE_VERSION}"

FROM --platform=${BUILDPLATFORM:-linux/amd64} alpine:latest
LABEL maintainer="Yevgeniy Valeyev <z.mazay@gmail.com>"
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/s3sync-service/s3sync-service .
CMD ["./s3sync-service"]
