---
title: 'Configuration'
date: 2023-12-19T00:00:00+00:00
weight: 3
summary: |
  To publish charts to buckets and to fetch from private buckets, you need to
  provide valid AWS credentials.
  You can do this in [the same manner](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html) as for `AWS CLI` tool.

  So, if you want to use the plugin and you are already using `AWS CLI` - you are
  good to go, no additional configuration required. Otherwise, follow [the official guide](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html)
  to set up credentials.
  
---

## AWS Access

To publish charts to buckets and to fetch from private buckets, you need to
provide valid AWS credentials.
You can do this in [the same manner](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html) as for `AWS CLI` tool.

So, if you want to use the plugin and you are already using `AWS CLI` - you are
good to go, no additional configuration required. Otherwise, follow [the official guide](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html)
to set up credentials.

To minimize security issues, remember to configure your IAM user policies
properly. As an example, a setup can provide only read access for users, and
write access for a CI that builds and pushes charts to your repository.

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

## Helm version mode

The plugin is able to detect if you are using Helm v2 or v3 automatically. If,
for some reason, the plugin does not detect Helm version properly, you can set
`HELM_S3_MODE` environment variable to value `2` or `3` to force the mode.

**Demonstration**

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
