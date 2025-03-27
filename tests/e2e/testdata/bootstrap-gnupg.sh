#!/usr/bin/env bash
set -euo pipefail

PGP_DIR="$(dirname "$0")/pgp"
GNUPGHOME="$(dirname "$0")/gnupg"

echo "[*] Setting up GNUPGHOME at: $GNUPGHOME"
rm -rf "$GNUPGHOME"
mkdir -p "$GNUPGHOME"
chmod 700 "$GNUPGHOME"

echo "[*] Initializing keyring..."
gpg --homedir "$GNUPGHOME" --list-keys > /dev/null

echo "[*] Importing public key..."
gpg --homedir "$GNUPGHOME" --import "$PGP_DIR/test-public.gpg"

echo "[*] Importing private key..."
gpg --homedir "$GNUPGHOME" --import "$PGP_DIR/test-private.gpg"

echo "[*] Setting ultimate trust level..."
gpg --homedir "$GNUPGHOME" --import-ownertrust < <(
  gpg --homedir "$GNUPGHOME" --list-keys --with-colons | awk -F: '/^fpr:/ { print $10 ":6:" }'
)

echo "[*] Exporting legacy keyring..."
gpg --homedir "$GNUPGHOME" --export > "$GNUPGHOME/pubring.gpg"
gpg --homedir "$GNUPGHOME" --export-secret-keys > "$GNUPGHOME/secring.gpg"

echo "[✓] GNUPGHOME is ready."
