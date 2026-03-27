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

RUN apk add --no-cache ffmpeg python3 py3-pip && pip3 install --break-system-packages "yt-dlp[default,curl-cffi]"

WORKDIR /app
COPY --from=builder /build/video-downloader-bot .
COPY --from=builder /build/example_config.yaml ./config.yaml

RUN mkdir -p files/static

ENTRYPOINT ["./video-downloader-bot"]
