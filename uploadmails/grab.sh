#!/bin/sh

apiurl="https://api.vkostre.org/api-01"

basedir=$1
if [ -z "$basedir" ]; then
        basedir="."
fi

unknowndir="${basedir}/__unknown__"
if [ ! -z "${unknowndir}" ]; then
        mkdir -p "${unknowndir}"
fi
workdir="${basedir}/files"

find "${workdir}" -type d -print | while read datadir; do
        if [ "${workdir}" = "${datadir}" ]; then
                continue
        fi
        find ${datadir} -type f -name "=XUTF-8XBX*X=" -print | while read utffile; do
                fn1=`basename ${utffile}`
                fn2=${fn1%X=}
                fn=${fn2#=XUTF-8XBX}
                newname=`echo -n "${fn}" | base64 -d`
                if [ $? -eq 0 ]; then
                        mv -f "${utffile}" "${datadir}/${newname}"
                fi
        done
        fl=0
        find ${datadir} -type f -name "*.sig" -print | while read sigfile; do
                datafile=${sigfile%.sig}
                if [ -f "$datafile" ]; then
                        sigsize=`stat --printf %s "$sigfile"`
                        datasize=`stat --printf %s "$datafile"`
                        if [ $datasize -gt 1048575 -o $sigsize -gt 1048575 ]; then
                                echo "File too big"
                                continue
                        fi
                        echo "${sigfile} ${datafile}"
                        result=`curl -f -s -X POST -F "file1=@${datafile}" -F "file2=@${sigfile}" ${apiurl}/upload`
                        if [ $? -eq 0 ]; then
                                echo "${datafile} was successfully uploaded!"
                                rm -f "${sigfile}" "${datafile}"
                        else
                                error=`echo "${result}" | sed -n -e 's/^\s*\"error\"\s*\:\s*\"\([[:alnum:]]\+\).*/\1/p'`
                                echo "Something wrong with ${sigfile} uploading: ${error}"
                                if [ "${error}" != "file_too_big" ]; then
                                        fl=1
                                fi
                                continue
                        fi
                fi
        done
        if [ $fl -eq 0 ]; then
                find ${datadir} -type f \( -name "*.sig" -o -name "*.pdf" -o -name "*.docx" -o -name "*.doc" -o -name "*.rtf" \) -print | while read somefile; do
                        mv -f "${somefile}" "${unknowndir}"
                done
                find ${datadir} -name "*" -delete
        fi
done
