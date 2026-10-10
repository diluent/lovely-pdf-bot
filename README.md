# lovely-pdf-bot

Telegram bot that collects images and packs them into a PDF — one image per A4 page, scaled to fit with aspect ratio preserved.

## Usage

1. Send images one by one
2. `/done` — get the PDF
3. `/cancel` — clear the queue
4. `/lang` — toggle between Russian and English (or `/lang ru` / `/lang en`)

The initial language follows the user's Telegram client language (Russian for `ru`, English otherwise).

## Run

```bash
# fill in your token
echo "TELEGRAM_BOT_TOKEN=your_token_here" > .env

go run .
```

## Build

```bash
go build -o lovely-pdf-bot .

./lovely-pdf-bot
```

Cross-compile for Linux:

```bash
GOOS=linux GOARCH=amd64 go build -o lovely-pdf-bot .
```

## Stack

- [go-telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api)
- [gofpdf](https://github.com/jung-kurt/gofpdf)
