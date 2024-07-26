---
title: 'Advanced Features'
date: 2023-12-19T00:00:00+00:00
weight: 6
summary: |
  The plugin has some advanced features that can be useful in various scenarios.
---

## Relative chart URLs

Charts can be `push`-ed with `--relative` flag so their URLs in the index file
will be relative to your repository root. This can be useful in various
scenarios, e.g. serving charts via HTTP, serving charts from replicated buckets,
etc.

Also, you can run `reindex` command with `--relative` flag to make all chart
URLs relative in an existing repository.

## Serving charts via HTTP

You can enable HTTP access to your S3 bucket and serve charts via HTTP URLs, so
your repository users won't have to install this plugin.

To do this, you need your charts to have relative URLs in the index. See
[Relative chart URLs](#relative-chart-urls).

<details>
<summary><b>Example of setting up a public repo using <a href="https://docs.aws.amazon.com/AmazonS3/latest/userguide/VirtualHosting.html">Virtual hosting of buckets</a></b></summary>

1. Create S3 bucket named `example-bucket` in EU (Frankfurt) `eu-central-1` region.

2. Go to "Permissions", edit Bucket Policy:

    ```
    {
      "Version": "2012-10-17",
      "Statement": [
        {
          "Effect": "Allow",
          "Principal": "*",
          "Action": [
            "s3:ListBucket",
            "s3:GetObject"
           ],
          "Resource": [
            "arn:aws:s3:::example-bucket",
            "arn:aws:s3:::example-bucket/*"
          ]
        }
      ]
    }
    ```

3. Initialize repository:

    ```
    $ helm s3 init s3://example-bucket
    Initialized empty repository at s3://example-bucket
    ```

4. Add repository:

    ```
    $ helm repo add example-bucket s3://example-bucket
    "example-bucket" has been added to your repositories
    ```

5. Create demo chart:

    ```
    $ helm create petstore
    Creating petstore

    $ helm package petstore --version 1.0.0
    Successfully packaged chart and saved it to: petstore-1.0.0.tgz
    ```

6. Push chart:

    ```
    $ helm s3 push ./petstore-1.0.0.tgz --relative
    Successfully uploaded the chart to the repository.
    ```

7. The bucket is public and chart repo is set up. Now users can use the repo
   without the need to install helm-s3 plugin.

   Add HTTP repo:

    ```
    $ helm repo add example-bucket-http https://example-bucket.s3.eu-central-1.amazonaws.com/
    "example-bucket-http" has been added to your repositories
    ```

   Search and download charts:

    ```
    $ helm search repo example-bucket-http
    NAME                            CHART VERSION	APP VERSION	DESCRIPTION
    example-bucket-http/petstore	1.0.0       	1.16.0     	A Helm chart for Kubernetes

    $ helm pull example-bucket-http/petstore --version 1.0.0
    ```
</details>

## ACL

In use cases where you share a repo across multiple AWS accounts, you may want
the ability to define object ACLs to allow charts to persist their permissions
across accounts. To do so, add the flag `--acl="ACL_POLICY"`. The list of ACLs
can be [found here](https://docs.aws.amazon.com/AmazonS3/latest/dev/acl-overview.html#canned-acl):

```bash
$ helm s3 push --acl="bucket-owner-full-control" ./epicservice-0.7.2.tgz mynewrepo
```

Note that if you do use ACL, you need to add `--acl` flag for all commands, even
for 'delete', because the index file is still updated when you remove a chart.

You can also set the default ACL be setting the `S3_ACL` environment variable.

## Timeout

The default timeout for all commands is 5 minutes. This is an opinionated
default to be suitable for MFA use, among other things.

If you don't use MFA, it may be reasonable to lower the timeout for most
commands, e.g. to 10 seconds. In contrast, in cases where you want to reindex a
big repository with thousands of charts, you definitely want to increase the
timeout.

Example:

```bash
$ helm s3 push --timeout=10s ./epicservice-0.7.2.tgz mynewrepo
```

## Using alternative S3-compatible vendors

The plugin assumes Amazon S3 by default. However, it can work with any
S3-compatible object storage, like [minio](https://www.minio.io/),
[DreamObjects](https://www.dreamhost.com/cloud/storage/) and others. To
configure the plugin to work alternative S3 backend, just define `AWS_ENDPOINT`
(and optionally `AWS_DISABLE_SSL` if you play with Minio locally):

```bash
$ export AWS_ENDPOINT=localhost:9000
$ export AWS_DISABLE_SSL=true
```

See [these integration tests](https://github.com/hypnoglow/helm-s3/blob/master/hack/test-e2e-local.sh)
that use local minio docker container for a complete example.

## Using S3 bucket ServerSide Encryption

To enable S3 SSE, export environment variable `AWS_S3_SSE` and set it to desired
type, e.g. `AES256`.

## S3 bucket location

The plugin will look for the bucket in the region inferred by the environment.
This can be controlled by exporting one of `HELM_S3_REGION`, `AWS_REGION` or
`AWS_DEFAULT_REGION`, in order of precedence.

Since [v0.11.0](https://github.com/hypnoglow/helm-s3/blob/master/CHANGELOG.md#0110---2022-05-24)
the plugin supports dynamic S3 bucket region retrieval, so in most cases you
don't need to provide the region. The plugin will detect it automatically and
work without issues.

## AWS SSO

The plugin supports AWS IAM Identity Center (aka AWS SSO) authentication.

To use AWS SSO, make sure you [configured it via AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/sso-configure-profile-token.html#sso-configure-profile-token-auto-sso):

```bash
$ aws configure sso
SSO session name (Recommended): my-sso
SSO start URL [None]: https://my-sso-portal.awsapps.com/start
SSO region [None]: us-east-1
SSO registration scopes [None]: sso:account:access

...

CLI default client Region [None]: us-east-1
CLI default output format [None]:
CLI profile name [...]: YOUR-PROFILE-NAME
```

Then, set `AWS_PROFILE` environment variable to the profile name you used in
the previous step:

```bash
$ export AWS_PROFILE=YOUR-PROFILE-NAME
```

Now you can use the plugin as usual.
