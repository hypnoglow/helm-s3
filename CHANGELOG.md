# Change log

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.16.2] - 2024-07-26

### Changed

- Go updated to 1.22.5
  [Refs: [#448](https://github.com/hypnoglow/helm-s3/pull/448)]

## [0.16.1] - 2024-07-16

### Changed

- `generated` timestamp field in the index file is now updated with current time
on push, reindex and delete.
[Refs: [#325](https://github.com/hypnoglow/helm-s3/pull/325)] thanks to [@josephprem](https://github.com/josephprem)

- Supported (and tested against) Helm versions updated to 3.14.4 and 3.15.2.
[Refs: [#439](https://github.com/hypnoglow/helm-s3/pull/439)]

- Go updated to 1.22.
[Refs: [#436](https://github.com/hypnoglow/helm-s3/pull/436)]

## [0.16.0] - 2023-12-07

### Added

- Added `--verbose` global flag. This flag enables debug output. Currently only
`helm s3 reindex` command supports it, but other commands may be updated in the
future.
[Refs: [#320](https://github.com/hypnoglow/helm-s3/pull/320)]

### Changed

- Supported (and tested against) Helm versions updated to 3.12.3 and 3.13.2.

- Helm version detection now fallbacks to Helm v3 in case of errors, as Helm v2
is deprecated for 3 years already.
[Refs: [#334](https://github.com/hypnoglow/helm-s3/pull/334)]

- Refactored install script so now the plugin does not require _bash_. This
solves a few issues with installing the plugin on some distributions.
[Refs: [#262](https://github.com/hypnoglow/helm-s3/pull/262) [#273](https://github.com/hypnoglow/helm-s3/issues/273) [#241](https://github.com/hypnoglow/helm-s3/pull/241)] thanks to [@yonahd](https://github.com/yonahd) and [@jouve](https://github.com/jouve)

### Fixed

- Fixed issues when the plugin was erroneously detecting Helm v2 instead of Helm v3.
[Refs: [#269](https://github.com/hypnoglow/helm-s3/pull/269) [#221](https://github.com/hypnoglow/helm-s3/issues/221)] thanks to [@clhuang](https://github.com/clhuang)

- Added more context info to errors returned by the plugin, so that it is easier
to understand what went wrong, e.g. what chart failed during `helm s3 reindex`.

## [0.15.1] - 2023-09-20

### Fixed

- Fixed a bug introduced in [0.15.0](#0150---2023-09-16) where `helm s3 init`
failed if the `repositories.yaml` file did not exist, e.g. immediately after
helm was installed.
[Refs: [#301](https://github.com/hypnoglow/helm-s3/pull/301) [#300](https://github.com/hypnoglow/helm-s3/issues/300)]

## [0.15.0] - 2023-09-16

### Added

- `helm s3 init` now supports `--force` and `--ignore-if-exists` flags.
`--force` flag replaces the index file if it exists.
`--ignore-if-exists` flag allows to exit normally without triggering an error if
the index file already exists.
[Refs: [#207](https://github.com/hypnoglow/helm-s3/pull/207) [#73](https://github.com/hypnoglow/helm-s3/issues/73)]

- Added support for AWS IAM Identity Center (aka AWS SSO). See README for instructions.
[Refs: [#274](https://github.com/hypnoglow/helm-s3/pull/274) [#143](https://github.com/hypnoglow/helm-s3/issues/143)]

### Changed

- Supported (and tested against) Helm versions updated to 3.11.3 and 3.12.3.

### Fixed

- Updated dependencies which fixes potential security vulnerabilities.

- Introduced minor documentation improvements.

## [0.14.0] - 2022-08-24

### Changed

- The plugin command line interface changed to use [cobra](https://github.com/spf13/cobra)
instead of kingpin. This provides more compatibility and the same UX as in Helm.
[Refs: [#202](https://github.com/hypnoglow/helm-s3/pull/202)]

- Go updated to 1.19.
[Refs: [#199](https://github.com/hypnoglow/helm-s3/pull/199)]

- Supported (and tested against) Helm versions updated to 3.9.3.
[Refs: [#201](https://github.com/hypnoglow/helm-s3/pull/201)]

- Completed the migration of the CI pipeline to Github Actions.
[Refs: [#166](https://github.com/hypnoglow/helm-s3/pull/166) [#205](https://github.com/hypnoglow/helm-s3/pull/205)]

### Fixed

- Fixed a bug where the plugin failed to fetch charts with special characters
in version (e.g. `v1.0.1+build.123`).
[Refs: [#158](https://github.com/hypnoglow/helm-s3/issues/158) [#204](https://github.com/hypnoglow/helm-s3/pull/204)]

## [0.13.0] - 2022-08-07

### Added

- Added support for Linux ARM.
[Refs: [#175](https://github.com/hypnoglow/helm-s3/pull/175)]

### Changed

- Go updated to 1.17.
[Refs: [#170](https://github.com/hypnoglow/helm-s3/pull/170)] thanks to [@allaryin](https://github.com/allaryin)

- Supported (and tested against) Helm versions updated to 3.8.2 and 3.9.2. 
Deprecated Helm version 2.17.0 is still supported, but will be removed in one of the following plugin releases.
[Refs: [#176](https://github.com/hypnoglow/helm-s3/pull/176) [#194](https://github.com/hypnoglow/helm-s3/pull/194) [#195](https://github.com/hypnoglow/helm-s3/pull/195)]

### Fixed

- Fixed an issue where `helm s3 delete` failed on charts previously pushed with `--relative` flag.
[Refs: [#134](https://github.com/hypnoglow/helm-s3/issues/134) [#191](https://github.com/hypnoglow/helm-s3/pull/191)]

- Fixed an issue where plugin installation failed in Alpine images.
[Refs: [#152](https://github.com/hypnoglow/helm-s3/issues/152) [#159](https://github.com/hypnoglow/helm-s3/pull/159)] thanks to [@sanyer](https://github.com/sanyer)

### Security

- Potentially vulnerable Go module dependencies updated to latest patched versions.
[Refs: [#169](https://github.com/hypnoglow/helm-s3/pull/169)] thanks to [@allaryin](https://github.com/allaryin)

### Miscellaneous

- Added Dependabot to manage dependency updates.
[Refs: [#178](https://github.com/hypnoglow/helm-s3/pull/178)]

- The plugin is published to ArtifactHub.
[Ref: [#173](https://github.com/hypnoglow/helm-s3/pull/173) [#174](https://github.com/hypnoglow/helm-s3/pull/174)]

## [0.12.0] - 2022-06-12

### Added

- Support for Windows.
[Refs: [#160](https://github.com/hypnoglow/helm-s3/pull/160)] thanks to @jwenz723
- Support for Apple Silicon (Mac ARM).
[Refs: [#167](https://github.com/hypnoglow/helm-s3/pull/167)]

### Changed

- Some parts of CI pipelines were moved to GitHub Actions.
[Refs: [#162](https://github.com/hypnoglow/helm-s3/pull/162) [#163](https://github.com/hypnoglow/helm-s3/pull/163) [#165](https://github.com/hypnoglow/helm-s3/pull/165)]
- Go updated to 1.16.
[Refs: [#162](https://github.com/hypnoglow/helm-s3/pull/162)]

## [0.11.0] - 2022-05-24

### Added

- Added dynamic bucket region discovery that removes the need for setting region manually via `HELM_S3_REGION` etc. in the majority of cases.
See [#146](https://github.com/hypnoglow/helm-s3/pull/146) for details on how this works.
[Refs: [#146](https://github.com/hypnoglow/helm-s3/pull/146)] Many thanks to @pregnor

### Changed

- Set supported Helm versions to v2.17, v3.4, v3.5.
[Refs: [#137](https://github.com/hypnoglow/helm-s3/pull/137)]

- Integration tests were reworked into Go e2e tests, all legacy tests removed.
[Refs: [#136](https://github.com/hypnoglow/helm-s3/pull/136)]

- AWS SDK updated to v1.37.18 to support AWS SSO
[Refs: [#123](https://github.com/hypnoglow/helm-s3/pull/123) [#138](https://github.com/hypnoglow/helm-s3/pull/138)]

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
