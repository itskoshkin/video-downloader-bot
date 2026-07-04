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

RUN apk add --no-cache ffmpeg python3 py3-pip && pip3 install -U --break-system-packages "yt-dlp[default,curl-cffi]"

WORKDIR /app
COPY --from=builder /build/video-downloader-bot .
COPY --from=builder /build/example_config.yaml ./config.yaml
COPY scripts/entrypoint.sh ./
RUN chmod +x entrypoint.sh

RUN mkdir -p files/static files/downloads files/converted logs

# Run as a non-root user with ownership of the dirs it writes to (files/, logs/).
# With bind-mounted volumes the HOST dirs must be writable by this uid (10001) — see README.
RUN adduser -D -H -u 10001 appuser && chown -R appuser:appuser /app
USER appuser

# yt-dlp self-update on container start is opt-in (best-effort); enable with -e YTDLP_SELFUPDATE=true.
# As non-root, pip can't update the system yt-dlp, so the self-update is skipped — update via
# `make update-ytdlp` (execs pip as root) or by rebuilding the image.
ENV YTDLP_SELFUPDATE=false

ENTRYPOINT ["./entrypoint.sh"]
