#!/usr/bin/env bash
set -uo pipefail

## Set up

export AWS_ACCESS_KEY_ID=EXAMPLEKEY123
export AWS_SECRET_ACCESS_KEY=EXAMPLESECRET123456
export AWS_DEFAULT_REGION=us-east-1
export AWS_ENDPOINT=localhost:9000
export AWS_DISABLE_SSL=true

docker container run -d --name helm-s3-minio \
    -p 9000:9000 \
    -e MINIO_ACCESS_KEY=$AWS_ACCESS_KEY_ID \
    -e MINIO_SECRET_KEY=$AWS_SECRET_ACCESS_KEY \
    minio/minio:latest server /data &>/dev/null

MCGOPATH=${GOPATH}/src/github.com/minio/mc
if [ ! -x "$(which mc 2>/dev/null)" ]; then
    go get -d github.com/minio/mc
    (cd ${MCGOPATH} && make)
fi

PATH=${MCGOPATH}:${PATH}
mc config host add helm-s3-minio http://localhost:9000 $AWS_ACCESS_KEY_ID $AWS_SECRET_ACCESS_KEY
mc mb helm-s3-minio/test-bucket

go build -o bin/helms3 ./cmd/helms3

## Test

bash "$(dirname ${BASH_SOURCE[0]})/integration-tests.sh"

## Tear down

docker container rm -f helm-s3-minio &>/dev/null

