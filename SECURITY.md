# Security Policy

## Reporting a vulnerability

Please report security issues privately to **Over The Cloud** via [overthecloud.it](https://overthecloud.it) (contact / abuse channels on the site). Do not open public GitHub issues for unfixed vulnerabilities.

## Public repository hygiene

This repository is intended to be public. It must **never** contain:

- Real `.env` files or production secrets
- API keys, OIDC client secrets, bouncer keys, TLS private keys
- Live `acme.json`, database dumps, or access logs with client IPs

Use `.env.example` as a template only. Bootstrap scripts generate random secrets locally.

## Deployment notes

- Keep `GATE_PORT` / admin UI off the public internet when possible (VPN / allowlist).
- Do not expose Postgres or `gate-api` directly.
- Mounting `/var/run/docker.sock` into `gate-api` is **optional** and powerful — only enable via `docker-compose.hub-update.yml` if you need CrowdSec Hub updates from the UI.
- Rotate `GATE_ADMIN_PASSWORD`, `GATE_SESSION_SECRET`, `BOUNCER_KEY_TRAEFIK`, and `POSTGRES_PASSWORD` if they ever leak.
