.PHONY: help deps test vet fmt fmt-check check scan validate validate-artifact extract-claims dry-run start smoke worker verify verify-m2-local verify-m3 verify-real-pilot verify-l2-providers verify-mvp-docker provider-capabilities demo demo-blocked

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
	@echo "  make verify-real-pilot Run L1 real CLI pilot fake-provider integration checks"
	@echo "  make verify-l2-providers Run L2 provider checks (fake HTTP + fake external-command; no real calls)"
	@echo "  make verify-mvp-docker Static checks for the Dockerized MVP runtime; no live calls"
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

verify-real-pilot:
	@echo "==> L1 real pilot CLI and fake external-command provider checks"
	@go test ./internal/shortform/pilot ./cmd/animus-news
	@echo "==> L1 connector and workflow documentation presence checks"
	@test -f docs/REAL_PILOT_V1.md
	@test -f docs/CONNECTORS.md
	@test -f docs/WORKFLOW_FINAL.md
	@test -f docs/CONNECTOR_ROADMAP.md
	@test -f docs/PROVIDER_CAPABILITY_MODEL.md
	@test -f docs/ledger/LAUNCH_SLICE_L1.md
	@test -f docs/reports/LAUNCH_SLICE_L1_status.md
	@grep -q "Source and research connectors" docs/CONNECTORS.md
	@grep -q "Visual video generation connectors" docs/CONNECTORS.md
	@grep -q "Future full production workflow" docs/WORKFLOW_FINAL.md
	@grep -q "Review Room" docs/WORKFLOW_FINAL.md
	@grep -q "release_candidate" docs/REAL_PILOT_V1.md

# verify-l2-providers checks the L2 provider integration layer using fake HTTP
# servers and fake external-command providers only. It never calls a real or
# paid provider, needs no secrets, and makes no network calls.
verify-l2-providers:
	@echo "==> L2 native review provider tests (fake HTTP server)"
	@go test ./internal/shortform/providers/review/claude
	@echo "==> L2 pilot api-review + external-command tests (fake providers)"
	@go test ./internal/shortform/pilot
	@echo "==> L2 provider capability registry tests"
	@go test ./internal/shortform/providers/capabilities
	@echo "==> L2 provider docs present"
	@set -e; for f in \
		docs/providers/PROVIDER_RESEARCH_L2.md \
		docs/providers/CLAUDE_API.md \
		docs/providers/CHATTERBOX_TTS.md \
		docs/providers/SEEDANCE2.md \
		docs/providers/OPENAI_API.md \
		docs/providers/CLAUDE_CODE_MCP.md \
		docs/runbooks/first_real_pilot.md \
		docs/runbooks/chatterbox_voice_wrapper.md \
		docs/runbooks/seedance_visual_wrapper.md \
		docs/PRODUCTION_DEPLOYMENT.md \
		.env.example; do \
		test -f "$$f" || { echo "missing $$f"; exit 1; }; \
	done
	@echo "==> L2 capability registry includes new providers and grants no live publish"
	@set -e; caps="$$(go run ./cmd/animus-news provider-capabilities)"; \
		for p in claude_api_review chatterbox_tts_external seedance2_visual_external openai_image claude_code_mcp_operator; do \
			echo "$$caps" | grep -q "\"$$p\"" || { echo "missing capability $$p"; exit 1; }; \
		done; \
		if echo "$$caps" | grep -q '"can_publish": true'; then echo "a provider claims live publish"; exit 1; fi
	@echo "==> L2 no secrets committed"
	@$(MAKE) scan >/dev/null
	@echo ""
	@echo "L2 PROVIDERS VERIFY: GREEN"

verify-mvp-docker:
	@echo "==> MVP Docker static configuration checks"
	@go test ./internal/shortform/pilot -run 'TestMVPDockerStatic|TestMVPDockerEntrypointRejectsEmptyPrompt' -count=1
	@echo "==> MVP Docker secret scan"
	@$(MAKE) scan >/dev/null
	@echo ""
	@echo "MVP DOCKER VERIFY: GREEN"

provider-capabilities:
	go run ./cmd/animus-news provider-capabilities

demo:
	go run ./cmd/animus-news demo --episode episode-0001 --out $(DEMO_OUT)/success --expect terminal

demo-blocked:
	go run ./cmd/animus-news demo --episode episode-0001 --inject unapproved_storyboard --out $(DEMO_OUT)/blocked --expect blocked:storyboard_image
