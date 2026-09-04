#!/bin/bash -e

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && cd .. && pwd )

exec ${DIR}/app/collabora.sh \
  --disable-cool-user-checking \
  --lo-template-path=${SNAP}/app/opt/collaboraoffice \
  --config-file=${SNAP_DATA}/config/coolwsd.xml
