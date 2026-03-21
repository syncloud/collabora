#!/bin/sh -xe

DIR=$( cd "$( dirname "$0" )" && pwd )
cd ${DIR}

BUILD_DIR=${DIR}/../build/snap/app

while ! docker build -t collabora:syncloud . ; do
  sleep 1
  echo "retry docker"
done

docker create --name=collabora collabora:syncloud
mkdir -p ${BUILD_DIR}
cd ${BUILD_DIR}
docker export collabora -o app.tar
tar xf app.tar
ls -la ${BUILD_DIR}/usr/bin/cool*
cp ${DIR}/collabora.sh ${BUILD_DIR}
cp ${BUILD_DIR}/usr/bin/coolforkit ${BUILD_DIR}/usr/bin/coolforkit.bin
cp ${DIR}/coolforkit ${BUILD_DIR}/usr/bin/coolforkit
rm -rf app.tar
rm -rf ${BUILD_DIR}/usr/src

