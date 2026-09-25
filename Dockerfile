FROM golang:1.26-alpine AS builder

ARG BIN_NAME="video-downloader-bot"
ARG MAIN_PATH="./cmd/main.go"
ARG GO_BUILD_FLAGS="-s -w"

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -ldflags="$GO_BUILD_FLAGS" -o $BIN_NAME $MAIN_PATH

FROM alpine:3.21

# --pre pulls the yt-dlp nightly channel — extractor fixes land there days before a stable release
RUN apk add --no-cache ffmpeg python3 py3-pip su-exec && pip3 install -U --pre --break-system-packages "yt-dlp[default,curl-cffi]"

WORKDIR /app
COPY --from=builder /build/video-downloader-bot .
COPY --from=builder /build/example_config.yaml ./config.yaml
COPY scripts/entrypoint.sh ./
RUN chmod +x entrypoint.sh

RUN mkdir -p files/static files/downloads files/converted logs

# The bot runs as a non-root user (uid 10001).
# No USER directive: entrypoint.sh starts as root, chowns the bind-mounted files/ and logs/ to appuser, then drops privileges via su-exec.
RUN adduser -D -H -u 10001 appuser && chown -R appuser:appuser /app

# yt-dlp self-update on container start is opt-in (best-effort); enable with -e YTDLP_SELFUPDATE=true.
# It runs as root in entrypoint.sh before the privilege drop; otherwise update via `make update-ytdlp` or by rebuilding the image.
ENV YTDLP_SELFUPDATE=false

ENTRYPOINT ["./entrypoint.sh"]
