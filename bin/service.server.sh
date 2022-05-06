#!/bin/bash

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && cd .. && pwd )

if [[ -z "$1" ]]; then
    echo "usage $0 [start]"
    exit 1
fi
ARCH=$(uname -m)
case $1 in
start)
    exec $DIR/app/collabora.sh --disable-cool-user-checking --lo-template-path=$SNAP/app/opt/collaboraoffice --version --config-file=$SNAP_DATA/config/coolwsd.xml
    ;;
*)
    echo "not valid command"
    exit 1
    ;;
esac
