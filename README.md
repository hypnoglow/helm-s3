<p align="left"><img src=".github/assets/icon_with_name.png" width="500" alt="helm-s3 Logo"></p>

[![main](https://github.com/hypnoglow/helm-s3/actions/workflows/main.yml/badge.svg)](https://github.com/hypnoglow/helm-s3/actions/workflows/main.yml)
[![release](https://github.com/hypnoglow/helm-s3/actions/workflows/release.yml/badge.svg)](https://github.com/hypnoglow/helm-s3/actions/workflows/release.yml)
[![codecov](https://codecov.io/gh/hypnoglow/helm-s3/branch/master/graph/badge.svg?token=lJqiDsDfPu)](https://codecov.io/gh/hypnoglow/helm-s3)
[![License MIT](https://img.shields.io/badge/license-MIT-blue.svg?style=flat)](LICENSE)
[![GitHub release](https://img.shields.io/github/release/hypnoglow/helm-s3.svg)](https://github.com/hypnoglow/helm-s3/releases)
[![Artifact Hub](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/s3)](https://artifacthub.io/packages/search?repo=s3)

**helm-s3** is a Helm plugin that provides Amazon S3 protocol support.

This allows you to have private or public Helm chart repositories hosted on
Amazon S3. See [this guide](https://docs.aws.amazon.com/prescriptive-guidance/latest/patterns/set-up-a-helm-v3-chart-repository-in-amazon-s3.html) to get a detailed example use case overview.

The plugin supports both Helm v2 and v3.

> [!NOTE]
> The documentation is available on [website](https://helm-s3.hypnoglow.io/).

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
      * [Relative chart URLs](#relative-chart-urls)
      * [Serving charts via HTTP](#serving-charts-via-http)
      * [ACLs](#acl)
      * [Timeout](#timeout)
      * [Using alternative S3-compatible vendors](#using-alternative-s3-compatible-vendors)
      * [Using S3 bucket ServerSide Encryption](#using-s3-bucket-serverside-encryption)
      * [S3 bucket location](#s3-bucket-location)
      * [AWS SSO](#aws-sso)
      * [Signed charts](#signed-charts)
   * [Additional Documentation](#additional-documentation)
   * [Community and Related Projects](#community-and-related-projects)
   * [Contributing](#contributing)
   * [License](#license)

## Install

The installation itself is simple as:

    $ helm plugin install https://github.com/hypnoglow/helm-s3.git

You can install a specific release version:

    $ helm plugin install https://github.com/hypnoglow/helm-s3.git --version 0.17.1

To use the plugin, you do not need any special dependencies. The installer will
download versioned release with prebuilt binary from [github releases](https://github.com/hypnoglow/helm-s3/releases).
However, if you want to build the plugin from source, or you want to contribute
to the plugin, please see [these instructions](.github/CONTRIBUTING.md).

### Docker Images

[![Docker Pulls](https://img.shields.io/docker/pulls/hypnoglow/helm-s3)](https://hub.docker.com/r/hypnoglow/helm-s3)

The plugin is also distributed as Docker images. Images are pushed to Docker Hub
tagged with plugin release version and suffixed with Helm version. The image
built from master branch is also available, note that it should be only used for
playing and testing, it is **strongly discouraged** to use that image for
production use cases. Refer to https://hub.docker.com/r/hypnoglow/helm-s3 for
details and all available tags.

## Configuration

### AWS Access

To publish charts to buckets and to fetch from private buckets, you need to
provide valid AWS credentials.
You can do this in [the same manner](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html) as for `AWS CLI` tool.

So, if you want to use the plugin and you are already using `AWS CLI` - you are
good to go, no additional configuration required. Otherwise, follow [the official guide](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html)
to set up credentials.

To minimize security issues, remember to configure your IAM user policies
properly. As an example, a setup can provide only read access for users, and
write access for a CI that builds and pushes charts to your repository.

<details>
<summary><b>Example Read Only IAM policy</b></summary>

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
</details>

<details>
<summary><b>Example Read and Write IAM policy</b></summary>

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
</details>

### Helm version mode

The plugin is able to detect if you are using Helm v2 or v3 automatically. If,
for some reason, the plugin does not detect Helm version properly, you can set
`HELM_S3_MODE` environment variable to value `2` or `3` to force the mode.

<details>
<summary>Demonstration</summary>

```bash
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
```
</details>

## Usage

*Note: example commands below are provided for Helm v3. If you still use Helm
v2, see alternatives marked with a tip 💡.*

For now let's omit the process of uploading repository index and charts to s3
and assume you already have your repository `index.yaml` file on s3 under path
`s3://bucket-name/charts/index.yaml` and a chart archive `epicservice-0.5.1.tgz`
under path `s3://bucket-name/charts/epicservice-0.5.1.tgz`.

Add your repository:

```bash
$ helm repo add coolcharts s3://bucket-name/charts
```

Now you can use it as any other Helm chart repository.
Try:

```bash
$ helm search coolcharts
NAME                       	VERSION	  DESCRIPTION
coolcharts/epicservice	    0.5.1     A Helm chart.
```

💡 *For Helm v2, use `helm search coolcharts`*.

To install the chart:

```bash
$ helm install coolchart/epicservice --version "0.5.1"
```

Fetching also works:

```bash
$ helm pull coolchart/epicservice --version "0.5.1"
```

💡 *For Helm v2, use `helm fetch`*.
    
Alternatively:

```bash
$ helm pull s3://bucket-name/charts/epicservice-0.5.1.tgz
```
    
### Init

To create a new repository, use `init`:

```bash
$ helm s3 init s3://bucket-name/charts
```

This command generates an empty **index.yaml** and uploads it to the S3 bucket
under `/charts` key.

To work with this repo by its name, first you need to add it using native helm
command:

```bash
$ helm repo add mynewrepo s3://bucket-name/charts
```

### Push

Now you can push your chart to this repo:

```bash
$ helm s3 push ./epicservice-0.7.2.tgz mynewrepo
```

You may want to push the chart with relative URL, see
[Relative chart URLs](#relative-chart-urls).

On push, both remote and local repo indexes are automatically updated (that
means you don't need to run `helm repo update`).

Your pushed chart is available:

```bash
$ helm search repo mynewrepo
NAME                    VERSION	 DESCRIPTION
mynewrepo/epicservice   0.7.2    A Helm chart.
```

💡 *For Helm v2, use `helm search mynewrepo`*.

Note that the plugin denies push when the chart with the same version already
exists in the repository. This behavior is intentional. It is useful, for
example, in CI automated pushing: if someone forgets to bump chart version - the
chart would not be overwritten. However, in some cases you want to replace
existing chart version. To do so, add `--force` flag to a push command:

```bash
$ helm s3 push --force ./epicservice-0.7.2.tgz mynewrepo
```

To see other available options, use `--help` flag:

```bash
$ helm s3 push --help
```

### Delete

To delete specific chart version from the repository:

```bash
$ helm s3 delete epicservice --version 0.7.2 mynewrepo
```

As always, both remote and local repo indexes updated automatically.

The chart is deleted from the repo:

```bash
$ helm search repo mynewrepo/epicservice
No results found
```

💡 *For Helm v2, use `helm search mynewrepo/epicservice`*

### Reindex

If your repository somehow became inconsistent or broken, you can use reindex to
recreate the index in accordance with the charts in the repository.

```bash
$ helm s3 reindex mynewrepo
```

You may want to reindex the repo with relative chart URLs, see
[Relative chart URLs](#relative-chart-urls).

## Uninstall

```bash
$ helm plugin remove s3
```

Thank you for using the plugin! 👋

## Advanced Features

### Relative chart URLs

Charts can be `push`-ed with `--relative` flag so their URLs in the index file
will be relative to your repository root. This can be useful in various
scenarios, e.g. serving charts via HTTP, serving charts from replicated buckets,
etc.

Also, you can run `reindex` command with `--relative` flag to make all chart
URLs relative in an existing repository.

### Serving charts via HTTP

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

### ACL

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

### Timeout

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

### Using alternative S3-compatible vendors

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

### Using S3 bucket ServerSide Encryption

To enable S3 SSE, export environment variable `AWS_S3_SSE` and set it to desired
type, e.g. `AES256`.

### S3 bucket location

The plugin will look for the bucket in the region inferred by the environment.
This can be controlled by exporting one of `HELM_S3_REGION`, `AWS_REGION` or 
`AWS_DEFAULT_REGION`, in order of precedence.

Since [v0.11.0](https://github.com/hypnoglow/helm-s3/blob/master/CHANGELOG.md#0110---2022-05-24)
the plugin supports dynamic S3 bucket region retrieval, so in most cases you
don't need to provide the region. The plugin will detect it automatically and
work without issues.

### AWS SSO

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

### Signed Charts

The plugin supports signed charts. See [Helm documentation](https://helm.sh/docs/topics/provenance/)
for more information how it works.

The plugin ensures that the `.prov` file is pushed to the S3 bucket along with
the chart. Then, when Helm is invoked with `--verify` flag, the `.prov` file
will be automatically downloaded with the chart and used for verification.

## Additional Documentation

Additional documentation is available in the [docs](docs) directory. This
currently includes:
- Estimated [usage cost calculation](docs/usage-cost.md)
- [Best Practices](docs/best-practice.md) for organizing your repositories.

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
