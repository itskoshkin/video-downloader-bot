# Video Downloader Bot

[🇬🇧](https://github.com/itskoshkin/video-downloader-bot/blob/master/README.md)/[🇷🇺](https://github.com/itskoshkin/video-downloader-bot/blob/master/README-ru.md)/🇺🇦

Telegram-бот для завантаження відео з X/Twitter, YouTube Shorts, Instagram та TikTok

- Надішли посилання боту напряму або скористайся [інлайн-режимом](https://telegram.org/blog/inline-bots)
- Бот відповість відео та посиланням без трекерів

**[@cat_video_downloader_bot](https://t.me/cat_video_downloader_bot)**

## Фічі

- 1️⃣ Надішли посилання (наприклад [youtu.be/lOwxBDDwdDU](https://youtu.be/lOwxBDDwdDU)) боту, отримай відео
- 2️⃣ Надішли посилання в будь-який чат через інлайн-режим, отримай відео
- 🔗 Підтримувані посилання:
  - <img src="https://raw.githubusercontent.com/FortAwesome/Font-Awesome/refs/heads/6.x/svgs/brands/x-twitter.svg" width="15" height="10" alt="X/Twitter"> X/Twitter (`https://x.com/<username>/status/<tweet_id>`, `twitter.com/...`, `fxtwitter.com/...`)
  - <img src="https://raw.githubusercontent.com/FortAwesome/Font-Awesome/refs/heads/6.x/svgs/brands/youtube.svg" width="10" height="10" alt="YouTube">⠀YouTube (`youtube.com/shorts/<video_id>`, `youtube.com/watch?v=...`, `youtu.be/...`)
  - <img src="https://raw.githubusercontent.com/FortAwesome/Font-Awesome/refs/heads/6.x/svgs/brands/instagram.svg" width="10" height="10" alt="Instagram">⠀Instagram (`instagram.com/reel/<reel_id>`, `ddinstagram.com/...`, `kkinstagram.com/...`)
  - <img src="https://raw.githubusercontent.com/FortAwesome/Font-Awesome/refs/heads/6.x/svgs/brands/tiktok.svg" width="10" height="10" alt="TikTok">⠀TikTok (`tiktok.com/@<username>/video/<post_id>`, `tiktok.com/t/...`, `vt/vm.tiktok.com/...`)
- ⚠️ Поточні обмеження:
  - 10 запитів на хвилину та 100 запитів на день (як у приватних повідомленнях, так і в інлайн-режимі)
  - Максимальна тривалість відео — 5 хвилин
  - Максимальний розмір файлу — 20 МБ
- ⚙️ Налаштування через команду `/settings`:
  - Мова бота (`en`/`ru`/`ua`)
  - Стиль підпису (тільки відео, відео і посилання, повний)
  - Швидкий інлайн-режим (не чекати завантаження назви)
- ❔ Доступні команди — `/start`, `/help`, `/settings`

## Збірка та запуск

### Вимоги

- [Go 1.26+](https://gist.github.com/itskoshkin/5f45fca15c30f859955dc146080a00d9)

- [PostgreSQL](https://gist.github.com/itskoshkin/0a9bb715f78f3a6133a8b5daf0ff898a)

- [yt-dlp](https://gist.github.com/itskoshkin/96b25adb0646b478b7b1980893ecd636) (для завантаження відео)

- [ffmpeg](https://gist.github.com/itskoshkin/536abbf6aae9d15a8eb396c3ed9df851) (для конвертації відео, інакше у Telegram може відображатися статичне зображення зі звуком)

- [Токен Telegram-бота](https://core.telegram.org/bots/faq) (отримати у [@BotFather](https://t.me/BotFather))

- [ID каналу-сховища](#Канал)

### Локально

1. Зібрати
   ```bash
   make build
   ```
   або
   ```bash
   go build -o video-downloader-bot ./cmd/main.go
   ```
2. Підготувати та відредагувати `config.yaml`
   ```bash
   cp example_config.yaml config.yaml && nano config.yaml
   ```
3. Запустити
   ```bash
   make run
   ```
   або
   ```bash
   ./video-downloader-bot
   ```

### Docker

1. Зібрати image
   ```bash
   make docker-build
   ```
   або
   ```bash
   docker build -t video-downloader-bot .
   ```
2. Підготувати та відредагувати `config.yaml`
   ```bash
   cp example_config.yaml config.yaml && nano config.yaml
   ```
3. Запустити контейнер
   ```bash
   make docker-run
   ```
   або
   ```bash
   docker run -d --name video-downloader-bot \
	-v $(pwd)/config.yaml:/app/config.yaml:ro \
	-v $(pwd)/files:/app/files \
	-v $(pwd)/logs:/app/logs \
	video-downloader-bot
   ```

## Різне

#### Канал

<details>

<summary><h5>Навіщо він потрібен?</h5></summary>

Інлайн-режим має жорсткий ліміт у 10 секунд на відповідь на запит
Більшість відео завантажується та конвертується довше, тому ліміт майже завжди перевищується
Як обхідне рішення бот надсилає повідомлення-заглушку з кнопкою, отримує ID повідомлення та редагує його, коли відео готове
Але редагування повідомлення для додавання медіа ([EditMessageMedia](https://core.telegram.org/bots/api#editmessagemedia)) потребує `file_id` — ідентифікатор файлу, вже завантаженого у Telegram і надісланого кудись
Тому боту потрібен канал, куди він може «скидати» відео, щоб отримати цей ID

</details>

1. Створи новий приватний канал
2. Додай до нього бота
3. Дай йому права адміністратора (публікація та видалення постів)
4. Отримай ID каналу та вкажи його в конфігу

#### Impersonation в `yt-dlp`

Деякі відео TikTok потребують *impersonation* для коректного завантаження, інакше може виникнути помилка. Такі випадки будуть видні в логах:
```
2026/03/24 22:22:22 WARN yt-dlp: WARNING: [TikTok] The extractor is attempting impersonation, but no impersonate target is available. If you encounter errors, then see https://github.com/yt-dlp/yt-dlp#impersonation  for information on installing the required dependencies
```
Щоб виправити, встанови `yt-dlp` з `curl-cffi`:
```bash
uv tool install "yt-dlp[default,curl-cffi]"
```
або
```bash
pip3 install "yt-dlp[default,curl-cffi]"
```
У Homebrew цей модуль не включений у формулу, наскільки я знаю

#### Куки для `yt-dlp`

Деякі відео призначені для дорослих або ти перевищив ліміт завантажень як гість — обидва випадки призведуть до помилки:
```bash
2026/03/24 22:22:22 WARN yt-dlp: ERROR: [TikTok] 761.............326: This post may not be comfortable for some audiences. Log in for access. Use --cookies-from-browser or --cookies for the authentication. See https://github.com/yt-dlp/yt-dlp/wiki/FAQ#how-do-i-pass-cookies-to-yt-dlp  for how to manually pass cookies
```
Щоб виправити, поклади `cookies.txt` у `./files/static/` або відредагуй конфіг під свій шлях до файлу

Інструкції щодо отримання файлу cookies — [у wiki yt-dlp](https://github.com/yt-dlp/yt-dlp/wiki/FAQ#how-do-i-pass-cookies-to-yt-dlp)

TLDR: зареєструвати *одноразові* акаунти, відкрити вікно інкогніто з вкладками YouTube/Instagram/TikTok, увійти в кожен акаунт, потім скористатися [розширенням для браузера](https://chromewebstore.google.com/detail/get-cookiestxt-locally/cclelndahbckbenkjhflpdbgdldlbecc) для експорту куків у TXT-файл
