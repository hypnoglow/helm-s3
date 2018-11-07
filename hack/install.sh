#!/usr/bin/env bash
set -euo pipefail

if [ -n "${HELM_S3_PLUGIN_NO_INSTALL_HOOK:-}" ]; then
    echo "Development mode: not downloading versioned release."
    exit 0
fi

validate_checksum() {
    if ! grep -q ${1} ${2}; then
        echo "Invalid checksum" > /dev/stderr
        exit 1
    fi
    echo "Checksum is valid."
}

on_exit() {
    exit_code=$?
    if [ ${exit_code} -ne 0 ]; then
        echo "helm-s3 install hook failed. Please remove the plugin using 'helm plugin remove s3' and install again." > /dev/stderr
    fi
    exit ${exit_code}
}
trap on_exit EXIT

version="$(cat plugin.yaml | grep "version" | cut -d '"' -f 2)"
echo "Downloading and installing helm-s3 v${version} ..."

binary_url=""
if [ "$(uname)" == "Darwin" ]; then
    binary_url="https://github.com/hypnoglow/helm-s3/releases/download/v${version}/helm-s3_${version}_darwin_amd64.tar.gz"
elif [ "$(uname)" == "Linux" ] ; then
    binary_url="https://github.com/hypnoglow/helm-s3/releases/download/v${version}/helm-s3_${version}_linux_amd64.tar.gz"
fi

if [ -z "${binary_url}" ]; then
    echo "Unsupported OS type"
    exit 1
fi
checksum_url="https://github.com/hypnoglow/helm-s3/releases/download/v${version}/helm-s3_${version}_checksums.txt"

mkdir -p "bin"
mkdir -p "releases/v${version}"
binary_filename="releases/v${version}.tar.gz"
checksums_filename="releases/v${version}_checksums.txt"

# Download binary and checksums files.
(
    if [ -x "$(which curl 2>/dev/null)" ]; then
        curl -sSL "${binary_url}" -o "${binary_filename}"
        curl -sSL "${checksum_url}" -o "${checksums_filename}"
    elif [ -x "$(which wget 2>/dev/null)" ]; then
        wget -q "${binary_url}" -O "${binary_filename}"
        wget -q "${checksum_url}" -O "${checksums_filename}"
    else
      echo "ERROR: no curl or wget found to download files." > /dev/stderr
    fi
)

# Verify checksum.
(
    if [ -x "$(which sha256sum 2>/dev/null)" ]; then
        checksum=$(sha256sum ${binary_filename} | awk '{ print $1 }')
        validate_checksum ${checksum} ${checksums_filename}
    elif [ -x "$(which openssl 2>/dev/null)" ]; then
        checksum=$(openssl dgst -sha256 ${binary_filename} | awk '{ print $2 }')
        validate_checksum ${checksum} ${checksums_filename}
    else
        echo "WARNING: no tool found to verify checksum" > /dev/stderr
    fi
)

# Unpack the binary.
tar xzf "${binary_filename}" -C "releases/v${version}"
mv "releases/v${version}/bin/helms3" "bin/helms3"
exit 0
