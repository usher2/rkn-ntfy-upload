FROM alpine:3.8

LABEL maintainer="Phil Kulin <schors@gmail.com>"

RUN apk add --no-cache openssl ca-certificates fetchmail procmail mpack \
        && addgroup -g 6000 fetch \
        && adduser -D -h /dev/null -G fetch -u 6000 -s /sbin/nologin fetch \
        && install -d -o fetch -g fetch /var/fetch
        && install -d -o fetch -g fetch /srv/fetch

COPY grab.sh unpack.sh /srv/upload/
COPY fetchmailrc procmailrc /etc

USER fetch

CMD ["fetchmail", "-N", "-f", "/etc/fetchmailrc"]

