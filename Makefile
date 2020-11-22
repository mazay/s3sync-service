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

DOCKER_PLATFORMS=linux/amd64,linux/arm/v6,linux/arm/v7,linux/arm64/v8,linux/386,linux/ppc64le
DOCKER_IMAGE_NAME=${DOCKER_BASE_REPO}:${RELEASE_VERSION}

GOLANG_OS_LIST=freebsd linux windows darwin
GOLANG_ARCH_LIST=386 amd64 arm

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

# Generates a set of build args
go-build-args := $(foreach OS,$(GOLANG_OS_LIST), \
	$(foreach ARCH,$(if $(filter $(OS),darwin), \
		$(filter-out 386 arm,$(GOLANG_ARCH_LIST)),$(GOLANG_ARCH_LIST)), \
		$(OS)-$(ARCH)))

build:
	$(eval FILENAME := $(call get-filename,$(OS)))
	go build -o $(FILENAME) -ldflags \
	"-X main.version=${RELEASE_VERSION}" ./src/

build-all: $(go-build-args)
$(go-build-args):
	$(eval OS := $(word 1,$(subst -, ,$@)))
	$(eval ARCH := $(word 2,$(subst -, ,$@)))
	$(eval FILENAME := $(call get-filename,$(OS)))
	GOOS=$(OS) GOARCH=$(ARCH) go build -o $(FILENAME) -ldflags \
	"-X main.version=${RELEASE_VERSION}" ./src/ && \
	tar -czvf s3sync-service-${RELEASE_VERSION}-$(OS)-$(ARCH).tar.gz $(FILENAME) && \
	rm $(FILENAME)

clean:
	rm -rf ./s3sync-service*

docker:
	DOCKER_CLI_EXPERIMENTAL=enabled
	docker buildx create --use
	docker buildx build \
	--build-arg RELEASE_VERSION=${RELEASE_VERSION} \
	--push \
	--tag $(DOCKER_IMAGE_NAME) -f ./Dockerfile .
	docker buildx rm

docker-multi-arch:
	DOCKER_CLI_EXPERIMENTAL=enabled
	docker run --rm --privileged multiarch/qemu-user-static --reset -p yes
	docker buildx create --use
	docker buildx build \
	--build-arg RELEASE_VERSION=${RELEASE_VERSION} \
	--push \
	--platform $(DOCKER_PLATFORMS) \
	--tag $(DOCKER_IMAGE_NAME) -f ./Dockerfile .
	docker buildx rm

test:
	GOFLAGS="-json" go test ./src/ -coverprofile cover.out
