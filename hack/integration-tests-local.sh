#!/usr/bin/env bash
set -x -e -uo pipefail

## Set up

export AWS_ACCESS_KEY_ID=EXAMPLEKEY123
export AWS_SECRET_ACCESS_KEY=EXAMPLESECRET123456
export AWS_DEFAULT_REGION=us-east-1
export AWS_ENDPOINT=localhost:9000
export AWS_DISABLE_SSL=true

DOCKER_NAME='helm-s3-minio'

cleanup() {
  if $(docker container ls | grep -q "${DOCKER_NAME}\$") ; then
    docker container rm --force --volumes "${DOCKER_NAME}" || :
  fi
}

cleanup

on_exit() {
  if [ -z "${SKIP_CLEANUP}" ]; then
    cleanup
  fi
}
trap on_exit EXIT

docker container run -d --rm --name helm-s3-minio \
    -p 9000:9000 \
    -e MINIO_ACCESS_KEY=$AWS_ACCESS_KEY_ID \
    -e MINIO_SECRET_KEY=$AWS_SECRET_ACCESS_KEY \
    minio/minio:latest server /data >/dev/null

PATH=${GOPATH}/bin:${PATH}
if [ ! -x "$(which mc 2>/dev/null)" ]; then
    pushd /tmp > /dev/null
    go get github.com/minio/mc
    popd > /dev/null
fi

sleep 3
mc config host add helm-s3-minio http://localhost:9000 $AWS_ACCESS_KEY_ID $AWS_SECRET_ACCESS_KEY
mc mb helm-s3-minio/test-bucket

go build -o bin/helms3 ./cmd/helms3

## Test

$(dirname ${BASH_SOURCE[0]})/integration-tests.sh
if [ $? -eq 0 ] ; then
    echo -e "\nAll tests passed!"
fi
