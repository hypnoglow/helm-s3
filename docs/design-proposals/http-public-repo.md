# Managing public repositories that expose HTTP(S) schema URLs instead of S3.

## Introduction to the problem

The community requests an ability to upload charts in such a way that
the index contains URLs with HTTP(S) schema instead of S3. This allows
setting up a static file server that serves directly from the S3 bucket, so repository
users can download charts directly using HTTP protocol.

The key problem here is the index file. In the index, we can have either `s3://...`
URLs, if using a repository as a private one, or `http(s)://...`, if we want a
repository to be public. **helm** allows having multiple URLs for one chart in
the index, but if we want a public repository - we don't even want to expose
that we are using S3 as a backend.

The index is static. We cannot upload charts setting `s3://...` URLs and
then change them to corresponding `http(s)://...` on the fly. If index is
served on HTTP(S), then **helm** downloads the index using HTTP(S) protocol
and **helm-s3** is not involved here, so it cannot "inject" into the process
and replace chart urls from `s3://...` to `http(s):/...`.

## Proposal

The proposed solution is to add metadata to the index file, so that **helm-s3** plugin
can distinguish public and private repositories and make correct decisions.

Metadata is managed through AWS S3 object metadata.

The sections below describe the full use case from both point of views:
the repository owner (developer) and the user.

Note that examples below assume that the bucket `s3://awesome-bucket`  is
already configured to serve files under `/charts` key on `https://charts.my-company.tld`.

### Repository owner

1. Initialize the repository.

    ```
    helm s3 init s3://awesome-bucket/charts --publish https://charts.my-company.tld
    ```

    1. **helm-s3** creates an empty `index.yaml` file and uploads it by path `s3://awesome-bucket/charts/index.yaml`.
    It also adds a metadata to index s3 object with the key `helm-s3-public-repo` and value `https://charts.my-company.tld`.

1. Add the repository locally to work with it.

    ```
    helm repo add my-charts s3://awesome-bucket/charts
    ```

    *note that developer (repository owner) adds the repo using s3 protocol*

    1. **helm** downloads index from `s3://awesome-bucket/charts/index.yaml` and puts it into the local cache `/Users/me/.helm/repository/cache/my-charts-index.yaml`.

1. Push chart to the repository.

    ```
    helm s3 push foo-0.5.2.tgz my-charts
    ```

    1. **helm-s3** uploads the file to the s3 bucket by path `s3://awesome-bucket/charts/foo-0.5.2.tgz`.

    1. **helm-s3** updates the remote index, setting the url to public one from metadata `https://charts.my-company.tld/foo-0.5.2.tgz`
    and not the `s3://...` url, because the index contains metadata informing that
    the repo is public.

    1. **helm-s3** also syncs the local cached index `/Users/me/.helm/repository/cache/my-charts-index.yaml` with the remote one.
    Note that because of that, charts are fetch-able by public urls locally.

1. Fetch the chart.

    ```
    helm fetch my-charts/foo-0.5.2.tgz
    ```

    1. **helm** downloads the chart from `https://charts.my-company.tld/foo-0.5.2.tgz`. **helm-s3** is not involved.

1. Reindex the repo.

    ```
    helm s3 reindex my-charts
    ```

    1. **helm-s3** fetches the index, takes the public url from its metadata, and reindexes
    everything in the s3 bucket setting public url for each chart.

1. Delete the chart.

    ```
    helm s3 delete foo --version=0.5.2 my-charts
    ```

    1. **helm-s3** removes the chart from the s3 bucket.

    1. **helm-s3** fetches the index, takes the public url from its metadata, then finds
    the corresponding chart in the index, and deletes it.

    1. **helm-s3** also syncs the local cached index `/Users/me/.helm/repository/cache/my-charts-index.yaml` with the remote one.

#### helm s3 publish

To support publishing existing repositories created with **helm-s3** introduce a new command:

```
helm s3 publish my-repo https://charts.my-company.tld
```

1. **helm-s3** adds a metadata to the index s3 object with the key `helm-s3-public-repo` and value `https://charts.my-company.tld`.

1. **helm-s3** reindexes the repo, replacing `s3://...` urls with public ones for all charts in the index.

### Repository user

Repository users can use it as any other public repository (e.g. official **stable**), because the index is available publicly by url `https://charts.my-company.tld/index.yaml`
and all entries in the index have public urls like `https://charts.my-company.tld/foo-0.5.2.tgz`.

So this typical workflow works out of the box:

```
helm repo add my-charts https://charts.my-company.tld
helm fetch my-charts/foo-0.5.2.tgz
```
