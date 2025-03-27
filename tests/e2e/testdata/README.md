# testdata

## PGP

Prepare GnuPG keys for testing:

```shell
./bootstrap-gnupg.sh
```

Sign:

```shell
GNUPGHOME=$(pwd)/gnupg \
helm package foo \
  --version 1.3.1 \
  --sign \
  --key "Test Key (helm-s3)" \
  --keyring "$(pwd)/gnupg/secring.gpg"
```

Verify:

```shell
GNUPGHOME=$(pwd)/gnupg \
helm verify foo-1.3.1.tgz
```
