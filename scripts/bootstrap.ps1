$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $PSScriptRoot
Set-Location $Root

function New-Hex([int]$Bytes) {
  $buf = New-Object byte[] $Bytes
  [System.Security.Cryptography.RandomNumberGenerator]::Create().GetBytes($buf)
  -join ($buf | ForEach-Object { $_.ToString("x2") })
}

if (-not (Test-Path .env)) {
  Copy-Item .env.example .env
  $key = New-Hex 32
  $secret = New-Hex 32
  $pass = New-Hex 16
  (Get-Content .env) `
    -replace "GATE_ADMIN_PASSWORD=.*", "GATE_ADMIN_PASSWORD=$pass" `
    -replace "GATE_SESSION_SECRET=.*", "GATE_SESSION_SECRET=$secret" `
    -replace "BOUNCER_KEY_TRAEFIK=.*", "BOUNCER_KEY_TRAEFIK=$key" `
    -replace "POSTGRES_PASSWORD=.*", "POSTGRES_PASSWORD=$pass" |
    Set-Content .env -Encoding UTF8
  Write-Host "Created .env with generated secrets. Edit ACME_EMAIL / DOMAIN before going live."
}

New-Item -ItemType Directory -Force -Path logs\traefik | Out-Null

if (-not (Test-Path "coraza\rules\REQUEST-920-PROTOCOL-ENFORCEMENT.conf")) {
  Write-Host "Fetching OWASP CRS (recommended)..."
  try { & "$PSScriptRoot\fetch-crs.ps1" } catch { Write-Warning "CRS fetch skipped — starter rules only." }
}

docker compose pull
docker compose build
docker compose up -d
Write-Host ""
Write-Host "The Gate UI: https://localhost:9444"
Write-Host "Login: see GATE_ADMIN_USER / GATE_ADMIN_PASSWORD in .env"
Write-Host "Add an Application pointing at http://whoami:80 for a smoke test."