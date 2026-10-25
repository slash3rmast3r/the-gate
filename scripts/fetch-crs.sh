#!/usr/bin/env bash
# Download OWASP CRS into coraza/rules for Coraza WASM
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

CRS_VER="${CRS_VER:-v4.14.0}"
echo "Fetching OWASP CRS ${CRS_VER}..."

curl -fsSL "https://github.com/coreruleset/coreruleset/archive/refs/tags/${CRS_VER}.tar.gz" \
  | tar -xz -C "$TMP"

SRC="$(echo "$TMP"/coreruleset-* )"
DEST="$ROOT/coraza"

if [[ -f "$SRC/crs-setup.conf.example" ]]; then
  cp "$SRC/crs-setup.conf.example" "$DEST/crs-setup.conf"
fi

rm -rf "$DEST/rules"
mkdir -p "$DEST/rules"
cp -a "$SRC/rules/"*.conf "$DEST/rules/"

echo "CRS installed under coraza/rules ($(ls "$DEST/rules" | wc -l) files)"
echo "Tip: start with DetectionOnly if you get too many 403:"
echo "  set SecRuleEngine DetectionOnly in traefik/dynamic/middlewares.yml.template for coraza-crs"