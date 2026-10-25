# Download OWASP CRS into coraza/rules for Coraza WASM
$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $PSScriptRoot
$CrsVer = if ($env:CRS_VER) { $env:CRS_VER } else { "v4.14.0" }
$Tmp = Join-Path $env:TEMP ("the-gate-crs-" + [guid]::NewGuid().ToString())
New-Item -ItemType Directory -Path $Tmp | Out-Null

try {
  $Url = "https://github.com/coreruleset/coreruleset/archive/refs/tags/$CrsVer.tar.gz"
  $Tar = Join-Path $Tmp "crs.tar.gz"
  Write-Host "Fetching OWASP CRS $CrsVer..."
  Invoke-WebRequest -Uri $Url -OutFile $Tar
  tar -xz -C $Tmp -f $Tar
  $Src = Get-ChildItem -Directory $Tmp | Where-Object { $_.Name -like "coreruleset-*" } | Select-Object -First 1
  if (-not $Src) { throw "CRS extract failed" }

  $Dest = Join-Path $Root "coraza"
  $SetupExample = Join-Path $Src.FullName "crs-setup.conf.example"
  if (Test-Path $SetupExample) {
    Copy-Item $SetupExample (Join-Path $Dest "crs-setup.conf") -Force
  }

  $RulesDest = Join-Path $Dest "rules"
  if (Test-Path $RulesDest) { Remove-Item $RulesDest -Recurse -Force }
  New-Item -ItemType Directory -Path $RulesDest | Out-Null
  Copy-Item (Join-Path $Src.FullName "rules\*.conf") $RulesDest -Force

  $Count = (Get-ChildItem $RulesDest -Filter *.conf).Count
  Write-Host "CRS installed under coraza/rules ($Count files)"
}
finally {
  Remove-Item $Tmp -Recurse -Force -ErrorAction SilentlyContinue
}