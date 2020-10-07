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
