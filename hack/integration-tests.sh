#!/usr/bin/env bash
set -euo pipefail
set -x

# Set up
BUCKET="test-bucket/charts"
CONTENT_TYPE="application/x-gzip"
MINIO="helm-s3-minio/${BUCKET}"
PUBLISH_URI="http://example.com/charts"
REPO="test-repo"
S3_URI="s3://${BUCKET}"
TEST_CASE=""

function cleanup() {
    rc=$?
    set +x
    rm -f postgresql-0.8.3.tgz
    helm repo remove "${REPO}" &>/dev/null

    if [[ ${rc} -eq 0 ]]; then
        echo -e "\nAll tests passed!"
    else
        echo -e "\nTest failed: ${TEST_CASE}"
    fi
}

trap cleanup EXIT

# Prepare chart to play with.
helm fetch stable/postgresql --version 0.8.3
helm repo remove "${REPO}" &>/dev/null || true

TEST_CASE="helm s3 init"
helm s3 init "${S3_URI}"
mc ls "${MINIO}/index.yaml" &>/dev/null
helm repo add "${REPO}" "${S3_URI}"

TEST_CASE="helm s3 push"
helm s3 push postgresql-0.8.3.tgz "${REPO}"
mc ls "${MINIO}/postgresql-0.8.3.tgz" &>/dev/null
helm search "${REPO}/postgres" | grep -q 0.8.3

TEST_CASE="helm s3 push fails"
! helm s3 push postgresql-0.8.3.tgz "${REPO}" 2>/dev/null

TEST_CASE="helm s3 push --force"
helm s3 push --force postgresql-0.8.3.tgz "${REPO}"

TEST_CASE="helm fetch"
helm fetch "${REPO}/postgresql" --version 0.8.3

TEST_CASE="helm s3 reindex --publish <uri>"
helm s3 reindex "${REPO}" --publish "${PUBLISH_URI}"
mc cat "${MINIO}/index.yaml" | grep -Fqw "${PUBLISH_URI}/postgresql-0.8.3.tgz"
mc stat "${MINIO}/index.yaml" | grep "X-Amz-Meta-Helm-S3-Publish-Uri" | grep -Fqw "${PUBLISH_URI}"

TEST_CASE="helm s3 reindex"
helm s3 reindex "${REPO}"
mc cat "${MINIO}/index.yaml" | grep -Fqw "${S3_URI}/postgresql-0.8.3.tgz"
mc stat "${MINIO}/index.yaml" | grep -w "X-Amz-Meta-Helm-S3-Publish-Uri\s*:\s*$"

TEST_CASE="helm s3 delete"
helm s3 delete postgresql --version 0.8.3 "${REPO}"
! mc ls -q "${MINIO}/postgresql-0.8.3.tgz" 2>/dev/null
! helm search "${REPO}/postgres" | grep -Fq 0.8.3

TEST_CASE="helm s3 push --content-type <type>"
helm s3 push --content-type=${CONTENT_TYPE} postgresql-0.8.3.tgz "${REPO}"
helm search "${REPO}/postgres" | grep -Fq 0.8.3
mc ls "${MINIO}/postgresql-0.8.3.tgz" &>/dev/null
mc stat "${MINIO}/postgresql-0.8.3.tgz" | grep "Content-Type" | grep -Fqw "${CONTENT_TYPE}"

# Cleanup to test published repo
helm repo remove "${REPO}"
mc rm --recursive --force "${MINIO}"

TEST_CASE="helm s3 init --publish <uri>"
helm s3 init "${S3_URI}" --publish "${PUBLISH_URI}"
mc ls "${MINIO}/index.yaml" &>/dev/null
mc stat "${MINIO}/index.yaml" | grep "X-Amz-Meta-Helm-S3-Publish-Uri" | grep -Fqw "${PUBLISH_URI}"
helm repo add "${REPO}" "${S3_URI}"

TEST_CASE="helm s3 push (publish)"
helm s3 push postgresql-0.8.3.tgz "${REPO}"
mc ls "${MINIO}/postgresql-0.8.3.tgz" &>/dev/null
mc cat "${MINIO}/index.yaml" | grep -Fqw "${PUBLISH_URI}/postgresql-0.8.3.tgz"
mc stat "${MINIO}/index.yaml" | grep "X-Amz-Meta-Helm-S3-Publish-Uri" | grep -Fqw "${PUBLISH_URI}"
helm search "${REPO}/postgres" | grep -Fq 0.8.3

TEST_CASE="helm fetch (publish)"
helm fetch "${REPO}/postgresql" --version 0.8.3

TEST_CASE="helm s3 delete (publish)"
helm s3 delete postgresql --version 0.8.3 "${REPO}"
mc stat "${MINIO}/index.yaml" | grep "X-Amz-Meta-Helm-S3-Publish-Uri" | grep -Fqw "${PUBLISH_URI}"
! mc ls -q "${MINIO}/postgresql-0.8.3.tgz" 2>/dev/null
! helm search "${REPO}/postgres" | grep -Fq 0.8.3
