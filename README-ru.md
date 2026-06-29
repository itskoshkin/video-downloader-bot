# Video Downloader Bot

[🇬🇧](https://github.com/itskoshkin/video-downloader-bot/blob/master/README.md)/🇷🇺/[🇺🇦](https://github.com/itskoshkin/video-downloader-bot/blob/master/README-ua.md)

Telegram-бот для скачивания видео из X/Twitter, YouTube Shorts, Instagram и TikTok

- Отправь ссылку боту напрямую или используй [инлайн-режим](https://telegram.org/blog/inline-bots)
- Бот в ответ пришлёт видео и ссылку без трекеров

**[@cat_video_downloader_bot](https://t.me/cat_video_downloader_bot)**

## Фичи

- 1️⃣ Отправь ссылку (например [youtu.be/lOwxBDDwdDU](https://youtu.be/lOwxBDDwdDU)) боту, получи видео
- 2️⃣ Отправь ссылку в любой чат через инлайн-режим, получи видео
- 🔗 Поддерживаемые ссылки:
  - <img src="https://raw.githubusercontent.com/FortAwesome/Font-Awesome/refs/heads/6.x/svgs/brands/x-twitter.svg" width="15" height="10" alt="X/Twitter"> X/Twitter (`https://x.com/<username>/status/<tweet_id>`, `twitter.com/...`, `fxtwitter.com/...`)
  - <img src="https://raw.githubusercontent.com/FortAwesome/Font-Awesome/refs/heads/6.x/svgs/brands/youtube.svg" width="10" height="10" alt="YouTube">⠀YouTube (`youtube.com/shorts/<video_id>`, `youtube.com/watch?v=...`, `youtu.be/...`)
  - <img src="https://raw.githubusercontent.com/FortAwesome/Font-Awesome/refs/heads/6.x/svgs/brands/instagram.svg" width="10" height="10" alt="Instagram">⠀Instagram (`instagram.com/reel/<reel_id>`, `ddinstagram.com/...`, `kkinstagram.com/...`)
  - <img src="https://raw.githubusercontent.com/FortAwesome/Font-Awesome/refs/heads/6.x/svgs/brands/tiktok.svg" width="10" height="10" alt="TikTok">⠀TikTok (`tiktok.com/@<username>/video/<post_id>`, `tiktok.com/t/...`, `vt/vm.tiktok.com/...`)
- ⚠️ Текущие ограничения:
  - 10 запросов в минуту и 100 запросов в день (как в личных сообщениях, так и в инлайн-режиме)
  - Максимальная длительность видео — 5 минут
  - Максимальный размер файла — 20 МБ
- ⚙️ Настройки через команду `/settings`:
  - Язык бота (`en`/`ru`/`ua`)
  - Стиль подписи (только видео, видео и ссылка, полностью)
  - Быстрый инлайн-режим (не ждать загрузки названия видео)
- ❔ Доступные команды — `/start`, `/help`, `/settings`

## Сборка и запуск

### Требования

- [Go 1.26+](https://gist.github.com/itskoshkin/5f45fca15c30f859955dc146080a00d9)

- [PostgreSQL](https://gist.github.com/itskoshkin/0a9bb715f78f3a6133a8b5daf0ff898a)

- [yt-dlp](https://gist.github.com/itskoshkin/96b25adb0646b478b7b1980893ecd636) (для скачивания видео)

- [ffmpeg](https://gist.github.com/itskoshkin/536abbf6aae9d15a8eb396c3ed9df851) (для конвертации видео, иначе в Telegram может отображаться статичная картинка со звуком)

- [Токен Telegram-бота](https://core.telegram.org/bots/faq) (получить у [@BotFather](https://t.me/BotFather))

- [ID канала-хранилища](#Канал)

### Локально

1. Собрать
   ```bash
   make build
   ```
   или
   ```bash
   go build -o video-downloader-bot ./cmd/main.go
   ```
2. Подготовить и отредактировать `config.yaml`
   ```bash
   cp example_config.yaml config.yaml && nano config.yaml
   ```
3. Запустить
   ```bash
   make run
   ```
   или
   ```bash
   ./video-downloader-bot
   ```

### Docker

1. Собрать image
   ```bash
   make docker-build
   ```
   или
   ```bash
   docker build -t video-downloader-bot .
   ```
2. Подготовить и отредактировать `config.yaml`
   ```bash
   cp example_config.yaml config.yaml && nano config.yaml
   ```
3. Запустить контейнер
   ```bash
   make docker-run
   ```
   или
   ```bash
   docker run -d --name video-downloader-bot \
	-v $(pwd)/config.yaml:/app/config.yaml:ro \
	-v $(pwd)/files:/app/files \
	-v $(pwd)/logs:/app/logs \
	video-downloader-bot
   ```

## Разное

#### Канал

<details>

<summary><h5>Зачем он нужен?</h5></summary>

Инлайн-режим имеет жёсткий лимит в 10 секунд на ответ на запрос
Большинство видео скачивается и конвертируется дольше, поэтому лимит почти всегда превышается
В качестве обходного решения бот отправляет сообщение-заглушку с кнопкой, получает ID сообщения и редактирует его, когда видео готово
Но редактирование сообщения для добавления медиа ([EditMessageMedia](https://core.telegram.org/bots/api#editmessagemedia)) требует `file_id` — идентификатор файла, уже загруженного в Telegram и отправленного куда-либо
Поэтому боту нужен канал, куда он может «сбрасывать» видео, чтобы получить этот ID

</details>

1. Создай новый приватный канал
2. Добавь в него бота
3. Дай ему права администратора (публикация и удаление постов)
4. Получи ID канала и укажи его в конфиге

#### Поддержание `yt-dlp` в актуальном состоянии

`yt-dlp` часто ломается — сайты меняют свои плееры, и экстракторы требуют частых обновлений, поэтому устаревший `yt-dlp` рано или поздно начнёт ронять загрузки (это и есть сигнал обновиться). Обновляй его **только через pip** (никогда не `yt-dlp -U` для pip-установки).

Docker-образ фиксирует последнюю версию на момент сборки. Чтобы обновить позже, запусти `pip install -U` **внутри работающего контейнера** — без рестарта, без даунтайма, и бот подхватит новый `yt-dlp` на следующей загрузке:

```bash
make update-ytdlp          # docker exec <container> pip install -U yt-dlp
```

Чтобы делать это автоматически, поставь на хост еженедельный крон:

```bash
make enable-cron           # добавляет в crontab еженедельный `docker exec ... pip install -U yt-dlp` (см. scripts/enable-cron.sh)
```

По умолчанию: понедельник 04:00, контейнер `video-downloader-bot`, лог `/var/log/crons/video-downloader-bot-ytdlp-weekly-update.log` (переопределяется через `CONTAINER` / `CRON_SCHEDULE` / `CRON_LOG`). Повторный запуск безопасен — запись заменяется, а не дублируется.

Как вариант, контейнер может сам обновлять `yt-dlp` **на старте** — это опционально и по умолчанию выключено, чтобы обычные рестарты были быстрыми. Включить для конкретного запуска:

```bash
docker run -e YTDLP_SELFUPDATE=true ... video-downloader-bot
```

> Обновление best-effort: если pip не достучится до PyPI (например, из-за DPI), `yt-dlp` останется на текущей версии.

#### Impersonation в `yt-dlp`

Некоторые видео TikTok требуют *impersonation* для корректного скачивания, иначе может появиться ошибка. Такие случаи будут видны в логах:
```
2026/03/24 22:22:22 WARN yt-dlp: WARNING: [TikTok] The extractor is attempting impersonation, but no impersonate target is available. If you encounter errors, then see https://github.com/yt-dlp/yt-dlp#impersonation  for information on installing the required dependencies
```
Чтобы исправить, необходимо установить `yt-dlp` с `curl-cffi`:
```bash
uv tool install "yt-dlp[default,curl-cffi]"
```
или
```bash
pip3 install "yt-dlp[default,curl-cffi]"
```
В Homebrew этот модуль не включен в формулу, насколько я знаю

#### Куки для `yt-dlp`

Некоторые видео имеют рейтинг «для взрослых» или превышен лимит скачивания (как гость) — оба случая приведут к ошибке:
```bash
2026/03/24 22:22:22 WARN yt-dlp: ERROR: [TikTok] 761.............326: This post may not be comfortable for some audiences. Log in for access. Use --cookies-from-browser or --cookies for the authentication. See https://github.com/yt-dlp/yt-dlp/wiki/FAQ#how-do-i-pass-cookies-to-yt-dlp  for how to manually pass cookies
```
Чтобы исправить, нужно поместить файл `cookies.txt` в `./files/static/` (или отредактировать конфиг под свой путь к файлу)

Инструкции по получению файла с куки — [в wiki yt-dlp](https://github.com/yt-dlp/yt-dlp/wiki/FAQ#how-do-i-pass-cookies-to-yt-dlp)

TLDR: зарегистрировать *одноразовые* аккаунты, открыть окно инкогнито с вкладками YouTube/Instagram/TikTok, войти в каждый аккаунт, затем использовать [это расширение для браузера](https://chromewebstore.google.com/detail/get-cookiestxt-locally/cclelndahbckbenkjhflpdbgdldlbecc) для экспорта куки в TXT-файл
