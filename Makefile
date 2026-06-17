.PHONY: help deps test vet fmt check scan validate validate-artifact extract-claims dry-run start smoke worker verify demo demo-blocked

BIN ?= build/animus-news
DEMO_OUT ?= build/verify-demo

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
	@echo "  make verify            M1 single-signal gate: build + vet + test + schema + e2e demo"
	@echo "  make demo              Run the short-form mock demo (success path)"
	@echo "  make demo-blocked      Run the short-form mock demo with an injected gate failure"

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

# verify is the single green/red signal for M1. It builds, vets, tests, compiles
# the CLI, runs the end-to-end mock demo (success and failure-injected variants),
# and schema-validates every produced short-form artifact. No network, no secrets.
verify:
	@echo "==> [1/6] go build ./..."
	@go build ./...
	@echo "==> [2/6] go vet ./..."
	@go vet ./...
	@echo "==> [3/6] go test ./..."
	@go test ./...
	@echo "==> [4/6] compile CLI -> $(BIN)"
	@go build -o $(BIN) ./cmd/animus-news
	@echo "==> [5/6] end-to-end mock demo (success + failure-injected)"
	@set -e; \
		$(BIN) demo --episode episode-0001 --out $(DEMO_OUT)/success --expect terminal; \
		echo "---"; \
		$(BIN) demo --episode episode-0001 --inject unapproved_storyboard --out $(DEMO_OUT)/blocked --expect blocked:storyboard_image
	@echo "==> [6/6] schema validation of produced short-form artifacts"
	@set -e; for f in $(DEMO_OUT)/success/episode-0001/*_manifest.json \
			$(DEMO_OUT)/success/episode-0001/production_candidate.json \
			$(DEMO_OUT)/success/episode-0001/release_approval.json; do \
		$(BIN) validate-shortform "$$f" >/dev/null; \
	done
	@echo ""
	@echo "M1 VERIFY: GREEN"

demo:
	go run ./cmd/animus-news demo --episode episode-0001 --out $(DEMO_OUT)/success --expect terminal

demo-blocked:
	go run ./cmd/animus-news demo --episode episode-0001 --inject unapproved_storyboard --out $(DEMO_OUT)/blocked --expect blocked:storyboard_image
