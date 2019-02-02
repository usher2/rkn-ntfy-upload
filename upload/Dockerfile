FROM golang:alpine

RUN addgroup -g 5000 upload \
	&& adduser -D -h /dev/null -G upload -u 5000 -s /sbin/nologin upload \
        && install -d -o upload -g upload /var/upload \
	&& install -d -o upload -g upload /var/lib/tasks \
	&& apk add git \
	&& install -d /go/src/app

WORKDIR /go/src/app
COPY code/ .
COPY entrypoint.sh /entrypoint.sh

RUN go get -d -v ./... \
	&& go install -v ./...

USER upload


ENTRYPOINT ["/entrypoint.sh"]

