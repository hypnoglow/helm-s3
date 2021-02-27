<p align="left"><img src=".github/assets/icon_with_name.png" width="500" alt="helm-s3 Logo"></p>

[![CircleCI](https://circleci.com/gh/hypnoglow/helm-s3/tree/master.svg?style=shield)](https://circleci.com/gh/hypnoglow/helm-s3/tree/master)
[![License MIT](https://img.shields.io/badge/license-MIT-blue.svg?style=flat)](LICENSE)
[![GitHub release](https://img.shields.io/github/release/hypnoglow/helm-s3.svg)](https://github.com/hypnoglow/helm-s3/releases)

The Helm plugin that provides Amazon S3 protocol support.

This allows you to have private or public Helm chart repositories hosted on Amazon S3. See [this guide](https://docs.aws.amazon.com/prescriptive-guidance/latest/patterns/set-up-a-helm-v3-chart-repository-in-amazon-s3.html) to get a detailed example use case overview.

The plugin supports both Helm v2 and v3 (Helm v3 support is available since [v0.9.0](https://github.com/hypnoglow/helm-s3/releases/tag/v0.9.0)).

## Table of contents

   * [Install](#install)
      * [Docker Images](#docker-images)
   * [Configuration](#configuration)
      * [AWS Access](#aws-access)
      * [Helm version mode](#helm-version-mode)
   * [Usage](#usage)
      * [Init](#init)
      * [Push](#push)
      * [Delete](#delete)
      * [Reindex](#reindex)
   * [Uninstall](#uninstall)
   * [Advanced Features](#advanced-features)
      * [ACLs](#acls)
      * [Using alternative S3-compatible vendors](#using-alternative-s3-compatible-vendors)
      * [Using S3 bucket ServerSide Encryption](#using-s3-bucket-serverside-encryption)
      * [S3 bucket location](#s3-bucket-location)
   * [Additional Documentation](#additional-documentation)
   * [Community and Related Projects](#community-and-related-projects)
   * [Contributing](#contributing)
   * [License](#license)

## Install

The installation itself is simple as:

    $ helm plugin install https://github.com/hypnoglow/helm-s3.git

You can install a specific release version:

    $ helm plugin install https://github.com/hypnoglow/helm-s3.git --version 0.10.0

To use the plugin, you do not need any special dependencies. The installer will
download versioned release with prebuilt binary from [github releases](https://github.com/hypnoglow/helm-s3/releases).
However, if you want to build the plugin from source, or you want to contribute
to the plugin, please see [these instructions](.github/CONTRIBUTING.md).

### Docker Images

[![Docker Pulls](https://img.shields.io/docker/pulls/hypnoglow/helm-s3)](https://hub.docker.com/r/hypnoglow/helm-s3)

The plugin is also distributed as Docker images. Images are pushed to Docker Hub tagged with plugin release 
version and suffixed with Helm version. The image built from master branch is also available, note that it should be
only used for playing and testing, it is **strongly discouraged** to use that image for production use cases. 
Refer to https://hub.docker.com/r/hypnoglow/helm-s3 for details and all available tags.

## Configuration

### AWS Access

To publish charts to buckets and to fetch from private buckets, you need to provide valid AWS credentials.
You can do this in [the same manner](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html) as for `AWS CLI` tool.

So, if you want to use the plugin and you are already using `AWS CLI` - you are
good to go, no additional configuration required. Otherwise, follow [the official guide](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html)
to set up credentials.

To minimize security issues, remember to configure your IAM user policies properly.
As an example, a setup can provide only read access for users, and write access
for a CI that builds and pushes charts to your repository.

**Example Read Only IAM policy**

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "s3:ListBucket",
                "s3:GetObject"
            ],
            "Resource": [
                "arn:aws:s3:::bucket-name",
                "arn:aws:s3:::bucket-name/*"
            ]
        }
    ]
}
```

**Example Read and Write IAM policy**

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "files",
            "Effect": "Allow",
            "Action": [
                "s3:PutObjectAcl",
                "s3:PutObject",
                "s3:GetObjectAcl",
                "s3:GetObject",
                "s3:DeleteObject"
            ],
            "Resource": [
                "arn:aws:s3:::bucket-name/repository-name/*",
                "arn:aws:s3:::bucket-name/repository-name"
            ]
        },
        {
            "Sid": "bucket",
            "Effect": "Allow",
            "Action": "s3:ListBucket",
            "Resource": "arn:aws:s3:::bucket-name"
        }
    ]
}
```

### Helm version mode

The plugin is able to detect if you are using Helm v2 or v3 automatically. If, for some reason, the plugin does not
detect Helm version properly, you can set `HELM_S3_MODE` environment variable to value `2` or `3` to force the mode.

Example:

    # We have Helm version 3:
    $ helm version --short
    v3.0.2+g19e47ee
    
    # For some reason, the plugin detects Helm version badly:
    $ helm s3 version --mode
    helm-s3 plugin version: 0.9.2
    Helm version mode: v2
    
    # Force the plugin to operate in v3 mode:
    $ HELM_S3_MODE=3 helm s3 version --mode
    helm-s3 plugin version: 0.9.2
    Helm version mode: v3

## Usage

*Note: some Helm CLI commands in v3 are incompatible with v2. Example commands below are provided for v2. For commands 
different in v3 there is a tip 💡 below each example.*

For now let's omit the process of uploading repository index and charts to s3 and assume
you already have your repository `index.yaml` file on s3 under path `s3://bucket-name/charts/index.yaml`
and a chart archive `epicservice-0.5.1.tgz` under path `s3://bucket-name/charts/epicservice-0.5.1.tgz`.

Add your repository:

    $ helm repo add coolcharts s3://bucket-name/charts

Now you can use it as any other Helm chart repository.
Try:

    $ helm search coolcharts
    NAME                       	VERSION	  DESCRIPTION
    coolcharts/epicservice	    0.5.1     A Helm chart.

💡 *For Helm v3, use `helm search repo coolcharts`*

To install the chart:

    $ helm install coolchart/epicservice --version "0.5.1"

Fetching also works:

    $ helm fetch coolchart/epicservice --version "0.5.1"
    
Alternatively:

    $ helm fetch s3://bucket-name/charts/epicservice-0.5.1.tgz
    
💡 *For Helm v3, use `helm pull coolchart/epicservice --version "0.5.1"`*

### Init

To create a new repository, use **init**:

    $ helm s3 init s3://bucket-name/charts

This command generates an empty **index.yaml** and uploads it to the S3 bucket
under `/charts` key.

To work with this repo by it's name, first you need to add it using native helm command:

    $ helm repo add mynewrepo s3://bucket-name/charts

### Push

Now you can push your chart to this repo:

    $ helm s3 push ./epicservice-0.7.2.tgz mynewrepo

When the bucket is replicated you should make the index's URLs relative so that the charts can be accessed from a replica bucket.

    $ helm s3 push --relative ./epicservice-0.7.2.tgz mynewrepo

On push, both remote and local repo indexes are automatically updated (that means
you don't need to run `helm repo update`).

Your pushed chart is available:

    $ helm search mynewrepo
    NAME                    VERSION	 DESCRIPTION
    mynewrepo/epicservice   0.7.2    A Helm chart.

💡 *For Helm v3, use `helm search repo mynewrepo`*

Note that the plugin denies push when the chart with the same version already exists
in the repository. This behavior is intentional. It is useful, for example, in
CI automated pushing: if someone forgets to bump chart version - the chart would
not be overwritten.

However, in some cases you want to replace existing chart version. To do so,
add `--force` flag to a push command:

    $ helm s3 push --force ./epicservice-0.7.2.tgz mynewrepo

To see other available options, use `--help` flag:

    $ helm s3 push --help

### Delete

To delete specific chart version from the repository:

    $ helm s3 delete epicservice --version 0.7.2 mynewrepo

As always, both remote and local repo indexes updated automatically.

The chart is deleted from the repo:

    $ helm search mynewrepo/epicservice
    No results found

💡 *For Helm v3, use `helm search repo mynewrepo/epicservice`*

### Reindex

If your repository somehow became inconsistent or broken, you can use reindex to recreate
the index in accordance with the charts in the repository.

    $ helm s3 reindex mynewrepo

When the bucket is replicated you should make the index's URLs relative so that the charts can be accessed from a replica bucket.

    $ helm s3 reindex --relative mynewrepo

## Uninstall

    $ helm plugin remove s3

## Advanced Features

### ACLs

In use cases where you share a repo across multiple AWS accounts,
you may want the ability to define object ACLS to allow charts to persist there
permissions across accounts.
To do so, add the flag `--acl="ACL_POLICY"`. The list of ACLs can be [found here](https://docs.aws.amazon.com/AmazonS3/latest/dev/acl-overview.html#canned-acl):

    $ helm s3 push --acl="bucket-owner-full-control" ./epicservice-0.7.2.tgz mynewrepo

You can also set the default ACL be setting the `S3_ACL` environment variable.

### Using alternative S3-compatible vendors

The plugin assumes Amazon S3 by default. However, it can work with any S3-compatible
object storage, like [minio](https://www.minio.io/), [DreamObjects](https://www.dreamhost.com/cloud/storage/)
and others. To configure the plugin to work alternative S3 backend, just define
`AWS_ENDPOINT` (and optionally `AWS_DISABLE_SSL`):

    $ export AWS_ENDPOINT=localhost:9000
    $ export AWS_DISABLE_SSL=true

See [these integration tests](https://github.com/hypnoglow/helm-s3/blob/master/hack/test-e2e-local.sh) that use local minio docker container for a complete example.

### Using S3 bucket ServerSide Encryption

To enable S3 SSE export environment variable `AWS_S3_SSE` and set it to desired type for example `AES256`.

### S3 bucket location

The plugin will look for the bucket in the region inferred by the environment. If the bucket is in another region, the commands will fail.

This can be controlled by exporting one of `HELM_S3_REGION`, `AWS_REGION` or `AWS_DEFAULT_REGION`, in order of precedence.

## Additional Documentation

Additional documentation is available in the [docs](docs) directory. This currently includes:
- estimated [usage cost calculation](docs/usage-cost.md)
- [best practices](docs/best-practice.md) for organizing your repositories.

## Community and Related Projects

- [Helm | Related Projects and Documentation](https://helm.sh/docs/community/related/)
- [Set up a Helm v3 chart repository in Amazon S3 - AWS Prescriptive Guidance](https://docs.aws.amazon.com/prescriptive-guidance/latest/patterns/set-up-a-helm-v3-chart-repository-in-amazon-s3.html)
- [Deploy Kubernetes resources and packages using Amazon EKS and a Helm chart repository in Amazon S3 - AWS Prescriptive Guidance](https://docs.aws.amazon.com/prescriptive-guidance/latest/patterns/deploy-kubernetes-resources-and-packages-using-amazon-eks-and-a-helm-chart-repository-in-amazon-s3.html)
- [Chart sources - Flux Helm Operator](https://docs.fluxcd.io/projects/helm-operator/en/stable/helmrelease-guide/chart-sources/#extending-the-supported-helm-repository-protocols)
- [How to create a Helm chart repository using Amazon S3](https://andrewlock.net/how-to-create-a-helm-chart-repository-using-amazon-s3/)

## Contributing

Contributions are welcome. Please see [these instructions](.github/CONTRIBUTING.md)
that will help you to develop the plugin.

## License

[MIT](LICENSE)
