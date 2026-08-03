#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"

DIST=katex.min.js
MANIFEST=katex.sum

if command -v sha256sum >/dev/null 2>&1; then
  sha256() { sha256sum "$@"; }
else
  sha256() { shasum -a 256 "$@"; }
fi

# --check verifies the vendored file against the manifest. It is offline: it
# only proves the checked-in blob still matches the hash that was reviewed
# when it was committed.
if [ "${1:-}" = "--check" ]; then
  if [ ! -f "$MANIFEST" ]; then
    echo "${MANIFEST} not found." >&2
    exit 1
  fi
  sha256 -c "$MANIFEST"
  exit 0
fi

VERSION="${1:-latest}"

if [ "$VERSION" = "latest" ]; then
  VERSION=$(npm view katex version)
fi

echo "Downloading katex@${VERSION}..."

INTEGRITY=$(npm view "katex@${VERSION}" dist.integrity)
case "$INTEGRITY" in
  sha512-*) ;;
  *)
    echo "Unexpected integrity format from npm: ${INTEGRITY}" >&2
    exit 1
    ;;
esac

URL="https://registry.npmjs.org/katex/-/katex-${VERSION}.tgz"
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

curl -sSL "$URL" -o "$tmp/katex.tgz"

# Verify the tarball before unpacking it, so a tampered archive is never
# extracted to disk.
ACTUAL="sha512-$(openssl dgst -sha512 -binary "$tmp/katex.tgz" | openssl base64 -A)"
if [ "$ACTUAL" != "$INTEGRITY" ]; then
  echo "Integrity mismatch for katex-${VERSION}.tgz:" >&2
  echo "  expected ${INTEGRITY}" >&2
  echo "  actual   ${ACTUAL}" >&2
  exit 1
fi

tar -xzf "$tmp/katex.tgz" -C "$tmp" package/dist/"$DIST"
cp "$tmp/package/dist/$DIST" "$DIST"

{
  echo "# katex ${VERSION}"
  echo "# tarball ${INTEGRITY}"
  sha256 "$DIST"
} > "$MANIFEST"

echo "Updated ${DIST} to version ${VERSION}."
echo
cat "$MANIFEST"
