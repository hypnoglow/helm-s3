#!/usr/bin/env sh

# This emulates GOPATH presence for go tool.
# This is need because helm installs plugins into ~/.helm/plugins.

projectRoot="$1"
pkg="$2"

if [ ! -e "${GOPATH}/src/${pkg}" ]; then
    mkdir -p $(dirname "${GOPATH}/src/${pkg}")
    ln -sfn "${projectRoot}" "${GOPATH}/src/${pkg}"
fi

version="${HELM_S3_PLUGIN_VERSION:-}"
if [ -z "${version}" ]; then
  version="$(cat plugin.yaml | grep "version" | cut -d '"' -f 2)"
fi

cd "${GOPATH}/src/${pkg}"
go build -o bin/helms3 -ldflags "-X main.version=${version}" ./cmd/helms3
