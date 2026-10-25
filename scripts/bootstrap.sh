#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

if [[ ! -f .env ]]; then
  cp .env.example .env
  KEY="$(openssl rand -hex 32)"
  SECRET="$(openssl rand -hex 32)"
  PASS="$(openssl rand -hex 16)"
  sed -i.bak "s/GATE_ADMIN_PASSWORD=.*/GATE_ADMIN_PASSWORD=${PASS}/" .env
  sed -i.bak "s/GATE_SESSION_SECRET=.*/GATE_SESSION_SECRET=${SECRET}/" .env
  sed -i.bak "s/BOUNCER_KEY_TRAEFIK=.*/BOUNCER_KEY_TRAEFIK=${KEY}/" .env
  sed -i.bak "s/POSTGRES_PASSWORD=.*/POSTGRES_PASSWORD=${PASS}/" .env
  rm -f .env.bak
  echo "Created .env with generated secrets. Edit ACME_EMAIL / DOMAIN before going live."
fi

mkdir -p logs/traefik
chmod +x scripts/*.sh 2>/dev/null || true

if [[ ! -f coraza/rules/REQUEST-920-PROTOCOL-ENFORCEMENT.conf ]]; then
  echo "Fetching OWASP CRS (recommended)..."
  ./scripts/fetch-crs.sh || echo "CRS fetch skipped — starter rules only."
fi

docker compose pull
docker compose build
docker compose up -d
echo
echo "The Gate UI: https://$(hostname -f 2>/dev/null || echo localhost):${GATE_PORT:-9444}"
echo "Login: see GATE_ADMIN_USER / GATE_ADMIN_PASSWORD in .env"
echo "Add an Application pointing at http://whoami:80 for a smoke test."