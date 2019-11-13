FROM golang:1.13.0-alpine AS builder
WORKDIR /go/src/s3sync-service
RUN apk add git
COPY *.go ./
RUN go get -d -v ./...
RUN go install -v ./...
RUN go build

FROM alpine:latest
LABEL maintainer="Yevgeniy Valeyev <z.mazay@gmail.com>"
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /go/src/s3sync-service/s3sync-service .
CMD ["./s3sync-service"]
