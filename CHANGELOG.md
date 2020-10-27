# Change log

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [0.10.0] - 2020-10-27

### Added

- Added support for `HELM_S3_REGION` environment variable to override AWS region for bucket location.
[Refs: [#51](https://github.com/hypnoglow/helm-s3/issues/51) [#117](https://github.com/hypnoglow/helm-s3/pull/117)]

- Added support for relative URLs in repository index: charts can be pushed with `--relative` flag.
[Refs: [#121](https://github.com/hypnoglow/helm-s3/pull/121) [#122](https://github.com/hypnoglow/helm-s3/pull/122)]

### Changed

- Update Helm versions the plugin is tested against: v2.16, v2.17, v3.3, v3.4.
[Refs: [#125](https://github.com/hypnoglow/helm-s3/pull/125)]

### Fixed

- Fixed issues when pushing large charts.
[Refs: [#112](https://github.com/hypnoglow/helm-s3/issues/112) [#120](https://github.com/hypnoglow/helm-s3/issues/120) [#124](https://github.com/hypnoglow/helm-s3/pull/124)]

## [0.9.2] - 2020-01-23

### Changed

- Updated AWS SDK to v1.25.50, allowing to use IAM roles for service accounts.
[Refs: [#109](https://github.com/hypnoglow/helm-s3/issues/109) [#110](https://github.com/hypnoglow/helm-s3/pull/110)]

## [0.9.1] - 2020-01-15

### Added

- `helm version` now has optional flag `--mode` that additionally prints the mode (Helm version) in which the plugin operates,
either v2 or v3.
- Added `HELM_S3_MODE` that can be used to forcefully change the mode (Helm version), in case when the plugin does not detect Helm version properly.

### Changed

- Changed the way the plugin detects Helm version. Now it parses `helm version` output instead of checking `helm env`
command existence.

## [0.9.0] - 2019-12-27

### Added

- Helm v3 support. The plugin can detect Helm version and use the corresponding "mode" to operate properly. This means
that Helm v2 is still supported, and will be until the sunset of v2 (approximately until the summer of 2020).
[Refs: [#95](https://github.com/hypnoglow/helm-s3/pull/95) [#98](https://github.com/hypnoglow/helm-s3/pull/98)]

- The plugin is now also distributed as Docker images. Images are pushed to Docker Hub tagged with plugin release 
version and suffixed with Helm version. The image built from master branch is also available, note that it should be
only used for playing and testing, it is **strongly discouraged** to use that image for production use cases. 
Refer to https://hub.docker.com/r/hypnoglow/helm-s3 for details and all available tags.
[Refs: [#79](https://github.com/hypnoglow/helm-s3/issues/79) [#88](https://github.com/hypnoglow/helm-s3/pull/88)]

### Changed

- Migrate to go modules & update Go to 1.12.
[Refs: [#86](https://github.com/hypnoglow/helm-s3/pull/86)] [@moeryomenko](https://github.com/moeryomenko)

- CI now runs tests on multiple Helm versions: v2.14, v2.15, v2.16, v3.0.
[Refs: [#89](https://github.com/hypnoglow/helm-s3/pull/89) [#97](https://github.com/hypnoglow/helm-s3/pull/97)]

- Huge rework on internal Helm integration code to provide support for both Helm v2 and v3.
[Refs: [#95](https://github.com/hypnoglow/helm-s3/pull/95) [#98](https://github.com/hypnoglow/helm-s3/pull/98)]

- Bumped almost all dependencies to more actual versions. Helm SDK now includes both v2.16.1 and v3.0.0.
[Refs: [#74](https://github.com/hypnoglow/helm-s3/pull/74) [#69](https://github.com/hypnoglow/helm-s3/issues/69) [#87](https://github.com/hypnoglow/helm-s3/pull/87)] [@willejs](https://github.com/willejs)

### Fixed

- Fixed incorrect s3 url when "proxy" runs on uninitialized repository.
[Refs: [#77](https://github.com/hypnoglow/helm-s3/issues/77) [#78](https://github.com/hypnoglow/helm-s3/pull/78)] [@horacimacias](https://github.com/horacimacias)

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
