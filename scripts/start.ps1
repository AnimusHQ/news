$ErrorActionPreference = "Stop"

Write-Host "==> Starting Animus News local MVP dry run"
go run ./cmd/animus-news dry-run episodes/0001-after-git-push
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}
