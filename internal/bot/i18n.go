package bot

import (
	"fmt"
	"strings"

	"lovely-pdf-bot/internal/pdf"
)

type lang string

const (
	langRU lang = "ru"
	langEN lang = "en"
)

type msgKey int

const (
	msgStart msgKey = iota
	msgDownloadFailed
	msgAccepted
	msgNoImages
	msgGenerating
	msgGenerateFailed
	msgSendFailed
	msgCancelled
	msgLangSwitched
	msgLangUsage
	msgOrientPortrait
	msgOrientLandscape
	msgOrientUsage
)

var messages = map[lang]map[msgKey]string{
	langRU: {
		msgStart:           "Привет! Отправляй картинки по одной, затем /done — получишь PDF.\n/cancel — сбросить очередь.\n/orientation — книжная/альбомная.\n/lang — switch to English.",
		msgDownloadFailed:  "Не удалось скачать файл: %v",
		msgAccepted:        "Принято (%d шт.). Продолжай или /done.",
		msgNoImages:        "Нет картинок. Сначала пришли хотя бы одну.",
		msgGenerating:      "Генерирую PDF из %d картинок…",
		msgGenerateFailed:  "Ошибка генерации PDF: %v",
		msgSendFailed:      "Не удалось отправить PDF: %v",
		msgCancelled:       "Очередь сброшена, удалено %d картинок.",
		msgLangSwitched:    "Язык переключён на русский.",
		msgLangUsage:       "Использование: /lang, /lang ru или /lang en.",
		msgOrientPortrait:  "Ориентация PDF: книжная.",
		msgOrientLandscape: "Ориентация PDF: альбомная.",
		msgOrientUsage:     "Использование: /orientation, /orientation portrait или /orientation landscape.",
	},
	langEN: {
		msgStart:           "Hi! Send images one by one, then /done to get a PDF.\n/cancel — clear the queue.\n/orientation — portrait/landscape.\n/lang — переключить на русский.",
		msgDownloadFailed:  "Failed to download file: %v",
		msgAccepted:        "Got it (%d total). Keep going or /done.",
		msgNoImages:        "No images yet. Send at least one first.",
		msgGenerating:      "Generating PDF from %d images…",
		msgGenerateFailed:  "PDF generation failed: %v",
		msgSendFailed:      "Failed to send PDF: %v",
		msgCancelled:       "Queue cleared, %d images removed.",
		msgLangSwitched:    "Language switched to English.",
		msgLangUsage:       "Usage: /lang, /lang ru or /lang en.",
		msgOrientPortrait:  "PDF orientation: portrait.",
		msgOrientLandscape: "PDF orientation: landscape.",
		msgOrientUsage:     "Usage: /orientation, /orientation portrait or /orientation landscape.",
	},
}

func (l lang) t(key msgKey, args ...any) string {
	return fmt.Sprintf(messages[l][key], args...)
}

// defaultLang picks the initial language from the user's Telegram client language.
func defaultLang(code string) lang {
	if strings.HasPrefix(code, "ru") {
		return langRU
	}
	return langEN
}

// parseLang returns the language for a /lang argument, ok=false if unknown.
func parseLang(s string) (lang, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "ru", "рус", "русский", "russian":
		return langRU, true
	case "en", "eng", "english", "английский":
		return langEN, true
	}
	return "", false
}

// commandDescriptions are shown in the Telegram command menu.
var commandDescriptions = map[lang][]struct{ cmd, desc string }{
	langRU: {
		{"start", "Как пользоваться ботом"},
		{"done", "Собрать PDF из присланных картинок"},
		{"cancel", "Сбросить очередь картинок"},
		{"orientation", "Книжная / альбомная ориентация"},
		{"lang", "Switch to English"},
	},
	langEN: {
		{"start", "How to use the bot"},
		{"done", "Build a PDF from the sent images"},
		{"cancel", "Clear the image queue"},
		{"orientation", "Portrait / landscape orientation"},
		{"lang", "Переключить на русский"},
	},
}

// menuButtons are the reply keyboard labels, mapped to the command each one runs.
var menuButtons = map[lang][]struct{ label, cmd string }{
	langRU: {
		{"📄 Готово", "done"},
		{"🗑 Сбросить", "cancel"},
		{"🔄 Ориентация", "orientation"},
		{"🌐 English", "lang"},
	},
	langEN: {
		{"📄 Done", "done"},
		{"🗑 Clear", "cancel"},
		{"🔄 Orientation", "orientation"},
		{"🌐 Русский", "lang"},
	},
}

// buttonCommand returns the command for a menu button label, empty if the text
// is not a button. All languages are checked, since the user may still have an
// old keyboard on screen.
func buttonCommand(text string) string {
	for _, btns := range menuButtons {
		for _, b := range btns {
			if b.label == text {
				return b.cmd
			}
		}
	}
	return ""
}

// parseOrientation returns the orientation for an /orientation argument,
// ok=false if unknown.
func parseOrientation(s string) (pdf.Orientation, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "portrait", "p", "книжная", "книжный":
		return pdf.Portrait, true
	case "landscape", "album", "l", "альбомная", "альбомный":
		return pdf.Landscape, true
	}
	return 0, false
}

func orientationMsg(o pdf.Orientation) msgKey {
	if o == pdf.Landscape {
		return msgOrientLandscape
	}
	return msgOrientPortrait
}
