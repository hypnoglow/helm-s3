# Contributing to helm s3 plugin

## Prerequisites

You need to have [task](https://taskfile.dev/) utility to run development tasks.

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
to the installer. `task` does this automatically:

```shell
$ task install
Development mode: not downloading versioned release.
Installed plugin: s3
```

This will first download all dependencies, then build the plugin and install it
to the local helm plugin directory.

If you see no errors - build was successful. Try to run some helm commands
that involve the plugin, or jump straight into plugin development.

## Testing

To run tests, you need to have development environment set up.

This command does some preparations:

```shell
task setup
```

Run unit tests:

```shell
task test-unit
```

Run e2e tests:

```shell
task test-e2e
```
