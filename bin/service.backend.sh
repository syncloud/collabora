#!/bin/bash -e

set -a
. $SNAP_DATA/config/backend.env
set +a

exec $SNAP/bin/backend
