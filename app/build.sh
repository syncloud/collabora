#!/bin/sh -xe

DIR=$( cd "$( dirname "$0" )" && pwd )
cd ${DIR}

BUILD_DIR=${DIR}/../build/snap/app
rm -rf ${BUILD_DIR}
mkdir -p ${BUILD_DIR}

cp -r /etc ${BUILD_DIR}
cp -r /opt ${BUILD_DIR}
cp -r /usr ${BUILD_DIR}
cp -r /bin ${BUILD_DIR}
cp -r /lib ${BUILD_DIR}

rm -rf ${BUILD_DIR}/usr/src

cp ${DIR}/collabora.sh ${BUILD_DIR}
cp ${BUILD_DIR}/usr/bin/coolforkit-ns ${BUILD_DIR}/usr/bin/coolforkit-ns.bin
cp ${DIR}/coolforkit-ns ${BUILD_DIR}/usr/bin/coolforkit-ns

test -x ${BUILD_DIR}/usr/bin/coolwsd
test -d ${BUILD_DIR}/opt/collaboraoffice
test -d ${BUILD_DIR}/usr/share/coolwsd/browser
du -sh ${BUILD_DIR}
