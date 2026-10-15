# x/crypto compatibility module

This directory contains only the `golang.org/x/crypto` packages required by
helm-s3 and its module graph. The unmaintained OpenPGP packages are intentionally excluded; Helm v2
uses the maintained `github.com/ProtonMail/go-crypto/openpgp` replacement from
the local compatibility module in `third_party/helm-v2`.

The original x/crypto source is licensed under the BSD 3-Clause License; see
[`LICENSE`](LICENSE).
