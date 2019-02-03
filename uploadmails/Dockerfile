FROM ubuntu:bionic

LABEL maintainer="Phil Kulin <schors@gmail.com>"

RUN groupadd -g 107 crontab \
        && groupadd -g 6000 fetch \
        && useradd -d /dev/null -g fetch -u 6000 -s /sbin/nologin fetch \
        && install -d -o fetch -g fetch /var/fetch \
        && install -d -o fetch -g fetch /srv/fetch \
        && apt-get update && apt-get install -q -y --no-install-recommends \
        cron \
        ca-certificates \
        openssl libengine-gost-openssl1.1 \
        locales \
        locales-all \
        curl \
        && rm -f /etc/cron.d/*

COPY grab.sh /srv/fetch/grab.sh
COPY cron_entrypoint.sh /entrypoint.sh
COPY crontab /etc/crontab

VOLUME /tmp

ENTRYPOINT ["/entrypoint.sh"]

CMD [ "/usr/sbin/cron", "-f", "-L", "15"]
