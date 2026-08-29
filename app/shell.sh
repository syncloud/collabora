#!/bin/sh -ex

DIR=$( cd "$( dirname "$0" )" && pwd )
OUT=${DIR}/../build/bin

rm -rf ${OUT}
mkdir -p ${OUT}
cp /bin/busybox ${OUT}/busybox

for applet in $(${OUT}/busybox --list); do
  ln -sf busybox ${OUT}/${applet}
done

test -x ${OUT}/cp
test -x ${OUT}/find
