# TODO

- [x] ~~The code is currently super dirty, need to refactor heavily.~~
- [x] ~~Get rid of Golang dependency. Plugin "install" hook should download
prebuilt **helms3** binary file from github releases.~~
- [x] ~~Make `helm s3` command able to upload the charts to the repo. Remember
that helm has no build-in command like `push` or `publish`, so we need to provide
easy way to push charts to the repository.~~
- [ ] On `helm s3 push` need to check that the file is a valid Helm chart.