#!/bin/sh -ex

DIR=${0%/*}
PATH=${DIR}/../build/bin:${PATH}
export PATH

BUILD_DIR=${DIR}/../build/snap/app

rm -rf ${BUILD_DIR}
mkdir -p ${BUILD_DIR}

for dir in etc opt usr bin lib lib64 sbin; do
  if [ -e /${dir} ]; then
    cp -a /${dir} ${BUILD_DIR}/
  fi
done

find ${BUILD_DIR} -type l | while read -r link; do
  target=$(readlink "$link")
  case "$target" in
    /nix/*)
      rm -f "$link"
      cp -aL "$target" "$link"
      ;;
  esac
done

rm -rf ${BUILD_DIR}/usr/src

cp ${DIR}/collabora.sh ${DIR}/loader.sh ${BUILD_DIR}
cp ${BUILD_DIR}/usr/bin/coolforkit-ns ${BUILD_DIR}/usr/bin/coolforkit-ns.bin
cp ${DIR}/coolforkit-ns ${BUILD_DIR}/usr/bin/coolforkit-ns

case $(uname -m) in
  x86_64) test -f ${BUILD_DIR}/usr/lib/x86_64-linux-gnu/ld-linux-x86-64.so.2 ;;
  aarch64) test -f ${BUILD_DIR}/usr/lib/aarch64-linux-gnu/ld-linux-aarch64.so.1 ;;
  *) echo "unsupported architecture $(uname -m)"; exit 1 ;;
esac

test -x ${BUILD_DIR}/usr/bin/coolwsd
test -d ${BUILD_DIR}/opt/collaboraoffice
test -d ${BUILD_DIR}/usr/share/coolwsd/browser
test -z "$(find ${BUILD_DIR} -type l -exec readlink {} \; | grep '^/nix/' | head -1)"
du -sh ${BUILD_DIR}
