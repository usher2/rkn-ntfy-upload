FROM ubuntu:bionic

LABEL maintainer="Phil Kulin <schors@gmail.com>"

RUN groupadd -g 6000 fetch \
        && useradd -m -d /home/fetch -g fetch -u 6000 -s /sbin/nologin fetch \
        && install -d -o fetch -g fetch /var/fetch \
        && apt-get update && apt-get install -q -y --no-install-recommends \
        openssl ca-certificates \
        netbase fetchmail procmail mpack

COPY unpack.sh /home/fetch/
COPY fetchmailrc procmailrc /etc/

RUN chown fetch:fetch /etc/fetchmailrc && chmod 400 /etc/fetchmailrc

VOLUME /home/fetch
USER fetch

CMD ["fetchmail", "-N", "-f", "/etc/fetchmailrc"]

