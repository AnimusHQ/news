#!/usr/bin/env bash
set -euo pipefail

fail() {
  printf 'animus-news mvp docker: %s\n' "$1" >&2
  exit 64
}

trim() {
  local value="$1"
  value="${value#"${value%%[![:space:]]*}"}"
  value="${value%"${value##*[![:space:]]}"}"
  printf '%s' "$value"
}

require_var() {
  local name="$1"
  local value
  value="$(trim "${!name:-}")"
  if [[ -z "$value" ]]; then
    fail "$name is required"
  fi
  export "$name=$value"
}

require_var EPISODE_ID
require_var PROMPT
require_var LANGUAGE
require_var DURATION
require_var PLATFORMS
require_var ANTHROPIC_API_KEY
require_var SEEDANCE_API_KEY

if [[ "$(trim "${ANIMUS_ALLOW_LIVE_PROVIDER_CALLS:-}")" != "1" ]]; then
  fail "ANIMUS_ALLOW_LIVE_PROVIDER_CALLS=1 is required"
fi

exec go run ./cmd/animus-news pilot generate-real \
  --episode-id "$EPISODE_ID" \
  --prompt "$PROMPT" \
  --language "$LANGUAGE" \
  --duration "$DURATION" \
  --platforms "$PLATFORMS" \
  --visual-provider external-command \
  --voice-provider external-command \
  --subtitle-provider script-timing \
  --render-provider ffmpeg \
  --claude-review api \
  --out "/workspace/episodes/$EPISODE_ID"
