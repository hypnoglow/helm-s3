#!/usr/bin/env sh

# This emulates GOPATH presence for go tool.
# This is need because helm installs plugins into ~/.helm/plugins.

projectRoot="$1"
pkg="$2"

if [ ! -e "${GOPATH}/src/${pkg}" ]; then
    mkdir -p $(dirname "${GOPATH}/src/${pkg}")
    ln -sfn "${projectRoot}" "${GOPATH}/src/${pkg}"
fi

cd "${GOPATH}/src/${pkg}"
go build -o bin/helms3 ./cmd/helms3