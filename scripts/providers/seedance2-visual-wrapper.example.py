#!/usr/bin/env python3
# NON-PRODUCTION EXAMPLE — Seedance 2 -> Animus external visual contract.
#
# Bridges the Animus `external_command_visual` boundary to the Seedance 2 async
# video API. Reads the Animus visual request JSON on stdin; for each shot it
# submits a task, polls until completed, downloads the MP4 into the requested
# output_dir, and prints the Animus visual response JSON on stdout.
#
# Standard library only. Never prints the API key. Exits non-zero on missing
# output and never writes outside output_dir / never publishes. This is a sample
# to adapt, not a supported component; default tests do NOT use it. See
# docs/providers/SEEDANCE2.md.
#
# Configure (wrapper-only env, never committed):
#   ANIMUS_VISUAL_COMMAND=/abs/path/to/this/script
#   ANIMUS_ALLOW_LIVE_PROVIDER_CALLS=1       (required for real calls)
#   ANIMUS_VISUAL_OUTPUT_ROOT=/abs/path/to/episode-or-episodes-root
#   SEEDANCE_API_KEY=<seedance-api-key>       (required)
#   SEEDANCE_BASE_URL=https://api.seedance2.ai   (default)
#   SEEDANCE_MODEL=seedance-2-0          (default)
#   SEEDANCE_POLL_TIMEOUT=600            (seconds, default)

import json
import os
import sys
import time
import urllib.error
import urllib.request


def fail(msg):
    print(f"seedance-wrapper: {msg}", file=sys.stderr)
    sys.exit(1)


def require_live_guard():
    if os.environ.get("ANIMUS_ALLOW_LIVE_PROVIDER_CALLS") != "1":
        fail("ANIMUS_ALLOW_LIVE_PROVIDER_CALLS=1 is required for Seedance network calls")


def contained_output_dir(path):
    root = os.environ.get("ANIMUS_VISUAL_OUTPUT_ROOT", "").strip()
    if not root:
        fail("ANIMUS_VISUAL_OUTPUT_ROOT is not set")
    out_real = os.path.realpath(path)
    root_real = os.path.realpath(root)
    try:
        if os.path.commonpath([root_real, out_real]) != root_real:
            fail("output_dir is outside ANIMUS_VISUAL_OUTPUT_ROOT")
    except ValueError:
        fail("output_dir is outside ANIMUS_VISUAL_OUTPUT_ROOT")
    return out_real


def api(method, url, key, body=None):
    data = json.dumps(body).encode("utf-8") if body is not None else None
    req = urllib.request.Request(
        url,
        data=data,
        method=method,
        headers={"content-type": "application/json", "authorization": "Bearer " + key},
    )
    try:
        with urllib.request.urlopen(req, timeout=120) as resp:
            return json.load(resp)
    except urllib.error.HTTPError as exc:
        # Do not echo the response body — avoid leaking anything sensitive.
        fail(f"seedance http error: status {exc.code}")
    except urllib.error.URLError as exc:
        fail(f"seedance request failed: {exc.reason}")


def download(url, out_path):
    try:
        with urllib.request.urlopen(url, timeout=300) as resp, open(out_path, "wb") as handle:
            handle.write(resp.read())
    except Exception as exc:  # noqa: BLE001 - example script
        fail(f"download failed: {exc}")


def main():
    require_live_guard()
    try:
        req = json.load(sys.stdin)
    except Exception as exc:  # noqa: BLE001 - example script
        fail(f"invalid request json: {exc}")

    key = os.environ.get("SEEDANCE_API_KEY", "").strip()
    if not key:
        fail("SEEDANCE_API_KEY is not set")
    base = os.environ.get("SEEDANCE_BASE_URL", "https://api.seedance2.ai").rstrip("/")
    model = os.environ.get("SEEDANCE_MODEL", "seedance-2-0")
    poll_timeout = int(os.environ.get("SEEDANCE_POLL_TIMEOUT", "600"))

    episode_id = req.get("episode_id")
    output_dir = req.get("output_dir")
    shots_in = req.get("shots") or []
    if not episode_id or not output_dir or not shots_in:
        fail("request missing episode_id, output_dir, or shots")
    output_dir = contained_output_dir(output_dir)
    os.makedirs(output_dir, exist_ok=True)

    out_shots = []
    for shot in shots_in:
        shot_id = shot.get("shot_id")
        if not shot_id:
            fail("shot missing shot_id")
        duration = max(4, min(15, int(round(shot.get("duration_sec") or 5))))
        created = api("POST", base + "/v1/videos/generations", key, {
            "model": model,
            "input": {
                "prompt": shot.get("prompt", ""),
                "generation_type": "text-to-video",
                "duration": duration,
                "aspect_ratio": "9:16",
                "resolution": "1080p",
            },
        })
        task_id = created.get("taskId") or created.get("id")
        if not task_id:
            fail(f"no task id for {shot_id}")

        url = None
        deadline = time.time() + poll_timeout
        while time.time() < deadline:
            status = api("GET", f"{base}/v1/tasks/{task_id}", key)
            state = status.get("status")
            if state == "completed":
                results = (status.get("data") or {}).get("results") or []
                if not results:
                    fail(f"completed without results for {shot_id}")
                url = results[0]
                break
            if state == "failed":
                fail(f"generation failed for {shot_id}")
            time.sleep(5)
        if not url:
            fail(f"timed out for {shot_id}")

        out_path = os.path.join(output_dir, f"{shot_id}.mp4")
        download(url, out_path)
        if not os.path.exists(out_path) or os.path.getsize(out_path) == 0:
            fail(f"missing output for {shot_id}")

        # Animus requires exactly 1080x1920 / 30fps. Ensure Seedance is configured
        # to produce that; this wrapper reports the requested geometry.
        out_shots.append({
            "shot_id": shot_id,
            "status": "generated",
            "output_path": out_path,
            "duration_sec": shot.get("duration_sec", duration),
            "width": shot.get("width", 1080),
            "height": shot.get("height", 1920),
            "fps": shot.get("fps", 30),
        })

    json.dump({
        "schema_version": "1.0",
        "episode_id": episode_id,
        "provider": "seedance2",
        "shots": out_shots,
    }, sys.stdout)
    sys.stdout.write("\n")


if __name__ == "__main__":
    main()
