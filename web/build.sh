#!/bin/bash -ex

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd ${DIR}

BUILD_DIR=${DIR}/../build/snap/web

npm ci --no-audit --no-fund
npm run build

rm -rf ${BUILD_DIR}
mkdir -p ${BUILD_DIR}
cp -r ${DIR}/dist ${BUILD_DIR}/
