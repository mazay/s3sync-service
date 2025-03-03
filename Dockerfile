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

FROM --platform=${BUILDPLATFORM:-linux/amd64} public.ecr.aws/docker/library/golang:1.24-alpine3.20 AS builder
ARG RELEASE_VERSION=devel
ARG TARGETOS
ARG TARGETARCH
ENV GOOS=${TARGETOS}
ENV GOARCH=${TARGETARCH}
WORKDIR /go/src/github.com/mazay/s3sync-service
# hadolint ignore=DL3018
RUN apk --no-cache add git
COPY go.mod go.sum ./
RUN go mod download
COPY service ./service
COPY *.go ./
# hadolint ignore=DL3059
RUN go build -ldflags "-X github.com/mazay/s3sync-service/service.version=${RELEASE_VERSION}"

FROM public.ecr.aws/docker/library/alpine:3.21.3
ARG TARGETPLATFORM
LABEL maintainer="Yevgeniy Valeyev <z.mazay@gmail.com>"
# hadolint ignore=DL3018
RUN apk --no-cache add ca-certificates
# hadolint ignore=DL3059
RUN adduser \
    --disabled-password \
    --no-create-home \
    -u 8888 \
    s3sync
USER s3sync
WORKDIR /app/
COPY --from=builder /go/src/github.com/mazay/s3sync-service/s3sync-service .
CMD ["./s3sync-service"]
