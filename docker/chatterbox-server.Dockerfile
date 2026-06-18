FROM python:3.11-slim

ARG CHATTERBOX_TTS_API_REF=main

ENV DEBIAN_FRONTEND=noninteractive
ENV PYTHONDONTWRITEBYTECODE=1
ENV PYTHONUNBUFFERED=1
ENV CHATTERBOX_API_DIR=/opt/chatterbox-tts-api
ENV CHATTERBOX_HOST=0.0.0.0
ENV CHATTERBOX_PORT=4123

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        ca-certificates \
        curl \
        ffmpeg \
        git \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /opt

RUN git clone --depth 1 https://github.com/travisvn/chatterbox-tts-api.git chatterbox-tts-api \
    && cd chatterbox-tts-api \
    && git checkout "${CHATTERBOX_TTS_API_REF}"

WORKDIR /opt/chatterbox-tts-api

RUN python -m pip install --no-cache-dir --upgrade pip \
    && if [ -f requirements.txt ]; then \
        python -m pip install --no-cache-dir -r requirements.txt; \
      else \
        python -m pip install --no-cache-dir fastapi uvicorn chatterbox-tts; \
      fi

COPY docker/chatterbox_server.py /app/chatterbox_server.py

EXPOSE 4123

HEALTHCHECK --interval=20s --timeout=10s --retries=30 --start-period=60s \
  CMD curl -fsS http://127.0.0.1:4123/health >/dev/null || exit 1

CMD ["python", "/app/chatterbox_server.py"]
