#!/bin/bash -xe

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd ${DIR}

BUILD_DIR=${DIR}/../build/snap/app

docker ps -a -q --filter ancestor=collabora:syncloud --format="{{.ID}}" | xargs docker stop | xargs docker rm || true
docker rmi collabora:syncloud || true
docker build -t collabora:syncloud .
docker create --name=collabora collabora:syncloud
mkdir -p ${BUILD_DIR}
cd ${BUILD_DIR}
docker export collabora -o app.tar
tar xf app.tar
cp ${DIR}/collabora.sh ${BUILD_DIR}
cp ${BUILD_DIR}/usr/bin/coolforkit ${BUILD_DIR}/usr/bin/coolforkit.bin
cp ${DIR}/coolforkit ${BUILD_DIR}/usr/bin/coolforkit
rm -rf app.tar
rm -rf ${BUILD_DIR}/usr/src
