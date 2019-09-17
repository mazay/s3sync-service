FROM golang:1.13.0-alpine

LABEL maintainer="Yevgeniy Valeyev <z.mazay@gmail.com>"

WORKDIR /go/src/s3sync-service

RUN apk add git && rm -rf /var/cache/apk/*

COPY *.go ./

RUN go get -d -v ./...
RUN go install -v ./...
RUN go build

RUN rm -rf /go/src/github.com \
           /go/src/gopkg.in \
           /go/src/golang.org \
           /go/bin/s3sync-service

RUN rm *.go

RUN addgroup -S s3sync && adduser -S s3sync -G s3sync
USER s3sync

CMD ["./s3sync-service"]
