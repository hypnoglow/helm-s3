---
title: 'Usage'
date: 2023-12-19T00:00:00+00:00
weight: 4
---

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

<!--more-->

## Init

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

## Push

Now you can push your chart to this repo:

```bash
$ helm s3 push ./epicservice-0.7.2.tgz mynewrepo
```

You may want to push the chart with relative URL, see
[Relative chart URLs](/docs/advanced-features/#relative-chart-urls).

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

## Delete

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

## Reindex

If your repository somehow became inconsistent or broken, you can use reindex to
recreate the index in accordance with the charts in the repository.

```bash
$ helm s3 reindex mynewrepo
```

You may want to reindex the repo with relative chart URLs, see
[Relative chart URLs](/docs/advanced-features/#relative-chart-urls).
