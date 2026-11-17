# The Gate

**Self-hosted edge control plane** — reverse proxy + WAF + bot/IP intelligence, managed from a web UI.

| | |
|---|---|
| **Author** | Carlo Savino |
| **Company** | [Over The Cloud](https://overthecloud.it) |
| **License** | [MIT](./LICENSE) |
| **Security** | [SECURITY.md](./SECURITY.md) |

The Gate is **not** a binary fork of SafeLine. It orchestrates open components (Traefik, CrowdSec, Coraza/OWASP CRS) with an application-centric admin UI.

---

## What you get

- **Applications** — public host → upstream, WAF mode, rate limit, auth, geo-block, TLS
- **Traefik v3** — routing, Let’s Encrypt, middlewares
- **CrowdSec** — decisions + AppSec (with Hub updates)
- **Coraza + OWASP CRS** — WAF ruleset
- **Security rules** — Cloudflare-style conditions (per-app or global)
- **Antibot JS challenge**, **DNSBL**, **bad-behavior auto-ban** (CrowdSec)
- **WWW redirect**, custom request/response headers, WebSocket sticky cookie
- **Allow / Deny** IP groups, traffic view, decisions, audit, certificates UI
- **Bulk actions** on selected applications
- **i18n** EN/IT, panel themes, custom error pages

```text
Internet  →  Traefik :80 / :443
               ├─ CrowdSec (ban + AppSec) + Coraza CRS
               ├─ rate-limit / geo-block / custom rules / auth
               └─ your upstreams

Admin UI  →  https://<host>:9444   (gate-ui → gate-api → Postgres)
```

---

## Requirements

- Docker Engine + Compose v2 (Docker Desktop on Windows)
- Free ports: `80`, `443`, and admin UI (default `9444`)
- For public TLS: DNS A/AAAA pointing at the host
- Optional: Authentik for SSO; MaxMind GeoLite2 for country blocking

---

## Quick start

### Windows (PowerShell)

```powershell
git clone https://github.com/slash3rmast3r/the-gate.git
cd the-gate
.\scripts\bootstrap.ps1
```

Open **https://localhost:9444** (self-signed cert — accept the warning).  
Login credentials are in `.env` (`GATE_ADMIN_USER` / `GATE_ADMIN_PASSWORD`).

### Linux

```bash
git clone https://github.com/slash3rmast3r/the-gate.git
cd the-gate
chmod +x scripts/*.sh
./scripts/bootstrap.sh
```

Open **https://\<host\>:9444**. Credentials in `.env`.

### Manual (any OS)

```bash
cp .env.example .env   # PowerShell: Copy-Item .env.example .env
# Edit secrets: GATE_ADMIN_PASSWORD, GATE_SESSION_SECRET, BOUNCER_KEY_TRAEFIK, POSTGRES_PASSWORD, ACME_EMAIL
./scripts/fetch-crs.sh # or .\scripts\fetch-crs.ps1
docker compose up -d --build
```

Bootstrap creates `.env` (if missing), fetches OWASP CRS, builds and starts the stack.

**Smoke test:** Applications → Add → host with DNS/hosts entry → upstream `http://whoami:80` → `curl -vk https://<host>/`.

---

## Stack services

| Service | Role |
|---------|------|
| `traefik` | Edge proxy, ACME, plugins |
| `crowdsec` | Ban decisions + AppSec |
| `gate-api` | Control plane API + Traefik sync |
| `gate-ui` | Admin HTTPS UI |
| `postgres` | Control-plane DB |
| `gate-errors` | Custom error pages |
| `whoami` | Optional test upstream |

Useful commands:

```bash
docker compose ps
docker compose logs -f gate-api traefik crowdsec
docker compose up -d --build
docker compose down
```

---

## Configuration (`.env`)

Copy from [`.env.example`](./.env.example). Never commit a real `.env`.

| Variable | Purpose |
|----------|---------|
| `GATE_PORT` | Admin UI HTTPS port (default `9444`) |
| `GATE_ADMIN_USER` / `GATE_ADMIN_PASSWORD` | Local login (password re-synced from env on API start) |
| `GATE_SESSION_SECRET` | Session cookie secret |
| `GATE_AUTH_MODE` | `local` \| `oidc` \| `both` |
| `GATE_OIDC_*` | Authentik/OIDC for admin UI |
| `AUTHENTIK_FORWARD_URL` | Outpost URL for app SSO |
| `BOUNCER_KEY_TRAEFIK` | CrowdSec ↔ Traefik bouncer key |
| `POSTGRES_PASSWORD` | Database password |
| `ACME_EMAIL` | Let’s Encrypt contact |
| `GATE_AUTO_UPDATE_PROTECTIONS` | Default auto-update CRS/Hub |
| `GATE_AUTO_UPDATE_HOURS` | Auto-update interval (hours) |
| `GATE_GEOIP_DB` | Path to GeoLite2 Country DB inside API container |
| `SUBNET_PREFIX` / `TZ` | Docker network prefix / timezone |

---

## Backups

There are **two** backup paths:

### 1) UI — control-plane ZIP (download)

**Settings → Backup now** calls `POST /api/backup` and downloads  
`the-gate-backup-YYYYMMDD-HHMMSS.zip` to your browser.

Contents: applications, settings (redacted), IP groups, recent audit, Traefik dynamic YAML.  
Good for config export; **not** a full disaster-recovery image.

### 2) Script — full host backup (recommended for DR)

```bash
./scripts/backup.sh
# optional custom dir:
./scripts/backup.sh ./backups/my-snapshot
```

Default output: **`./backups/gate-<timestamp>/`** (gitignored), including:

| Path | Content |
|------|---------|
| `gate.sql` | Postgres dump |
| `traefik-dynamic/` | Generated Traefik app YAML |
| `traefik-letsencrypt/` | ACME / certs storage |
| `crowdsec/` | CrowdSec data/config |
| `coraza/` | CRS rules |
| `env.snapshot` | Copy of `.env` (sensitive!) |

On Windows PowerShell without bash:

```powershell
docker compose exec -T postgres pg_dump -U gate gate > backup.sql
```

Keep `./backups/` off public shares and out of git.

---

## Keeping The Gate updated

Updates are **layered**. Do not use `:latest` on Traefik/CrowdSec/Postgres in production.

| Layer | What | How |
|-------|------|-----|
| **A — Intelligence** | OWASP CRS + CrowdSec Hub | Settings UI and/or `GATE_AUTO_UPDATE_*` |
| **B — Images** | Traefik, CrowdSec, Postgres, nginx | Pin bump in `docker-compose.yml` + `pull` / `up -d` |
| **C — Plugins** | CrowdSec bouncer + Coraza WASM | Versions in `traefik/traefik.yml` |
| **D — Product** | gate-api / gate-ui | `git pull` + `compose build` or GHCR tags |

**CrowdSec Hub from UI** needs Docker socket (elevated trust):

```bash
docker compose -f docker-compose.yml -f docker-compose.hub-update.yml up -d
```

Leave this disabled on shared / untrusted hosts.

### “Break-free” upgrade guidance

Components are **coupled**. Safe vs risky:

| Change | Usually safe | Needs care |
|--------|--------------|------------|
| CRS / Hub only | Yes — independent of Traefik image | Test WAF false positives |
| Postgres `16.x` patch | Yes within major 16 | Backup first |
| nginx patch (gate-errors) | Yes | Low risk |
| CrowdSec `v1.6.x` patch | Usually yes with same Traefik plugin | Smoke AppSec + decisions |
| Traefik `v3.3.x` patch | Usually yes with **same** plugin pins | Smoke HTTPS + middlewares |
| Traefik minor (`v3.3` → `v3.4`) | — | Re-check plugin release notes; bump plugins if required |
| CrowdSec minor (`1.6` → `1.7`) | — | Check AppSec + bouncer plugin requirements |
| Traefik / CrowdSec **major** | — | Staging only; treat as a project upgrade |
| Plugin major (bouncer / Coraza) | — | Must match Traefik major; never bump alone blindly |

**Current pins (see compose / `traefik.yml`):**

- `traefik:v3.3` + plugins `crowdsec-bouncer-traefik-plugin` / `coraza-http-wasm-traefik`
- `crowdsecurity/crowdsec:v1.6.9`
- `postgres:16-alpine`, `nginx:1.27-alpine`
- Default CRS `v4.14.0`

**Practical cadence:** weekly intelligence (A); monthly image/plugin review (B+C); rebuild product when you change code (D).  
After any image bump: backup → deploy → check `docker compose ps`, admin login, one HTTPS app, WAF/rules, certificates page.

---

## CI / CD and Renovate

| Piece | Status | Path |
|-------|--------|------|
| **CI pipeline** | Present | [`.github/workflows/ci.yml`](./.github/workflows/ci.yml) — Gitleaks, Go vet/build/test, frontend build, Compose validate, Docker build |
| **Release** | Present | [`.github/workflows/release.yml`](./.github/workflows/release.yml) — tag `v*` → GHCR (`the-gate-api`, `the-gate-ui`) |
| **Renovate** | Present | [`renovate.json`](./renovate.json) — Compose/Dockerfile, Go/npm, Actions, Traefik plugin pins, CRS pin |

Enable the **Renovate GitHub App** on the repository so dependency PRs are opened automatically. Majors for Docker/plugins require approval (no blind `:latest`).

---

## Authentik SSO (two separate integrations)

The Gate supports Authentik in **two independent ways**. You can use one, both, or neither.

### A) Admin UI login (OIDC)

Protects **only** the Gate panel (`https://<host>:9444`).

1. In Authentik create an **OIDC / OAuth2** application.
2. Redirect URI: `https://<gate-host>:9444/api/auth/oidc/callback`
3. In `.env`:

```env
GATE_AUTH_MODE=both
GATE_OIDC_ISSUER=https://auth.example.com/application/o/the-gate/
GATE_OIDC_CLIENT_ID=...
GATE_OIDC_CLIENT_SECRET=...
GATE_OIDC_REDIRECT_URL=https://gate.example.com:9444/api/auth/oidc/callback
```

4. `docker compose up -d` and use **Sign in with Authentik / OIDC** on the login page.

`local` = password only · `oidc` = Authentik only · `both` = either.

### B) Per-application SSO (Traefik forward-auth) — already built-in

Protects **each selected Application** (public host → upstream), not the admin UI.

| Piece | Where |
|-------|--------|
| Outpost URL (global) | `.env` → `AUTHENTIK_FORWARD_URL` |
| SSO on/off per app | UI → Applications → **Access auth** → **Authentik SSO** |

Default:

```env
AUTHENTIK_FORWARD_URL=http://authentik-server:9000/outpost.goauthentik.io/auth/traefik
```

**Setup checklist**

1. Authentik (and its Traefik/Proxy **Outpost**) must be reachable from The Gate Docker network.
2. In Authentik: Application + Provider (forward-auth / proxy) for the **public hostnames** of your Gate apps (e.g. `app.example.com`).
3. Attach that Application to the Outpost; set access policies (groups) in Authentik.
4. In The Gate: edit the app → Access auth → **Authentik SSO** → Save (Traefik sync attaches `authentik-forward`).

Flow: browser → Traefik → Authentik outpost → if allowed → upstream.  
Headers such as `X-authentik-username` / `groups` / `email` are forwarded.

**Notes**

- Auth modes per app: `none` | `basic` | `authentik` (mutually exclusive on that app).
- One shared `AUTHENTIK_FORWARD_URL` for all apps; policies/domains live in Authentik.
- Admin OIDC and per-app forward-auth do **not** replace each other.

---

## Project layout

```text
docker-compose.yml            # full stack
docker-compose.hub-update.yml # optional docker.sock for Hub UI updates
backend/                      # gate-api (Go)
frontend/                     # gate-ui (React)
traefik/                      # static + dynamic (API-managed)
coraza/                       # OWASP CRS
crowdsec/                     # acquis / AppSec wiring
errorpages/                   # custom error HTML
scripts/                      # bootstrap, fetch-crs, backup
.github/workflows/            # CI + release
renovate.json                 # dependency updates
```

---

## Operational notes

1. Without a full CRS fetch, WAF mode `full` is nearly empty — run bootstrap / `fetch-crs` / Settings update.
2. Do not publish Postgres or `gate-api` publicly; restrict admin UI (firewall/VPN).
3. Cloudflare orange-cloud can break HTTP-01 Let’s Encrypt — grey-cloud during issue or use DNS-01.
4. Host upstreams: `http://host.docker.internal:<port>` (configured on Traefik).
5. Geo-block needs GeoLite2 under `./data/geoip` (see `GATE_GEOIP_DB`).
6. Never commit secrets; see [`SECURITY.md`](./SECURITY.md).

---

## License & copyright

```text
Copyright (c) 2026 Carlo Savino — Over The Cloud
https://overthecloud.it
```

MIT — [`LICENSE`](./LICENSE). Third-party components keep their own licenses — [`NOTICE`](./NOTICE).

**Over The Cloud** — [https://overthecloud.it](https://overthecloud.it)
