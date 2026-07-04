# Video Downloader Bot

🇬🇧/[🇷🇺](https://github.com/itskoshkin/video-downloader-bot/blob/master/README-ru.md)/[🇺🇦](https://github.com/itskoshkin/video-downloader-bot/blob/master/README-ua.md)

Telegram bot to download videos from X/Twitter, YouTube Shorts, Instagram and TikTok

- Send a link to bot directly or use [inline mode](https://telegram.org/blog/inline-bots)
- Bot will reply with video and detracked link

**[@cat_video_downloader_bot](https://t.me/cat_video_downloader_bot)**

## Features

- 1️⃣ Send a link (e.g. [youtu.be/lOwxBDDwdDU](https://youtu.be/lOwxBDDwdDU)) to bot, get video
- 2️⃣ Send a link in any chat via inline mode, get video
- 🔗 Supported links:
  - <img src="https://raw.githubusercontent.com/FortAwesome/Font-Awesome/refs/heads/6.x/svgs/brands/x-twitter.svg" width="15" height="10" alt="X/Twitter"> X/Twitter (`https://x.com/<username>/status/<tweet_id>`, `twitter.com/...`, `fxtwitter.com/...`)
  - <img src="https://raw.githubusercontent.com/FortAwesome/Font-Awesome/refs/heads/6.x/svgs/brands/youtube.svg" width="10" height="10" alt="YouTube">⠀YouTube (`youtube.com/shorts/<video_id>`, `youtube.com/watch?v=...`, `youtu.be/...`)
  - <img src="https://raw.githubusercontent.com/FortAwesome/Font-Awesome/refs/heads/6.x/svgs/brands/instagram.svg" width="10" height="10" alt="Instagram">⠀Instagram (`instagram.com/reel/<reel_id>`, `ddinstagram.com/...`, `kkinstagram.com/...`)
  - <img src="https://raw.githubusercontent.com/FortAwesome/Font-Awesome/refs/heads/6.x/svgs/brands/tiktok.svg" width="10" height="10" alt="TikTok">⠀TikTok (`tiktok.com/@<username>/video/<post_id>`, `tiktok.com/t/...`, `vt/vm.tiktok.com/...`)
- ⚠️ Current limits:
  - 10 requests per minute and 100 requests/day (both in DM and inline)
  - Max video duration is 5 minutes
  - Max video size is 20 MB
- ⚙️ Set your preferences via `/settings` command:
  - Bot language (`en`/`ru`/`ua`)
  - Caption style (only video, video and link, full)
  - Fast inline mode (don't wait for video title)
- ❔ Available commands - `/start`, `/help`, `/settings`

## Build & Run

### Prerequisites

- [Go 1.26+](https://gist.github.com/itskoshkin/5f45fca15c30f859955dc146080a00d9)

- [PostgreSQL](https://gist.github.com/itskoshkin/0a9bb715f78f3a6133a8b5daf0ff898a)

- [yt-dlp](https://gist.github.com/itskoshkin/96b25adb0646b478b7b1980893ecd636) (for downloading videos)

- [ffmpeg](https://gist.github.com/itskoshkin/536abbf6aae9d15a8eb396c3ed9df851) (for converting videos, otherwise you may see static image with sound in Telegram)

- [Telegram Bot Token](https://core.telegram.org/bots/faq) (get from [@BotFather](https://t.me/BotFather))

- [Dump Channel ID](#Channel)

### Local

1. Build
   ```bash
   make build
   ```
   or
   ```bash
   go build -o video-downloader-bot ./cmd/main.go
   ```
2. Prepare and edit `config.yaml`
   ```bash
   cp example_config.yaml config.yaml && nano config.yaml
   ```
3. Run
   ```bash
   make run
   ```
   or
   ```bash
   ./video-downloader-bot
   ```

### Docker

1. Build image
   ```bash
   make docker-build
   ```
   or
   ```bash
   docker build -t video-downloader-bot .
   ```
2. Prepare and edit `config.yaml`
   ```bash
   cp example_config.yaml config.yaml && nano config.yaml
   ```
3. Run container
   ```bash
   make docker-run
   ```
   or
   ```bash
   docker run -d --name video-downloader-bot \
		-v $(pwd)/config.yaml:/app/config.yaml:ro \
		-v $(pwd)/files:/app/files \
		-v $(pwd)/logs:/app/logs \
		video-downloader-bot
   ```

## TODO

- [ ] Refactor flat `telegram` package with proper DI
- [ ] Compress video if file size limit reached?
- [ ] Retry on connection errors? (e.g. `context deadline exceeded`)
- [ ] Add webhook?
- [ ] Delete videos in dump channel
- [ ] Add license
- [ ] Add cleanup for rate-limiter (fix memory leak)
- [ ] Add tests? (links.go)
- [ ] Docker container healthcheck
- [ ] Move app settings from YAML config to Postgres (currently the DB only stores bot data; config stays in `config.yaml`)

<details>

<summary><h3>Done</h3></summary>

- [x] Fix square videos

</details>

## Misc

#### Channel

<details>

<summary><h5>Why do I need this?</h5></summary>

Inline mode has a hard time limit of 10 seconds for answering a query  
Most videos take longer than that to download and convert, so the limit is almost always exceeded  
To work around this, the bot sends a placeholder message with a button, retrieves message ID and then edits it once the video is ready  
But editing a message to add media ([EditMessageMedia](https://core.telegram.org/bots/api#editmessagemedia)) requires a `file_id` — an ID of a file already uploaded to Telegram and sent somewhere
Thus, bot needs a channel where it can "dump" videos to get that id

</details>

1. Create a new private channel
2. Add bot to it
3. Give it admin rights (send and delete posts, exactly)
4. Get ID of channel and put it in config

#### Keeping `yt-dlp` up to date

`yt-dlp` breaks often — sites change their players and extractors need frequent updates, so a stale `yt-dlp` will eventually start failing downloads. Always update it **via pip** (never `yt-dlp -U` for a pip-based install).

The Docker image pins the latest release at build time. To update later, run `pip install -U` **inside the running container** — no restart, no downtime, and the bot uses the new `yt-dlp` on the next download:

```bash
make update-ytdlp          # docker exec <container> pip install -U yt-dlp
```

To do this automatically, install a **weekly update** cron job on the host:

```bash
make enable-cron           # adds a weekly `docker exec ... pip install -U yt-dlp` to your crontab (see scripts/enable-cron.sh)
```

Defaults: Monday 04:00, container `video-downloader-bot`, log `/var/log/crons/video-downloader-bot-ytdlp-weekly-update.log` (override via `CONTAINER` / `CRON_SCHEDULE` / `CRON_LOG` env vars). Re-running is safe — it replaces the existing entry instead of duplicating it.

Alternatively, the container can self-update `yt-dlp` **on start** — opt-in, off by default so normal restarts stay fast. Enable it per-run:

```bash
docker run -e YTDLP_SELFUPDATE=true ... video-downloader-bot
```

> Updates are best-effort: if pip can't reach PyPI (e.g. behind DPI), `yt-dlp` stays on the current version.

#### `yt-dlp` impersonation

Some TikTok videos require *impersonation* to be downloaded correctly, otherwise you may receive an error. You will see such cases in log:
```
2026/03/24 22:22:22 WARN yt-dlp: WARNING: [TikTok] The extractor is attempting impersonation, but no impersonate target is available. If you encounter errors, then see https://github.com/yt-dlp/yt-dlp#impersonation  for information on installing the required dependencies
```
To fix this, install `yt-dlp` with `curl-cffi`:
```bash
uv tool install "yt-dlp[default,curl-cffi]"
```
or
```bash
pip3 install "yt-dlp[default,curl-cffi]"
```
Homebrew doesn't have it bundled AFAIK

#### `yt-dlp` cookies

Some videos are rated for adults, sometimes you hit the limit of downloading as a guest, both will result in error:
```bash
2026/03/24 22:22:22 WARN yt-dlp: ERROR: [TikTok] 761.............326: This post may not be comfortable for some audiences. Log in for access. Use --cookies-from-browser or --cookies for the authentication. See https://github.com/yt-dlp/yt-dlp/wiki/FAQ#how-do-i-pass-cookies-to-yt-dlp  for how to manually pass cookies
```
To fix this, place `cookies.txt` in `./files/static/` (or edit config to comply with your file and path)

See [yt-dlp wiki](https://github.com/yt-dlp/yt-dlp/wiki/FAQ#how-do-i-pass-cookies-to-yt-dlp) for instructions on how to get the cookies file

TLDR – you have to register a *throw-away* accounts, open incognito window with YouTube/Instagram/Tiktok tabs, login in each tab, then use [browser extension](https://chromewebstore.google.com/detail/get-cookiestxt-locally/cclelndahbckbenkjhflpdbgdldlbecc) to export cookies as TXT file

#### Non-root container

The Docker image runs as a non-root user (uid **10001**). With bind mounts (`-v $PWD/files:/app/files`, `-v $PWD/logs:/app/logs`) the **host** `files/` and `logs/` dirs must be writable by that uid — e.g. `sudo chown -R 10001 files logs` on the host — otherwise the bot can't write downloads or logs. Updating `yt-dlp` inside a running non-root container needs root, so `make update-ytdlp` (and the weekly cron) run `docker exec -u root`.

#### Error messages

Download failures are classified before they reach the user: auth/age-gated content (login required, age-restricted, "log in for access", …) shows a localized 🔒 "behind a login/age gate" message instead of a raw `yt-dlp` dump, while the full error is still logged server-side.

#### File cleanup

A background sweeper deletes leftover files older than 30 min from the download and converted folders every 15 min (it skips `files/static/` and dotfiles), so orphaned fragments from failed or interrupted downloads don't pile up.

#### Log retention

Rotated app logs are pruned by count / total size / age (`app.log.max_old_files` / `max_old_size_mb` / `max_old_age_days`, defaults `10 / 0 / 30`, where `0` = unlimited) and can be gzipped (`app.log.gzip_old_logs`). Consecutive identical log lines are collapsed into a single `(previous message repeated N more time(s))`.

#### Reconnect backoff & DB pool

On Telegram polling errors the bot backs off with capped exponential delay (1s → 15s, reset after 2 min of stability) instead of hammering the API. The Postgres connection pool is tunable via `app.database.max_idle_conns` / `max_open_conns` (defaults `2 / 10`).
