#!/usr/bin/env python3
# NON-PRODUCTION EXAMPLE — Chatterbox TTS -> Animus external voice contract.
#
# Bridges the Animus `external_command_voice` boundary to a self-hosted
# Chatterbox TTS server. Reads the Animus voice request JSON on stdin, calls
# Chatterbox, writes a WAV into the requested output_dir, and prints the Animus
# voice response JSON on stdout.
#
# Standard library only (no pip install). Never prints secrets. Exits non-zero on
# any error. This is a sample to adapt, not a supported component; default tests
# do NOT use it. See docs/providers/CHATTERBOX_TTS.md.
#
# Configure (wrapper-only env, never committed):
#   ANIMUS_VOICE_COMMAND=/abs/path/to/this/script
#   ANIMUS_ALLOW_LIVE_PROVIDER_CALLS=1       (required for real calls)
#   ANIMUS_VOICE_OUTPUT_ROOT=/abs/path/to/episode-or-episodes-root
#   CHATTERBOX_BASE_URL=http://localhost:4123   (required)
#   CHATTERBOX_VOICE=<voice name>               (optional)
#   CHATTERBOX_VOICE_CONSENT_REFERENCE=<id>     (required if cloning a voice)

import json
import os
import sys
import urllib.error
import urllib.request
import wave


def fail(msg):
    print(f"chatterbox-wrapper: {msg}", file=sys.stderr)
    sys.exit(1)


def require_live_guard():
    if os.environ.get("ANIMUS_ALLOW_LIVE_PROVIDER_CALLS") != "1":
        fail("ANIMUS_ALLOW_LIVE_PROVIDER_CALLS=1 is required for Chatterbox network calls")


def contained_output_dir(path):
    root = os.environ.get("ANIMUS_VOICE_OUTPUT_ROOT", "").strip()
    if not root:
        fail("ANIMUS_VOICE_OUTPUT_ROOT is not set")
    out_real = os.path.realpath(path)
    root_real = os.path.realpath(root)
    try:
        if os.path.commonpath([root_real, out_real]) != root_real:
            fail("output_dir is outside ANIMUS_VOICE_OUTPUT_ROOT")
    except ValueError:
        fail("output_dir is outside ANIMUS_VOICE_OUTPUT_ROOT")
    return out_real


def wav_info(path):
    try:
        with wave.open(path, "rb") as w:
            rate = w.getframerate()
            frames = w.getnframes()
            return (frames / float(rate)) if rate else 0.0, rate
    except wave.Error:
        return 0.0, 0


def main():
    require_live_guard()
    try:
        req = json.load(sys.stdin)
    except Exception as exc:  # noqa: BLE001 - example script
        fail(f"invalid request json: {exc}")

    episode_id = req.get("episode_id")
    text = (req.get("text") or "").strip()
    output_dir = req.get("output_dir")
    if not episode_id or not output_dir or not text:
        fail("request missing episode_id, text, or output_dir")
    output_dir = contained_output_dir(output_dir)

    base = os.environ.get("CHATTERBOX_BASE_URL", "").strip().rstrip("/")
    if not base:
        fail("CHATTERBOX_BASE_URL is not set")
    os.makedirs(output_dir, exist_ok=True)
    out_path = os.path.join(output_dir, "voiceover.wav")

    payload = {"input": text, "stream_format": "audio"}
    voice = os.environ.get("CHATTERBOX_VOICE")
    if voice:
        payload["voice"] = voice

    http_req = urllib.request.Request(
        base + "/v1/audio/speech",
        data=json.dumps(payload).encode("utf-8"),
        headers={"content-type": "application/json"},
        method="POST",
    )
    key = os.environ.get("CHATTERBOX_API_KEY", "").strip()
    if key:
        http_req.add_header("authorization", "Bearer " + key)
    try:
        with urllib.request.urlopen(http_req, timeout=600) as resp:
            audio = resp.read()
    except urllib.error.HTTPError as exc:
        fail(f"chatterbox http error: status {exc.code}")
    except urllib.error.URLError as exc:
        fail(f"chatterbox request failed: {exc.reason}")

    if not audio:
        fail("chatterbox returned empty audio")
    with open(out_path, "wb") as handle:
        handle.write(audio)

    duration, sample_rate = wav_info(out_path)
    response = {
        "schema_version": "1.0",
        "episode_id": episode_id,
        "provider": "chatterbox",
        "output_path": out_path,
        "duration_sec": round(duration, 2),
        "sample_rate": sample_rate,
    }
    consent = os.environ.get("CHATTERBOX_VOICE_CONSENT_REFERENCE")
    if consent:
        response["voice_consent_reference"] = consent

    json.dump(response, sys.stdout)
    sys.stdout.write("\n")


if __name__ == "__main__":
    main()
