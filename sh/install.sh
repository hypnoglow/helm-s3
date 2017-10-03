#!/usr/bin/env bash

if [ -n "${HELM_S3_PLUGIN_NO_INSTALL_HOOK}" ]; then
    echo "Development mode: not downloading versioned release."
    exit 0
fi

version="$(cat plugin.yaml | grep "version" | cut -d '"' -f 2)"
echo "Downloading and installing helm-s3 v${version} ..."

url=""
if [ "$(uname)" == "Darwin" ]; then
    url="https://github.com/hypnoglow/helm-s3/releases/download/v${version}/helm-s3_${version}_darwin_amd64.tar.gz"
elif [ "$(uname)" == "Linux" ] ; then
    url="https://github.com/hypnoglow/helm-s3/releases/download/v${version}/helm-s3_${version}_linux_amd64.tar.gz"
fi

if [ -z "${url}" ]; then
    echo "Unsupported OS type"
    exit 1
fi

# TODO: verify checksum?

mkdir -p "bin"
mkdir -p "releases/v${version}"

wget -q "${url}" -O "releases/v${version}.tar.gz"
tar xzf "releases/v${version}.tar.gz" -C "releases/v${version}"
mv "releases/v${version}/bin/helms3" "bin/helms3"