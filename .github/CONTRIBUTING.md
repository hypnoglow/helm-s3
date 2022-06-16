# Contributing to helm s3 plugin

## Development

First of all, clone the repository to your machine.

```shell
$ git clone https://github.com/hypnoglow/helm-s3.git
$ cd helm-s3
```

After that you need to install the plugin from the filesystem.

On regular plugin installation, helm triggers post-install hook
that downloads prebuilt versioned release of the plugin binary and installs it.
To disable this behavior, you need to pass `HELM_S3_PLUGIN_NO_INSTALL_HOOK=true`
to the installer:

```shell
$ HELM_S3_PLUGIN_NO_INSTALL_HOOK=true helm plugin install .
Development mode: not downloading versioned release.
Installed plugin: s3
```

Next, you may want to ensure if you have all prerequisites to build
the plugin from source:

```shell
make deps build-local
```

If you see no output - build was successful. Try to run some helm commands
that involve the plugin, or jump straight into plugin development.

## Testing

Run unit tests:

```shell
make test-unit
```

Run e2e tests locally:

```shell
make test-e2e-local
```
