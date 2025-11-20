---
title: 'Install'
date: 2023-12-19T00:00:00+00:00
weight: 2
---

The installation itself is simple as:

    $ helm plugin install https://github.com/hypnoglow/helm-s3.git

You can install a specific release version:

    $ helm plugin install https://github.com/hypnoglow/helm-s3.git --version 0.17.1

<!--more-->

To use the plugin, you do not need any special dependencies. The installer will
download versioned release with prebuilt binary from [github releases](https://github.com/hypnoglow/helm-s3/releases).
However, if you want to build the plugin from source, or you want to contribute
to the plugin, please see [these instructions](https://github.com/hypnoglow/helm-s3/blob/master/.github/CONTRIBUTING.md).

## Docker Images

[![Docker Pulls](https://img.shields.io/docker/pulls/hypnoglow/helm-s3)](https://hub.docker.com/r/hypnoglow/helm-s3)

The plugin is also distributed as Docker images. Images are pushed to Docker Hub
tagged with plugin release version and suffixed with Helm version. The image
built from master branch is also available, note that it should be only used for
playing and testing, it is **strongly discouraged** to use that image for
production use cases. Refer to https://hub.docker.com/r/hypnoglow/helm-s3 for
details and all available tags.
