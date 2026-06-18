#!/usr/bin/env python3
"""Launch the upstream Chatterbox TTS API behind the MVP service contract."""

import os
import sys
from typing import Iterable, Tuple


def _load_upstream_app():
    api_dir = os.environ.get("CHATTERBOX_API_DIR", "/opt/chatterbox-tts-api")
    if api_dir not in sys.path:
        sys.path.insert(0, api_dir)
    try:
        from app.main import app  # type: ignore
    except Exception as exc:  # noqa: BLE001 - startup blocker should be explicit.
        print(
            "chatterbox-server: failed to load upstream Chatterbox API app: "
            f"{exc.__class__.__name__}: {exc}",
            file=sys.stderr,
        )
        sys.exit(78)
    return app


class BearerAuthApp:
    def __init__(self, app, api_key: str):
        self.app = app
        self.api_key = api_key.strip()

    async def __call__(self, scope, receive, send):
        if self.api_key and scope.get("type") == "http" and scope.get("path") != "/health":
            headers = _headers(scope.get("headers") or [])
            expected = "Bearer " + self.api_key
            if headers.get("authorization") != expected:
                await _send_unauthorized(send)
                return
        await self.app(scope, receive, send)


def _headers(raw: Iterable[Tuple[bytes, bytes]]) -> dict[str, str]:
    out: dict[str, str] = {}
    for key, value in raw:
        out[key.decode("latin1").lower()] = value.decode("latin1")
    return out


async def _send_unauthorized(send):
    body = b'{"error":"unauthorized"}'
    await send(
        {
            "type": "http.response.start",
            "status": 401,
            "headers": [
                (b"content-type", b"application/json"),
                (b"content-length", str(len(body)).encode("ascii")),
            ],
        }
    )
    await send({"type": "http.response.body", "body": body})


def main() -> None:
    try:
        import uvicorn
    except Exception as exc:  # noqa: BLE001
        print(f"chatterbox-server: uvicorn is unavailable: {exc}", file=sys.stderr)
        sys.exit(78)

    app = BearerAuthApp(_load_upstream_app(), os.environ.get("CHATTERBOX_API_KEY", ""))
    host = os.environ.get("CHATTERBOX_HOST", "0.0.0.0")
    port = int(os.environ.get("CHATTERBOX_PORT", "4123"))
    uvicorn.run(app, host=host, port=port, access_log=False)


if __name__ == "__main__":
    main()
