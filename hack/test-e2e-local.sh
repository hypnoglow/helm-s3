#!/usr/bin/env bash
set -e -uo pipefail

[ -n "${DEBUG:-}" ] && set -x

## Set up

export AWS_ACCESS_KEY_ID=EXAMPLEKEY123
export AWS_SECRET_ACCESS_KEY=EXAMPLESECRET123456
export AWS_DEFAULT_REGION=us-east-1
export AWS_ENDPOINT=localhost:9000
export AWS_DISABLE_SSL=true

DOCKER_NAME='helm-s3-minio'
RUN="${1:-.*}"

cleanup() {
  if docker container ls | grep -q "${DOCKER_NAME}$" ; then
    docker container rm --force --volumes "${DOCKER_NAME}" &>/dev/null || :
  fi
}

cleanup

on_exit() {
  if [ -z "${SKIP_CLEANUP:-}" ]; then
    cleanup
  fi
}
trap on_exit EXIT

PATH=$(go env GOPATH)/bin:${PATH}
if [ ! -x "$(which mc 2>/dev/null)" ]; then
    echo "[e2e] --> ERROR: mc not found, please install it with \`go install github.com/minio/mc@latest\`"
    exit 1
fi

echo "[e2e] --> Start minio server ..."

docker container run -d --rm --name "${DOCKER_NAME}" \
    -p 9000:9000 \
    -e MINIO_ACCESS_KEY=$AWS_ACCESS_KEY_ID \
    -e MINIO_SECRET_KEY=$AWS_SECRET_ACCESS_KEY \
    minio/minio:latest server /data >/dev/null

echo "[e2e] --> Wait for minio server to become available ..."

# give minio time to become service available.
sleep 3
mc alias set helm-s3-minio http://localhost:9000 $AWS_ACCESS_KEY_ID $AWS_SECRET_ACCESS_KEY
mc mb helm-s3-minio/test-bucket

## Test

echo "[e2e] --> Run tests ..."

go test -v -count=1 ./tests/e2e/... -run "${RUN}"
if [ $? -eq 0 ] ; then
    echo "[e2e] --> All tests passed!"
fi
