#!/usr/bin/env sh

PROJECT_NAME="helm-s3"
PROJECT_GH="hypnoglow/$PROJECT_NAME"

set -e
set -u

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

initArch() {
    arch=$(uname -m)
    case $arch in
        x86_64|amd64) arch="amd64" ;;
        aarch64|arm64) arch="arm64" ;;
        *)
        echo "Arch '$(uname -m)' not supported!" >&2
        exit 1
        ;;
    esac
}

initOS() {
    os=$(uname -s)
    binary_extension=""
    case "$(uname)" in
        Darwin) os="darwin" ;;
        Linux) os="linux" ;;
        CYGWIN*|MINGW*|MSYS_NT*) os="windows"; binary_extension=".exe" ;;
        *)
        echo "OS '$(uname)' not supported!" >&2
        exit 1
        ;;
    esac
}

on_exit() {
    exit_code=$?
    if [ ${exit_code} -ne 0 ]; then
        echo "${PROJECT_NAME} install hook failed. Please remove the plugin using 'helm plugin remove s3' and install again." > /dev/stderr
    fi
    rm -rf "releases"
    exit ${exit_code}
}
trap on_exit EXIT

version="$(grep "version" plugin.yaml | cut -d '"' -f 2)"
echo "Downloading and installing ${PROJECT_NAME} v${version} ..."

initArch

initOS

binary_url="https://github.com/${PROJECT_GH}/releases/download/v${version}/${PROJECT_NAME}_${version}_${os}_${arch}.tar.gz"
checksum_url="https://github.com/${PROJECT_GH}/releases/download/v${version}/${PROJECT_NAME}_${version}_checksums.txt"

mkdir "releases"
binary_filename="releases/v${version}.tar.gz"
checksums_filename="releases/v${version}_checksums.txt"

# Download binary and checksums files.
(
    if command -v curl >/dev/null 2>&1; then
        curl -sSL "${binary_url}" -o "${binary_filename}"
        curl -sSL "${checksum_url}" -o "${checksums_filename}"
    elif command -v wget >/dev/null 2>&1; then
        wget -q "${binary_url}" -O "${binary_filename}"
        wget -q "${checksum_url}" -O "${checksums_filename}"
    else
      echo "ERROR: no curl or wget found to download files." > /dev/stderr
    fi
)

# Verify checksum.
(
    if command -v sha256sum >/dev/null 2>&1; then
        checksum=$(sha256sum ${binary_filename} | awk '{ print $1 }')
        validate_checksum ${checksum} ${checksums_filename}
    elif command -v openssl >/dev/null 2>&1; then
        checksum=$(openssl dgst -sha256 ${binary_filename} | awk '{ print $2 }')
        validate_checksum ${checksum} ${checksums_filename}
    else
        echo "WARNING: no tool found to verify checksum" > /dev/stderr
    fi
)

# Unpack the binary.
tar xzf "${binary_filename}" "bin/helm-s3${binary_extension}"
