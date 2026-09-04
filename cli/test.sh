#!/bin/bash -ex

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

BIN=${DIR}/../build/snap/bin
HOOKS=${DIR}/../build/snap/meta/hooks

${BIN}/cli --help
${BIN}/backend --help

for hook in install configure pre-refresh post-refresh; do
  test -x ${HOOKS}/${hook}
done
