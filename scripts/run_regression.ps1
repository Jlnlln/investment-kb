$ErrorActionPreference = "Stop"
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$OutputEncoding = [System.Text.Encoding]::UTF8

$Root = Split-Path -Parent $PSScriptRoot
$OutputRoot = Join-Path $Root "testdata\output"
$Vault = Join-Path $OutputRoot "vault"
$Config = Join-Path $OutputRoot "regression_config.yaml"
$DataDir = Join-Path $Root "data"
$IdState = Join-Path $DataDir "id_state.json"
$ImportHashes = Join-Path $DataDir "import_hashes.json"
$Utf8NoBom = New-Object System.Text.UTF8Encoding($false)

function Convert-ToYamlPath {
    param([Parameter(Mandatory = $true)][string]$Path)
    return $Path.Replace("\", "/")
}

function Write-NoBomText {
    param(
        [Parameter(Mandatory = $true)][string]$Path,
        [Parameter(Mandatory = $true)][string]$Value
    )
    $parent = Split-Path -Parent $Path
    if ($parent -and !(Test-Path $parent)) {
        New-Item -ItemType Directory -Force -Path $parent | Out-Null
    }
    [System.IO.File]::WriteAllText($Path, $Value, $Utf8NoBom)
}

function Invoke-KB {
    param(
        [Parameter(ValueFromRemainingArguments = $true)]
        [string[]]$KbArgs
    )

    Push-Location $Root
    try {
        Write-Host ""
        Write-Host "go run ./cmd/kb $($KbArgs -join ' ')"
        & go run ./cmd/kb @KbArgs
        if ($LASTEXITCODE -ne 0) {
            throw "kb command failed: go run ./cmd/kb $($KbArgs -join ' ')"
        }
    } finally {
        Pop-Location
    }
}

Write-Host "=== investment-kb regression ==="
Write-Host "workspace: $Root"

if (Test-Path $OutputRoot) {
    Remove-Item -LiteralPath $OutputRoot -Recurse -Force
}
New-Item -ItemType Directory -Force -Path $Vault | Out-Null
New-Item -ItemType Directory -Force -Path (Join-Path $Vault "raw") | Out-Null
New-Item -ItemType Directory -Force -Path (Join-Path $Vault "qa") | Out-Null
New-Item -ItemType Directory -Force -Path (Join-Path $Vault "rules") | Out-Null
New-Item -ItemType Directory -Force -Path (Join-Path $Vault "rules\candidate_rules") | Out-Null
New-Item -ItemType Directory -Force -Path (Join-Path $Vault "rules\validation_cards") | Out-Null
New-Item -ItemType Directory -Force -Path (Join-Path $Vault "knowledge\macro_cards") | Out-Null
New-Item -ItemType Directory -Force -Path (Join-Path $Vault "observations\market_cards") | Out-Null
New-Item -ItemType Directory -Force -Path (Join-Path $Vault "templates") | Out-Null
New-Item -ItemType Directory -Force -Path $DataDir | Out-Null

if (Test-Path $IdState) {
    Remove-Item -LiteralPath $IdState -Force
}
if (Test-Path $ImportHashes) {
    Remove-Item -LiteralPath $ImportHashes -Force
}
Write-NoBomText -Path $IdState -Value "{}"
Write-NoBomText -Path $ImportHashes -Value "{}"

$vaultYaml = Convert-ToYamlPath -Path $Vault
$configYaml = @"
obsidian_vault_path: "$vaultYaml"

files:
  raw_input_inbox_dir: "inbox/qa"
  raw_material_dir: "raw"
  raw_material_index: "raw/index.md"
  qa_dir: "qa"
  qa_index: "qa/index.md"
  candidate_rule_dir: "rules/candidate_rules"
  candidate_rule_index: "rules/candidate_rules/index.md"
  market_case: "rules/market_cases.md"
  validation_card_template: "templates/validation_card_template.md"
  validation_card_dir: "rules/validation_cards"
  macro_knowledge_dir: "knowledge/macro_cards"
  macro_knowledge_index: "knowledge/macro_cards/index.md"
  market_observation_dir: "observations/market_cards"
  market_observation_index: "observations/market_cards/index.md"

ai:
  provider: "custom"
  model: "glm-5.1"
  base_url: "https://api.z.ai/api/anthropic"
  api_key_env: "ANTHROPIC_AUTH_TOKEN"
  timeout_seconds: 300
  temperature: 0

timezone: "Asia/Beijing"
"@
Write-NoBomText -Path $Config -Value $configYaml

Write-Host ""
Write-Host "=== generated regression_config.yaml ==="
Get-Content -LiteralPath $Config -Encoding UTF8 | ForEach-Object { Write-Host $_ }
Write-Host "=== end regression_config.yaml ==="

Write-Host "[1/4] rule_candidate mock"
Invoke-KB extract --input ".\testdata\inputs\rule_safety_margin.md" --source "mock-source" --mock --config $Config

Write-Host "[2/4] macro_knowledge mock: rate"
Invoke-KB extract --input ".\testdata\inputs\know_rate.md" --source "mock-source" --mock --force-type "macro_knowledge" --mock-index "1" --config $Config

Write-Host "[3/4] macro_knowledge mock: revenue/income"
Invoke-KB extract --input ".\testdata\inputs\know_revenue_income.md" --source "mock-source" --mock --force-type "macro_knowledge" --mock-index "2" --config $Config

Write-Host "[4/4] validate"
Invoke-KB validate --config $Config

Write-Host ""
Write-Host "=== regression completed ==="
