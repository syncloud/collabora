#!/bin/bash -e
DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

. ${DIR}/loader.sh
find_loader ${DIR}

exec ${LOADER} --library-path $LIBS ${DIR}/usr/bin/coolwsd "$@"
