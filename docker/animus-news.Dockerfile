FROM golang:1.24-bookworm

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        bash \
        ca-certificates \
        curl \
        ffmpeg \
        python3 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN chmod +x \
        /app/docker/animus_news_entrypoint.sh \
        /app/scripts/providers/chatterbox-voice-wrapper.example.py \
        /app/scripts/providers/seedance2-visual-wrapper.example.py

ENV ANIMUS_VISUAL_COMMAND=/app/scripts/providers/seedance2-visual-wrapper.example.py
ENV ANIMUS_VISUAL_INPUT_ROOT=/workspace/episodes
ENV ANIMUS_VISUAL_OUTPUT_ROOT=/workspace/episodes
ENV ANIMUS_VISUAL_TIMEOUT=1800s

ENV ANIMUS_VOICE_COMMAND=/app/scripts/providers/chatterbox-voice-wrapper.example.py
ENV ANIMUS_VOICE_INPUT_ROOT=/workspace/episodes
ENV ANIMUS_VOICE_OUTPUT_ROOT=/workspace/episodes
ENV ANIMUS_VOICE_TIMEOUT=600s
ENV CHATTERBOX_BASE_URL=http://chatterbox:4123

ENV ANIMUS_FFMPEG_BINARY=ffmpeg
ENV ANIMUS_FFPROBE_BINARY=ffprobe
ENV ANIMUS_FFMPEG_TIMEOUT=600s

ENTRYPOINT ["/app/docker/animus_news_entrypoint.sh"]
