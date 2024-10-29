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

PLATFORMS=darwin/amd64 darwin/arm64 \
windows/amd64 windows/386 windows/arm \
linux/amd64 linux/386 linux/arm linux/arm64 \
freebsd/amd64 freebsd/386 freebsd/arm freebsd/arm64

# Set docker repo to Docker Hub if nothing else provided
ifndef DOCKER_BASE_REPO
DOCKER_BASE_REPO=zmazay/s3sync-service
endif

# Validate build arguments
ifndef RELEASE_VERSION
$(error RELEASE_VERSION value is not set)
endif

ifeq ($(RELEASE_VERSION), master)
RELEASE_VERSION=latest
endif

# Generate OS specific filename
define get-filename
	$(if $(filter $(1),windows),s3sync-service.exe,s3sync-service)
endef

build:
	$(eval FILENAME := $(call get-filename,$(OS)))
	go build -o bin/$(FILENAME) -ldflags \
	"-X github.com/mazay/s3sync-service/service.version=${RELEASE_VERSION}" && \
	tar -czvf bin/s3sync-service-${RELEASE_VERSION}-$(OS)-$(ARCH).tar.gz bin/$(FILENAME) && \
	rm bin/$(FILENAME)

build-all: $(PLATFORMS)
$(PLATFORMS):
	$(eval OS := $(word 1,$(subst /, ,$@)))
	$(eval ARCH := $(word 2,$(subst /, ,$@)))
	$(eval FILENAME := $(call get-filename,$(OS)))
	GOOS=$(OS) GOARCH=$(ARCH) go build -o bin/$(FILENAME) -ldflags \
	"-X github.com/mazay/s3sync-service/service.version=${RELEASE_VERSION}" && \
	tar -czvf bin/s3sync-service-${RELEASE_VERSION}-$(OS)-$(ARCH).tar.gz bin/$(FILENAME) && \
	rm bin/$(FILENAME)

clean:
	rm -rf bin/s3sync-service*

test:
	GOFLAGS="-json" go test -timeout 5m ./... -coverprofile cover.out
