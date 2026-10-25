#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
STAMP="$(date +%Y%m%d-%H%M%S)"
OUT="${1:-./backups/gate-${STAMP}}"
mkdir -p "$OUT"

docker compose exec -T postgres pg_dump -U "${POSTGRES_USER:-gate}" "${POSTGRES_DB:-gate}" > "$OUT/gate.sql"
cp -a traefik/dynamic "$OUT/traefik-dynamic" 2>/dev/null || true
cp .env "$OUT/env.snapshot" 2>/dev/null || true
echo "Backup written to $OUT"