#
# Building docker multi-arch image:
#
# RELEASE_VERSION=1.2.3 make docker-multi-arch
#

DOCKER_PLATFORMS=linux/amd64,linux/arm/v6,linux/arm/v7,linux/arm64/v8,linux/386,linux/ppc64le
DOCKER_BASE_REPO=zmazay/s3sync-service

DOCKER_IMAGE_NAME=${DOCKER_BASE_REPO}:${RELEASE_VERSION}

# Validate docker build arguments
ifndef RELEASE_VERSION
$(error RELEASE_VERSION value is not set)
endif

docker-multi-arch:
	DOCKER_CLI_EXPERIMENTAL=enabled
	docker buildx create --use
	docker buildx build \
	--push \
	--platform $(DOCKER_PLATFORMS) \
	--tag $(DOCKER_IMAGE_NAME) -f ./Dockerfile .
