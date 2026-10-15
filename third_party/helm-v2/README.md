# Helm v2 compatibility module

This directory contains the subset of Helm v2.17.0 packages required by helm-s3.
It is kept as a local module so the unmaintained `golang.org/x/crypto/openpgp`
implementation used by Helm v2 can be replaced with the maintained,
drop-in-compatible `github.com/ProtonMail/go-crypto/openpgp` fork.

The original Helm source is licensed under the Apache License 2.0; see
[`LICENSE`](LICENSE).
