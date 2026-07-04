#!/bin/sh

# Installs a weekly cron job on the host that updates yt-dlp inside the running bot container (docker exec pip install -U), keeping it fresh without a restart or a rebuild
# Override defaults via env vars, e.g.: CONTAINER=my-bot CRON_SCHEDULE="0 4 * * 0" CRON_LOG=~/update.log ./scripts/enable-cron.sh
# Idempotent: re-running replaces the previous entry instead of duplicating it

set -e

CONTAINER="${CONTAINER:-video-downloader-bot}"
CRON_SCHEDULE="${CRON_SCHEDULE:-0 4 * * 1}" # Mon 04:00
CRON_LOG="${CRON_LOG:-/var/log/crons/video-downloader-bot-ytdlp-weekly-update.log}"
mkdir -p "$(dirname "$CRON_LOG")"

DOCKER_BIN="$(command -v docker || true)"
if [ -z "$DOCKER_BIN" ]; then
	echo "error: 'docker' not found in PATH" >&2
	exit 1
fi

MARKER="# ytdlp-weekly-update:${CONTAINER}" # Inline marker makes the entry findable/replaceable
CRON_LINE="${CRON_SCHEDULE} ${DOCKER_BIN} exec -u root ${CONTAINER} pip3 install -U --no-cache-dir --break-system-packages 'yt-dlp[default,curl-cffi]' >> ${CRON_LOG} 2>&1 ${MARKER}"

( crontab -l 2>/dev/null | grep -vF "$MARKER" || true; echo "$CRON_LINE" ) | crontab - # Drop any previous entry for this container, then append the current one

echo "Installed weekly yt-dlp update cron:"
echo "  $CRON_LINE"
