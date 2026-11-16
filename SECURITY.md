# Security Policy

## Supported versions

The Gate is released under MIT. Use the latest `main` / tagged `v*` release for production. Security fixes land on `main` first.

## Reporting a vulnerability

Please report security issues privately to **Over The Cloud**:

- Web: [https://overthecloud.it](https://overthecloud.it)
- Prefer coordinated disclosure; do not open a public GitHub issue for unpatched vulnerabilities.

Include: affected version/commit, reproduction steps, impact, and (if possible) a suggested fix.

## Operational hardening (operators)

1. Never commit a real `.env`. Use `.env.example` and generate strong secrets.
2. Do not expose Postgres or `gate-api` on the public Internet. Publish only the admin UI port (`GATE_PORT`, default `9444`) and Traefik `80`/`443` as needed.
3. Prefer VPN / allowlists for the admin UI in production.
4. Keep protection feeds current (OWASP CRS + CrowdSec Hub) via Settings or `GATE_AUTO_UPDATE_*`.
5. Mount `docker.sock` (`docker-compose.hub-update.yml`) only on hosts you fully trust.
6. Rotate `BOUNCER_KEY_TRAEFIK`, `GATE_SESSION_SECRET`, and DB/admin passwords if leaked.
7. Runtime Traefik files under `traefik/dynamic/app-*.yml` and rendered `middlewares.yml` may contain secrets — they are gitignored; do not publish them.

## Third-party components

Traefik, CrowdSec, Coraza, OWASP CRS, and plugins retain their own licenses and security processes — see [`NOTICE`](./NOTICE).
