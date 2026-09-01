$ErrorActionPreference = 'Stop'

$Root = if ($PSScriptRoot) { $PSScriptRoot } else { (Get-Location).Path }
$Runner = Join-Path $Root 'run.ps1'

if (-not (Test-Path -LiteralPath $Runner)) {
  Write-Error "Expected PowerShell runner was not found: $Runner"
  exit 1
}

& $Runner @args
exit $LASTEXITCODE
