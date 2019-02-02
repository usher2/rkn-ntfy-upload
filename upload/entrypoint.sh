#!/bin/sh

set -e

args=""

if [ ! -z "${DATA_DIRECTORY}" ]; then
	args="${args} -d ${DATA_DIRECTORY}"
fi

if [ ! -z "${DEBUG_LEVEL}" ]; then
	args="${args} -l ${DEBUG_LEVEL}"
fi

if [ ! -z "${MAX_FILES}" ]; then
	args="${args} -a ${MAX_FILES}"
fi

if [ ! -z "${MIN_FILES}" ]; then
	args="${args} -i ${MIN_FILES}"
fi

if [ ! -z "${MAX_FILE_SIZE}" ]; then
	args="${args} -s ${MAX_FILE_SIZE}"
fi

if [ ! -z "${TASK_COMPLETE_TTL}" ]; then
	args="${args} -c ${TASK_COMPLETE_TTL}"
fi

if [ ! -z "${LISTEN_PORT}" ]; then
	args="${args} -p ${LISTEN_PORT}"
fi

if [ ! -z "${TOKEN}" ]; then
	args="${args} -x ${TOKEN}"
fi

if [ ! -z "${DB_FILE}" ]; then
	args="${args} -b ${DB_FILE}"
fi

/go/bin/app ${args}

