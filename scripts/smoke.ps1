$ErrorActionPreference = "Stop"

function Invoke-Step {
    param(
        [string]$Name,
        [scriptblock]$Command
    )

    Write-Host "==> $Name"
    & $Command
    if ($LASTEXITCODE -ne 0) {
        exit $LASTEXITCODE
    }
}

Invoke-Step "go test ./..." { go test ./... }
Invoke-Step "go vet ./..." { go vet ./... }
Invoke-Step "scan secrets" { go run ./cmd/animus-news scan-secrets . }
Invoke-Step "validate pilot episode" { go run ./cmd/animus-news validate-episode episodes/0001-after-git-push }
Invoke-Step "validate pilot research artifact" { go run ./cmd/animus-news validate --json episodes/0001-after-git-push/research_pack.json }
Invoke-Step "dry-run pilot episode" { go run ./cmd/animus-news dry-run episodes/0001-after-git-push }
