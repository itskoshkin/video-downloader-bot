#!/bin/sh

# yt-dlp is pinned to the latest release at image build time (see Dockerfile)
# On container start it can self-update via pip with YTDLP_SELFUPDATE=true

if [ "${YTDLP_SELFUPDATE:-false}" = "true" ]; then
	echo "YTDLP_SELFUPDATE=true: self-updating yt-dlp via pip..."
	if pip3 install -U --no-cache-dir --break-system-packages "yt-dlp[default,curl-cffi]"; then
		echo "yt-dlp updated to $(yt-dlp --version 2>/dev/null)."
	else
		echo "yt-dlp self-update failed, continuing with the version baked into the image."
	fi
fi

exec ./video-downloader-bot "$@" # Go application becomes PID 1 and receives SIGTERM/SIGINT directly, preserving the it's graceful shutdown
