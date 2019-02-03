#!/bin/sh

basedir=$1
if [ "x$basedir" = "x" ]; then
        echo "Usage: $0 BASEDIR < <MSG_CONTENT>"
        exit 1
fi

if [ ! -e "$basedir" ]; then
        echo "BASEDIR must be present"
        exit 1
fi

dir="${basedir}/files"
if [ ! -e "$dir" ]; then
        mkdir -p "$dir"
fi
tmpdir="${basedir}/_"
uuid=`cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1`
workdir="${tmpdir}/${uuid}"
mkdir -p "$workdir"

munpack -q -t -C "$workdir"
if [ $? -eq 0 ]; then
        echo "Success"
        mv -f "${workdir}" "${dir}/${uuid}"
        exit 0
else
        echo "Failed"
        exit 1
fi

