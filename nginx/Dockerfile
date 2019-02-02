FROM nginx:alpine

LABEL maintainer="Phil Kulin <schors@gmail.com>"

RUN addgroup -S -g 82 www-data \
        && adduser -D -S -h /var/www -G www-data -u 82 -s /sbin/nologin www-data \
        && chown -R www-data:www-data /var/cache/nginx

RUN rm /etc/nginx/conf.d/default.conf
COPY etc/conf.d/ /etc/nginx/conf.d
COPY etc/nginx.conf /etc/nginx/
COPY etc/dhparam.pem /etc/nginx/

