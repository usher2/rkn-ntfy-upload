#!/bin/sh

apiurl="https://api.vkostre.org/api-01"

datadir=$1
if [ -z "$datadir" ]; then
        datadir="./"
fi

find ${datadir} -name "*.sig" -print | while read sigfile; do
        datafile=${sigfile%.sig}
        if [ -f "$datafile" ]; then
                echo "${sigfile} ${datafile}"
                result=`curl -f -s -X POST -F "file1=@${datafile}" -F "file2=@${sigfile}" ${apiurl}/upload`
                if [ $? -eq 0 ]; then
                        echo "${datafile} was successfully uploaded!"
                else
                        error=`echo "${result}" | sed -n -e 's/^\s*\"error\"\s*\:\s*\"\([[:alnum:]]\+\).*/\1/p'`
                        echo "Something wrong with ${sigfile} uploading: ${error}"
                        continue
                fi
                task=`echo "${result}" | sed -n -e 's/^\s*\"task\"\s*\:\s*\"\([[:alnum:]]\+\).*/\1/p'`
                for i in 1 2 3 4 5 6 7 8 9 10 ; do
                        status=`curl -f -s -X GET ${apiurl}/task/${task} | sed -n -e 's/^\s*\"status\"\s*\:\s*\"\([[:alnum:]]\+\).*/\1/p'`
                        if [ $? -eq 0 ]; then
                                 if [ "x" != "x${status}" -a "$status" != "received" ]; then
                                             echo "(${i}) ${datafile}: $status"
                                             break
                                 fi
                        fi
                        if [ ${i} -eq 10 ]; then
                                echo "Something wrong with ${sigfile} uploading: status timeout: ${i}"
                        else
                                sleep 1
                        fi
                done
        fi
done
