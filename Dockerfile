FROM golang:1.9.7-alpine

LABEL maintainer="Yevgeniy Valeyev <z.mazay@gmail.com>"

WORKDIR /go/src/s3sync-service

RUN apk add git && rm -rf /var/cache/apk/*

COPY *.go ./

RUN go get -d -v ./...
RUN go install -v ./...
RUN go build

RUN rm -rf /go/src/github.com /go/src/gopkg.in
RUN rm *.go

CMD ["./s3sync-service"]
