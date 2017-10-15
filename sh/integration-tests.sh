#!/usr/bin/env bash
set -euo pipefail
set -x

helm s3 init s3://test-bucket/charts
if [ $? -ne 0 ]; then
    echo "Failed to initialize repo"
    exit 1
fi

mc ls helm-s3-minio/test-bucket/charts/index.yaml
if [ $? -ne 0 ]; then
    echo "Repository was not actually initialized"
    exit 1
fi

helm repo add test-repo s3://test-bucket/charts
if [ $? -ne 0 ]; then
    echo "Failed to add repo"
    exit 1
fi

# Prepare chart to play with.
helm fetch stable/postgresql --version 0.8.3

helm s3 push postgresql-0.8.3.tgz test-repo
if [ $? -ne 0 ]; then
    echo "Failed to push chart to repo"
    exit 1
fi

mc ls helm-s3-minio/test-bucket/charts/postgresql-0.8.3.tgz
if [ $? -ne 0 ]; then
    echo "Chart was not actually uploaded"
    exit 1
fi

# Update the index so we can find the uploaded chart.
helm repo update

helm search test-repo/postgres | grep -q 0.8.3
if [ $? -ne 0 ]; then
    echo "Failed to find uploaded chart"
    exit 1
fi

helm fetch test-repo/postgresql --version 0.8.3
if [ $? -ne 0 ]; then
    echo "Failed to fetch chart from repo"
    exit 1
fi

rm postgresql-0.8.3.tgz
helm repo remove test-repo

set +x

