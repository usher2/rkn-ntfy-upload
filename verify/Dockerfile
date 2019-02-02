FROM alpine:3.8

LABEL maintainer="Phil Kulin <schors@gmail.com>"

RUN apk add --no-cache python3 openssl ca-certificates \
	&& addgroup -g 5000 upload \
        && adduser -D -h /dev/null -G upload -u 5000 -s /sbin/nologin upload \
        && install -d -o upload -g upload /var/upload \
	&& install -d -o upload -g upload /srv/upload

COPY verify.py /srv/upload/verify.py
COPY entrypoint.sh /entrypoint.sh

USER upload

ENTRYPOINT ["/entrypoint.sh"]

CMD ["python3","/srv/upload/verify.py"]

