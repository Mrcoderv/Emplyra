# Emplyra one-click launcher (Windows — PowerShell).
#
# Starts all three pieces:
#   1. Database  — Postgres 16 via Docker (backend/docker-compose.yml)
#   2. Backend   — Go API server (backend/, reads backend/.env), port 8080
#   3. Frontend  — Next.js app (frontend/), port 3000
#
# Press Enter (or Ctrl+C) in this window to stop the backend + frontend and
# stop (not remove) the database container. Data persists in the named volume.
#
# Run directly:
#   powershell -ExecutionPolicy Bypass -File .\run.ps1
# ...or double-click run.bat

$ErrorActionPreference = 'Stop'

$Root    = if ($PSScriptRoot) { $PSScriptRoot } else { (Get-Location).Path }
$Backend = Join-Path $Root 'backend'
$Frontend = Join-Path $Root 'frontend'
$RunDir  = Join-Path $Root '.run'
$DbContainer = 'emplyra-db'
$DbUser      = 'emplyra'
$DbName      = 'emplyra'

function Step($Message) { Write-Host "==> $Message" -ForegroundColor Cyan }
function Info($Message) { Write-Host "[ok] $Message" -ForegroundColor Green }
function Warn($Message) { Write-Host "[!] $Message" -ForegroundColor Yellow }
function Fail($Message) { Write-Host "[x] $Message" -ForegroundColor Red; exit 1 }

foreach ($cmd in @('docker', 'go', 'node', 'pnpm')) {
  if (-not (Get-Command $cmd -ErrorAction SilentlyContinue)) {
    Fail "[$cmd] not found on PATH. Install it and retry."
  }
}

New-Item -ItemType Directory -Force -Path $RunDir | Out-Null

# --- 1. Database ----------------------------------------------------------
Step 'Starting database (Postgres 16 via Docker)...'
docker info *> $null
if ($LASTEXITCODE -ne 0) { Fail 'Docker is not running. Start Docker Desktop and retry.' }

$composeFile = Join-Path $Backend 'docker-compose.yml'
docker compose -f $composeFile --project-directory $Backend up -d db
if ($LASTEXITCODE -ne 0) { Fail 'Could not start the database container. Is port 5432 already in use?' }

Step 'Waiting for Postgres to accept connections...'
$ready = $false
for ($i = 0; $i -lt 60; $i++) {
  docker exec $DbContainer pg_isready -U $DbUser -d $DbName *> $null
  if ($LASTEXITCODE -eq 0) { $ready = $true; break }
  Start-Sleep -Seconds 1
}
if (-not $ready) { Fail 'Postgres did not become ready in time.' }
Info 'Database ready.'

# --- 2. Backend -----------------------------------------------------------
Step 'Building backend...'
Push-Location $Backend
try { go build -o (Join-Path $RunDir 'server.exe') ./cmd/server } finally { Pop-Location }
if ($LASTEXITCODE -ne 0) { Fail 'Backend build failed.' }

Step 'Starting backend...'
$backend = Start-Process -FilePath (Join-Path $RunDir 'server.exe') `
  -WorkingDirectory $Backend -PassThru `
  -RedirectStandardOutput (Join-Path $RunDir 'backend.log') `
  -RedirectStandardError (Join-Path $RunDir 'backend.err.log') `
  -WindowStyle Hidden

# --- 3. Frontend ----------------------------------------------------------
Step 'Installing frontend dependencies (if needed)...'
if (-not (Test-Path (Join-Path $Frontend 'node_modules'))) {
  Push-Location $Frontend
  try { pnpm install } finally { Pop-Location }
} else {
  Info 'node_modules already present, skipping install.'
}

Step 'Starting frontend...'
$pnpm = (Get-Command pnpm).Source
$frontend = Start-Process -FilePath $pnpm -ArgumentList 'dev' `
  -WorkingDirectory $Frontend -PassThru `
  -RedirectStandardOutput (Join-Path $RunDir 'frontend.log') `
  -RedirectStandardError (Join-Path $RunDir 'frontend.err.log') `
  -WindowStyle Hidden

# --- Readiness ------------------------------------------------------------
function Test-Port($Port) {
  $client = [System.Net.Sockets.TcpClient]::new()
  try {
    $task = $client.ConnectAsync('127.0.0.1', $Port)
    if ($task.Wait(1000)) { return $true }
    return $false
  } catch { return $false }
  finally { $client.Dispose() }
}
Step 'Waiting for services to come up...'
$ready = $false
for ($i = 0; $i -lt 60; $i++) {
  if ((Test-Port 8080) -and (Test-Port 3000)) { $ready = $true; break }
  Start-Sleep -Seconds 1
}
if (-not $ready) {
  Fail 'Backend and/or frontend did not come up in time. Check the logs in .run\*.log'
}

Write-Host ''
Info '---------------- Emplyra is running ----------------'
Info 'Frontend  : http://localhost:3000'
Info 'Backend   : http://localhost:8080'
Info 'API       : http://localhost:8080/api/v1'
Info "Health    : http://localhost:8080/healthz"
Info ("Logs      : " + (Join-Path $RunDir 'backend.log'))
Info 'Admin user: admin@emplyra.local / ChangeMe123!'
Write-Host ''
Warn 'Press Enter (or Ctrl+C) to stop everything.'
Write-Host ''

try {
  Read-Host | Out-Null   # wait here
} finally {
  if ($frontend) { taskkill /F /T /PID $frontend.Id *> $null }
  if ($backend)  { Stop-Process -Id $backend.Id  -Force -ErrorAction SilentlyContinue }
  docker compose -f $composeFile --project-directory $Backend stop db | Out-Null
  Write-Host ''
  Info 'All stopped. Database data is preserved in the docker volume (emplyra_pgdata).'
}