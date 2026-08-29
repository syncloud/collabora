#!/bin/bash -ex

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

if [[ -z "$4" ]]; then
    echo "usage $0 spec distro app version"
    exit 1
fi

SPEC=$1
DISTRO=$2
APP=$3
VERSION=$4

cd ${DIR}/../test
./deps.sh
py.test -x -s ${SPEC} --distro=${DISTRO} --ver=${VERSION} --app=${APP}
