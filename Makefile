.PHONY: help deps test vet fmt check scan validate validate-artifact extract-claims dry-run start smoke worker

help:
	@echo "Animus News local commands"
	@echo "  make deps              Download Go module dependencies"
	@echo "  make test              Run Go tests"
	@echo "  make vet               Run go vet"
	@echo "  make scan              Run local secret scan"
	@echo "  make validate          Validate the pilot episode bundle"
	@echo "  make validate-artifact Validate one pilot artifact"
	@echo "  make extract-claims    Extract claims from the pilot script"
	@echo "  make dry-run           Run the safe local MVP dry run"
	@echo "  make start             Alias for dry-run"
	@echo "  make smoke             Run release-readiness local checks"
	@echo "  make worker            Start Temporal worker; requires Temporal service"

deps:
	go mod download

test:
	go test ./...

vet:
	go vet ./...

fmt:
	go fmt ./...

check: fmt vet test

scan:
	go run ./cmd/animus-news scan-secrets .

validate:
	go run ./cmd/animus-news validate-episode episodes/0001-after-git-push

validate-artifact:
	go run ./cmd/animus-news validate --json episodes/0001-after-git-push/research_pack.json

extract-claims:
	go run ./cmd/animus-news extract-claims episodes/0001-after-git-push

dry-run:
	go run ./cmd/animus-news dry-run episodes/0001-after-git-push

start: dry-run

smoke: test vet scan validate validate-artifact extract-claims dry-run

worker:
	go run ./cmd/animus-news worker
