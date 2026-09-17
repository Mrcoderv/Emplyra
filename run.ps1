# Emplyra one-click launcher (Windows — PowerShell).
#
# Starts all three pieces:
#   1. Database  — local PostgreSQL 16
#   2. Backend   — Go API server (backend/, reads backend/.env), port 8080
#   3. Frontend  — Next.js app (frontend/), port 3000
#
# Press Enter (or Ctrl+C) in this window to stop the backend + frontend.
#
# Run directly:
#   powershell -ExecutionPolicy Bypass -File .\run.ps1
# ...or double-click run.bat

$ErrorActionPreference = 'Stop'

$Root    = if ($PSScriptRoot) { $PSScriptRoot } else { (Get-Location).Path }
$Backend = Join-Path $Root 'backend'
$Frontend = Join-Path $Root 'frontend'
$RunDir  = Join-Path $Root '.run'

function Step($Message) { Write-Host "==> $Message" -ForegroundColor Cyan }
function Info($Message) { Write-Host "[ok] $Message" -ForegroundColor Green }
function Warn($Message) { Write-Host "[!] $Message" -ForegroundColor Yellow }
function Fail($Message) { Write-Host "[x] $Message" -ForegroundColor Red; exit 1 }

foreach ($cmd in @('go', 'node', 'pnpm', 'psql')) {
  if (-not (Get-Command $cmd -ErrorAction SilentlyContinue)) {
    Fail "[$cmd] not found on PATH. Install it and retry."
  }
}

New-Item -ItemType Directory -Force -Path $RunDir | Out-Null

# --- 1. Database ----------------------------------------------------------
$DbHost = if ($env:DB_HOST) { $env:DB_HOST } else { 'localhost' }
$DbPort = if ($env:DB_PORT) { $env:DB_PORT } else { '5432' }
Step "Checking local Postgres on ${DbHost}:${DbPort}..."
& psql -h $DbHost -p $DbPort -U emplyra -d emplyra -c '\q' 2>$null
if ($LASTEXITCODE -ne 0) {
  Fail "Postgres is not accepting connections on ${DbHost}:${DbPort}. Is it running?"
}
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
  Write-Host ''
  Info 'All stopped.'
}
