.PHONY: help deps test vet fmt fmt-check check scan validate validate-artifact extract-claims dry-run start smoke worker verify verify-m2-local verify-m3 provider-capabilities demo demo-blocked

BIN ?= build/animus-news
DEMO_OUT ?= build/verify-demo

help:
	@echo "Animus News local commands"
	@echo "  make deps              Download Go module dependencies"
	@echo "  make test              Run Go tests"
	@echo "  make vet               Run go vet"
	@echo "  make fmt-check         Check Go formatting without modifying files"
	@echo "  make scan              Run local secret scan"
	@echo "  make validate          Validate the pilot episode bundle"
	@echo "  make validate-artifact Validate one pilot artifact"
	@echo "  make extract-claims    Extract claims from the pilot script"
	@echo "  make dry-run           Run the safe local MVP dry run"
	@echo "  make start             Alias for dry-run"
	@echo "  make smoke             Run release-readiness local checks"
	@echo "  make worker            Start Temporal worker; requires Temporal service"
	@echo "  make verify            M3 single-signal gate: fmt + build + vet + test + scan + schema + e2e demo"
	@echo "  make verify-m2-local   Run M2 local adapter and determinism checks"
	@echo "  make verify-m3         Run M3 provider boundary, registry, and replay checks"
	@echo "  make provider-capabilities Print provider capability registry JSON"
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

fmt-check:
	@test -z "$$(gofmt -l $$(find . -path ./.git -prune -o -name '*.go' -print))" || \
		(echo "Go files need gofmt:"; gofmt -l $$(find . -path ./.git -prune -o -name '*.go' -print); exit 1)

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

# verify is the single green/red signal for M3. It checks formatting, builds,
# vets, tests, scans for secrets, compiles the CLI, runs the end-to-end mock demo
# (success and failure-injected variants), and schema-validates every produced
# short-form artifact. No network, no secrets, no live provider calls.
verify:
	@echo "==> [1/9] gofmt check"
	@$(MAKE) fmt-check
	@echo "==> [2/9] go build ./..."
	@go build ./...
	@echo "==> [3/9] go vet ./..."
	@go vet ./...
	@echo "==> [4/9] go test ./..."
	@go test ./...
	@echo "==> [5/9] secret scan"
	@$(MAKE) scan
	@echo "==> [6/9] provider capability registry"
	@go run ./cmd/animus-news provider-capabilities >/dev/null
	@echo "==> [7/9] compile CLI -> $(BIN)"
	@go build -o $(BIN) ./cmd/animus-news
	@echo "==> [8/9] end-to-end mock demo (success + failure-injected)"
	@set -e; \
		$(BIN) demo --episode episode-0001 --out $(DEMO_OUT)/success --expect terminal; \
		echo "---"; \
		$(BIN) demo --episode episode-0001 --inject unapproved_storyboard --out $(DEMO_OUT)/blocked --expect blocked:storyboard_image
	@echo "==> [9/9] schema validation of produced short-form artifacts"
	@set -e; for f in $(DEMO_OUT)/success/episode-0001/*_manifest.json \
			$(DEMO_OUT)/success/episode-0001/production_candidate.json \
			$(DEMO_OUT)/success/episode-0001/release_approval.json; do \
		$(BIN) validate-shortform "$$f" >/dev/null; \
	done
	@echo ""
	@echo "M3 VERIFY: GREEN"

verify-m2-local:
	@echo "==> M2 local adapter contract checks"
	@go test ./internal/shortform/providers/localexec \
		./internal/shortform/providers/render \
		./internal/shortform/providers/subtitles \
		./internal/shortform/providers/uploadpost
	@echo "==> M2 workflow determinism checks"
	@go test ./internal/workflows -run 'TestShortFormWorkflow(ReplayIsDeterministic|DeterministicResultFixture)' -count=1

verify-m3:
	@echo "==> M3 provider boundary and registry checks"
	@go test ./internal/shortform/providers/mcp \
		./internal/shortform/providers/render/davinci \
		./internal/shortform/providers/voice/omnivoice \
		./internal/shortform/providers/capabilities
	@echo "==> M3 workflow replay/boundary checks"
	@go test ./internal/workflows -run 'TestShortFormWorkflow(ReplayIsDeterministic|DeterministicResultFixture|BlockedPathFixtures)|TestShortFormWorkflowDoesNotCallExternalProviderBoundaries' -count=1
	@echo "==> M3 provider capability CLI"
	@go run ./cmd/animus-news provider-capabilities >/dev/null

provider-capabilities:
	go run ./cmd/animus-news provider-capabilities

demo:
	go run ./cmd/animus-news demo --episode episode-0001 --out $(DEMO_OUT)/success --expect terminal

demo-blocked:
	go run ./cmd/animus-news demo --episode episode-0001 --inject unapproved_storyboard --out $(DEMO_OUT)/blocked --expect blocked:storyboard_image
