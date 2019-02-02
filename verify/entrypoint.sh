#!/bin/sh

set -e

if [ ! -z "${DATA_DIRECTORY}" ]; then
	set $@ "-d" "${DATA_DIRECTORY}"
fi

if [ ! -z "${API_URL}" ]; then
	set $@ "-u" "${API_URL}"
fi

if [ ! -z "${CA_PATH}" ]; then
	set $@ "-c" "${CA_PATH}"
fi

if [ ! -z "${TOKEN}" ]; then
	set $@ "-t" "${TOKEN}"
fi

exec $@

