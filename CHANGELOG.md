# Change log

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [0.8.0]

### Added

- Added possibility to enable S3 serverside encryption.
[Refs: [#52](https://github.com/hypnoglow/helm-s3/pull/52)] @nexusix

- Added possibility to specify Content-Type for uploaded charts.
[Refs: [#59](https://github.com/hypnoglow/helm-s3/issues/59) [#60](https://github.com/hypnoglow/helm-s3/pull/60)] @bashims

- Added checksum verification on plugin installation.
[Refs: [#63](https://github.com/hypnoglow/helm-s3/pull/63)]

### Changed

- On `helm s3 reindex`, only `*.tgz` files in the bucket directory are taken into
account, everything else is ignored.
[Refs: [#57](https://github.com/hypnoglow/helm-s3/issues/57) [#58](https://github.com/hypnoglow/helm-s3/pull/58)] @kylehodgetts

- Default Content-Type for uploaded charts is set to `application/gzip`.
[Refs: [#59](https://github.com/hypnoglow/helm-s3/issues/59) [#60](https://github.com/hypnoglow/helm-s3/pull/60)] @bashims

- `make` is no longer required to install the plugin.
[Refs: [#62](https://github.com/hypnoglow/helm-s3/issues/62) [#64](https://github.com/hypnoglow/helm-s3/pull/64)] @willhayslett

## [0.7.0]

### Added

- Added global `--acl` flag to address issues for setups with multiple Amazon 
accounts. Thanks to [@razaj92](https://github.com/razaj92) for the Pull Request!
[Ref: [#37](https://github.com/hypnoglow/helm-s3/issues/37)]
- Added `--dry-run` flag to `helm s3 push` command. It simulates a push, but doesn't 
actually touch anything. This option is useful, for example, to indicate if 
a chart upload would fail due to the version not being changed. 
[Ref: [#44](https://github.com/hypnoglow/helm-s3/issues/44)]
- Added `--ignore-if-exists` flag to `helm s3 push` command. It allows to exit 
normally without triggering an error if the pushed chart already exists. A clean
exit code may be useful to avoid some error management in the CI/CD. 
[Ref: [#41](https://github.com/hypnoglow/helm-s3/issues/41)]

### Changed

- Moved `helm s3 reindex` command out of beta, as it seems there are no more 
issues related to it.
