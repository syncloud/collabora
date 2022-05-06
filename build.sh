#!/bin/bash -ex

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd ${DIR}

BUILD_DIR=${DIR}/build/snap

cp -r bin ${BUILD_DIR}
cp -r config ${BUILD_DIR}
cp -r hooks ${BUILD_DIR}
