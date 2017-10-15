#!/usr/bin/env bash
set -uo pipefail

## Set up

export AWS_ACCESS_KEY_ID=EXAMPLEKEY123
export AWS_SECRET_ACCESS_KEY=EXAMPLESECRET123456
export AWS_DEFAULT_REGION=us-east-1

docker container run -d --name helm-s3-minio \
    -p 9000:9000 \
    -e MINIO_ACCESS_KEY=$AWS_ACCESS_KEY_ID \
    -e MINIO_SECRET_KEY=$AWS_SECRET_ACCESS_KEY \
    minio/minio:latest server /data &>/dev/null

if [ ! -x "$(which mc 2>/dev/null)" ]; then
    go get -d github.com/minio/mc
    (cd ${GOPATH}/src/github.com/minio/mc && make)
fi

mc config host add helm-s3-minio http://localhost:9000 $AWS_ACCESS_KEY_ID $AWS_SECRET_ACCESS_KEY
mc mb helm-s3-minio/test-bucket

go build -o bin/helms3 -ldflags "-X github.com/hypnoglow/helm-s3/pkg/awsutil.awsDisableSSL=true -X github.com/hypnoglow/helm-s3/pkg/awsutil.awsEndpoint=localhost:9000" ./cmd/helms3

## Test

bash "$(dirname ${BASH_SOURCE[0]})/integration-tests.sh"
if [ $? -eq 0 ] ; then
    echo -e "\nAll tests passed!"
fi

## Tear down

docker container rm -f helm-s3-minio &>/dev/null

