# The Gate

**Author:** Carlo Savino · **Company:** [Over The Cloud](https://overthecloud.it) · **License:** [MIT](./LICENSE) · **Web:** [overthecloud.it](https://overthecloud.it)

CI: GitHub Actions (secret scan, Go, frontend build, Compose validate, Docker image build). Release tags `v*` publish images to GHCR.

WAF / reverse-proxy **self-hosted** indipendente, pensato per girare interamente in **Docker Compose**.

Stack: **Traefik v3** + **CrowdSec** (ban + AppSec) + **Coraza / OWASP CRS**, con UI di gestione applicazioni (stile SafeLine) — senza registry Chaitin e senza limiti sul numero di siti.

```text
Internet  →  Traefik :80 / :443
               ├─ CrowdSec (ban + AppSec) + Coraza CRS
               ├─ rate-limit / Basic Auth / Authentik forward-auth
               └─ upstream (i tuoi servizi)

Admin UI  →  https://<host>:9444  (gate-ui)
               └─ gate-api → Postgres + scrive YAML dinamici Traefik
```

---

## Requisiti (comuni)

| Requisito | Note |
|-----------|------|
| **Docker Engine + Compose v2** | Su Windows: Docker Desktop. Su Linux: Docker CE + plugin `compose` |
| Porte libere | `80`, `443` (traffico + Let’s Encrypt), `9444` (UI admin, modificabile) |
| DNS (produzione) | Record A/AAAA verso l’host per TLS pubblici |
| (Opzionale) Authentik | Solo se usi SSO admin e/o forward-auth sulle app |

---

## Setup rapido con Docker

### Windows (PowerShell + Docker Desktop)

1. Avvia **Docker Desktop** e attendi che il motore sia Ready.
2. Clona (o copia) il progetto e apri PowerShell nella cartella del repo:

```powershell
git clone https://github.com/slash3rmast3r/the-gate.git
cd the-gate
.\scripts\bootstrap.ps1
```

Lo script:

- crea `.env` con secret generati (se non esiste)
- scarica OWASP CRS (consigliato)
- esegue `docker compose pull`, `build`, `up -d`

3. Apri **https://localhost:9444** (certificato self-signed: accetta l’avviso).
4. Login: `GATE_ADMIN_USER` / `GATE_ADMIN_PASSWORD` nel file `.env`.

**Manuale (senza bootstrap):**

```powershell
cd the-gate
Copy-Item .env.example .env
# modifica .env (password, ACME_EMAIL, BOUNCER_KEY_TRAEFIK, …)
.\scripts\fetch-crs.ps1
docker compose up -d --build
```

**Comandi utili (Windows):**

```powershell
docker compose ps
docker compose logs -f gate-api traefik crowdsec
docker compose down
docker compose up -d --build
```

> Backup: `scripts\backup.sh` è pensato per bash (WSL / Git Bash). Su PowerShell puro:  
> `docker compose exec -T postgres pg_dump -U gate gate > backup.sql`

---

### Linux (server / VPS)

1. Installa Docker se manca (es. [docs ufficiali](https://docs.docker.com/engine/install/)):

```bash
docker --version
docker compose version
```

2. Clona il progetto ed esegui il bootstrap:

```bash
git clone https://github.com/slash3rmast3r/the-gate.git
cd the-gate
chmod +x scripts/*.sh
./scripts/bootstrap.sh
```

Lo script crea `.env`, fetch CRS, build e `docker compose up -d`.

3. Apri **https://\<IP-o-hostname\>:9444** (self-signed). Credenziali in `.env`.

**Manuale (senza bootstrap):**

```bash
cd the-gate
cp .env.example .env
nano .env   # password, ACME_EMAIL, BOUNCER_KEY_TRAEFIK, …
./scripts/fetch-crs.sh
docker compose up -d --build
```

**Comandi utili (Linux):**

```bash
docker compose ps
docker compose logs -f gate-api traefik crowdsec
docker compose down
docker compose up -d --build
./scripts/backup.sh
```

---

## Dopo il primo avvio (smoke test)

Nella UI → **Applications** → **Add application**:

| Campo | Valore di test |
|--------|----------------|
| Name | whoami |
| Public host | un hostname con DNS verso questo host (o hosts file) |
| Upstream URL | `http://whoami:80` |
| WAF mode | `crowdsec-only` all’inizio, poi `full` |
| Rate limit | soft |
| Auth | none (poi Basic / Authentik se serve) |

Poi: `curl -vk https://<host>/`

Servizi avviati da Compose: `postgres`, `crowdsec`, `traefik`, `gate-api`, `gate-ui`, `gate-errors`, `whoami`.

---

## Feature del prodotto (checklist verifica)

| Feature discussa | Stato | Dove |
|------------------|-------|------|
| Gestione Applications stile SafeLine | OK | UI Applications + sync Traefik |
| WAF CRS + CrowdSec / modi per app | OK | `full` / `light` / `crowdsec-only` / `off` |
| Rate limit + anti-DoS | OK | soft / strict / **dos** / off |
| Ban bot / IP (decisions) | OK | CrowdSec + UI Decisions |
| Anti-bruteforce login admin | OK | 5 fail → lock 15 min |
| Liste protezioni on-demand da UI | OK | Settings → CRS / CrowdSec Hub |
| Auto-update feed | OK | `GATE_AUTO_UPDATE_PROTECTIONS` |
| Basic Auth per app | OK | Auth mode → Basic |
| SSO Authentik sulle app | OK | Auth mode → Authentik + `AUTHENTIK_FORWARD_URL` |
| SSO Authentik sulla UI Gate | OK | `GATE_AUTH_MODE` + `GATE_OIDC_*` |
| Pagine errore custom | OK | `errorpages/` + flag per app |
| Logs access/error per app | OK | pulsante **Logs** → `/apps/:id/logs` |
| Audit control-plane | OK | UI Audit |
| Export JSON app | OK | Settings |
| Docker Compose Win + Linux | OK | `scripts/bootstrap.ps1` / `.sh` |
| Licenza MIT + copyright OTC | OK | `LICENSE`, `NOTICE`, UI, API |

> Non in scope (esplicitamente non clonate da SafeLine): editor WYSIWYG pagine errore in UI, captcha JS proprietario, timeline “Attacks” stile detector CE.

---

## Variabili `.env` essenziali

Copia da `.env.example` (il bootstrap lo fa da solo).

| Variabile | Ruolo |
|-----------|--------|
| `GATE_PORT` | Porta UI HTTPS (default `9444`) |
| `GATE_ADMIN_USER` / `GATE_ADMIN_PASSWORD` | Login locale (password riallineata a ogni avvio API) |
| `GATE_SESSION_SECRET` | Cookie di sessione |
| `GATE_AUTH_MODE` | `local` \| `oidc` \| `both` |
| `GATE_OIDC_ISSUER` / `CLIENT_ID` / `CLIENT_SECRET` / `REDIRECT_URL` | SSO admin (Authentik OIDC) |
| `AUTHENTIK_FORWARD_URL` | Outpost per SSO sulle app |
| `BOUNCER_KEY_TRAEFIK` | Chiave CrowdSec ↔ Traefik |
| `POSTGRES_PASSWORD` | Database control plane |
| `ACME_EMAIL` | Contatto Let’s Encrypt |
| `GATE_AUTO_UPDATE_PROTECTIONS` | `true` = aggiorna CRS/Hub in automatico |
| `SUBNET_PREFIX` | Rete Docker (default `172.23.223`) |
| `TZ` | Timezone |

Non lasciare valori `CHANGE_ME_*` in produzione.

---

## Authentik (opzionale)

**Login UI The Gate (OIDC)** — in Authentik crea un’Application OIDC con redirect  
`https://<gate-host>:9444/api/auth/oidc/callback`, poi:

```env
GATE_AUTH_MODE=both
GATE_OIDC_ISSUER=https://auth.example.com/application/o/the-gate/
GATE_OIDC_CLIENT_ID=...
GATE_OIDC_CLIENT_SECRET=...
GATE_OIDC_REDIRECT_URL=https://gate.example.com:9444/api/auth/oidc/callback
```

Riavvia: `docker compose up -d`.

**App dietro The Gate:** Outpost Traefik in Authentik → imposta `AUTHENTIK_FORWARD_URL` → in UI sull’app scegli **Authentik SSO**.

> Authentik deve essere raggiungibile dalla rete Docker di The Gate (stessa rete o rete esterna collegata).

---

## Layout progetto

```
docker-compose.yml   # stack completo
backend/             # gate-api (Go)
frontend/            # gate-ui (React)
traefik/             # config + dynamic/ (gestita dall’API)
coraza/              # OWASP CRS
crowdsec/            # acquis log + AppSec
errorpages/          # pagine errore personalizzabili
scripts/             # bootstrap Win/Linux, fetch-crs, backup
```

---

## Note operative

1. Senza CRS completo (`bootstrap` / `fetch-crs` / pulsante Settings) il modo WAF `full` è quasi vuoto.
2. **CrowdSec Hub da UI** richiede il mount di `docker.sock` (non attivo di default). Abilitalo solo se serve:
   `docker compose -f docker-compose.yml -f docker-compose.hub-update.yml up -d`
3. Non pubblicare Postgres né `gate-api` su Internet; limita l’accesso a `:9444` (firewall/VPN).
4. Cloudflare “arancione” può interferire con HTTP-01 Let’s Encrypt: grey-cloud in fase di emissione o DNS-01.
5. Upstream sull’host fisico: `http://host.docker.internal:<porta>` (già configurato su Traefik).
6. Non committare `.env` — solo `.env.example` (placeholder). Vedi [`SECURITY.md`](./SECURITY.md).

The Gate **non** è un fork binario di SafeLine: è uno stack Docker open standard con UX di gestione simile.

---

## License & copyright

```text
Copyright (c) 2026 Carlo Savino — Over The Cloud
https://overthecloud.it
```

The Gate is released under the **MIT License** — see [`LICENSE`](./LICENSE).  
Third-party components (Traefik, CrowdSec, Coraza, OWASP CRS, …) retain their own licenses — see [`NOTICE`](./NOTICE).

**Over The Cloud** — [https://overthecloud.it](https://overthecloud.it)
