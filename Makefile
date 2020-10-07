#
# Building docker multi-arch image:
#
# RELEASE_VERSION=1.2.3 make docker-multi-arch
#

DOCKER_PLATFORMS=linux/amd64,linux/arm/v6,linux/arm/v7,linux/arm64/v8,linux/386,linux/ppc64le
DOCKER_BASE_REPO=zmazay/s3sync-service
DOCKER_IMAGE_NAME=${DOCKER_BASE_REPO}:${RELEASE_VERSION}

GOLANG_OS_LIST=freebsd linux windows
GOLANG_ARCH_LIST=386 amd64 arm

# Validate build arguments
ifndef RELEASE_VERSION
$(error RELEASE_VERSION value is not set)
endif

# Generate OS specific filename
ifeq ($(OS), "windows")
	filename=s3sync-service.exe
else
	filename=s3sync-service
endif

# Generates a set of build args
go-build-args := $(foreach OS,$(GOLANG_OS_LIST), \
 $(foreach ARCH,$(GOLANG_ARCH_LIST), \
 	$(OS)-$(ARCH) ) )

build:
	go build -o $(filename) -ldflags \
	"-X main.version=${RELEASE_VERSION}" ./src/

build-all: $(go-build-args)
$(go-build-args):
	$(eval OS := $(word 1,$(subst -, ,$@)))
	$(eval ARCH := $(word 2,$(subst -, ,$@)))
	GOOS=$(OS) GOARCH=$(ARCH) go build -o $(filename) -ldflags \
	"-X main.version=${RELEASE_VERSION}" ./src/ && \
	tar -czvf s3sync-service-${RELEASE_VERSION}-$(OS)-$(ARCH).tar.gz $(filename) && \
	rm $(filename)

clean:
	rm -rf s3sync-service-*.tar.gz

docker-multi-arch:
	DOCKER_CLI_EXPERIMENTAL=enabled
	docker buildx create --use
	docker buildx build \
	--build-arg RELEASE_VERSION=${RELEASE_VERSION} \
	--push \
	--platform $(DOCKER_PLATFORMS) \
	--tag $(DOCKER_IMAGE_NAME) -f ./Dockerfile .
