#!/bin/sh

# yt-dlp is pinned to the latest release at image build time (see Dockerfile)
# On container start it can self-update via pip with YTDLP_SELFUPDATE=true
# The container starts as root only to prepare the bind mounts, then drops to appuser (uid 10001)

APP_USER="appuser"

if [ "$(id -u)" = "0" ]; then
	# Bind-mounted host dirs come in owned by whoever created them (usually root), so the bot can't write downloads or logs
	# Fix ownership on every start, so a fresh deploy or a recreated host dir can't silently break downloads
	mkdir -p files/static files/downloads files/converted logs
	chown -R "$APP_USER:$APP_USER" files logs

	if [ "${YTDLP_SELFUPDATE:-false}" = "true" ]; then
		echo "YTDLP_SELFUPDATE=true: self-updating yt-dlp via pip..."
		if pip3 install -U --pre --no-cache-dir --break-system-packages "yt-dlp[default,curl-cffi]"; then
			echo "yt-dlp updated to $(yt-dlp --version 2>/dev/null)."
		else
			echo "yt-dlp self-update failed, continuing with the version baked into the image."
		fi
	fi

	exec su-exec "$APP_USER" ./video-downloader-bot "$@" # su-exec replaces the shell, so the Go app still becomes PID 1 and receives SIGTERM/SIGINT directly
fi

# Started with --user: no root to fix ownership or run pip, so just run the bot
if [ "${YTDLP_SELFUPDATE:-false}" = "true" ]; then
	echo "YTDLP_SELFUPDATE=true ignored: the container runs without root, update via 'make update-ytdlp' instead."
fi

exec ./video-downloader-bot "$@" # Go application becomes PID 1 and receives SIGTERM/SIGINT directly, preserving its graceful shutdown
