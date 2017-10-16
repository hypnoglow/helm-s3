#!/usr/bin/env sh

# This emulates GOPATH presence for go tool.
# This is need because helm installs plugins into ~/.helm/plugins.

projectRoot="$1"
pkg="$2"

if [ ! -e "${GOPATH}/src/${pkg}" ]; then
    mkdir -p $(dirname "${GOPATH}/src/${pkg}")
    ln -sfn "${projectRoot}" "${GOPATH}/src/${pkg}"
fi

version="$(cat plugin.yaml | grep "version" | cut -d '"' -f 2)"

cd "${GOPATH}/src/${pkg}"
go build -o bin/helms3 -ldflags "-X main.version=${version}" ./cmd/helms3