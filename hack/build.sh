#!/usr/bin/env bash

set \
  -o errexit \
  -o nounset \
  -o pipefail

# This emulates GOPATH presence for go tool.
# This is need because helm installs plugins into ~/.helm/plugins.

projectRoot="$1"
pkg="$2"
gopath="$(go env GOPATH)"

if [ ! -e "${gopath}/src/${pkg}" ]; then
    mkdir -p "$(dirname "${gopath}/src/${pkg}")"
    ln -sfn "${projectRoot}" "${gopath}/src/${pkg}"
fi

version="${HELM_S3_PLUGIN_VERSION:-}"
if [ -z "${version}" ]; then
  version="$(cat plugin.yaml | grep "version" | cut -d '"' -f 2)"
fi

cd "${gopath}/src/${pkg}"
go build -o bin/helm-s3 -ldflags "-X main.version=${version}" ./cmd/helm-s3
