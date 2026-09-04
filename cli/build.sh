#!/bin/bash -ex

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd ${DIR}

HOOKS_OUT=${DIR}/../build/snap/meta/hooks
BIN_OUT=${DIR}/../build/snap/bin
mkdir -p ${HOOKS_OUT} ${BIN_OUT}

go vet ./...
go test ./...

CGO_ENABLED=0 go build -o ${HOOKS_OUT}/install ./cmd/install
CGO_ENABLED=0 go build -o ${HOOKS_OUT}/configure ./cmd/configure
CGO_ENABLED=0 go build -o ${HOOKS_OUT}/pre-refresh ./cmd/pre-refresh
CGO_ENABLED=0 go build -o ${HOOKS_OUT}/post-refresh ./cmd/post-refresh
CGO_ENABLED=0 go build -o ${BIN_OUT}/cli ./cmd/cli
CGO_ENABLED=0 go build -o ${BIN_OUT}/backend ./cmd/backend
