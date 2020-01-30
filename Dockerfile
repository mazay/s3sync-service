FROM golang:1.13.7-alpine AS builder
WORKDIR /go/src/s3sync-service
RUN apk add git curl
COPY *.go ./
COPY go.mod ./
RUN go mod vendor
RUN go build

FROM alpine:latest
LABEL maintainer="Yevgeniy Valeyev <z.mazay@gmail.com>"
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/s3sync-service/s3sync-service .
CMD ["./s3sync-service"]
