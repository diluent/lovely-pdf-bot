package bot

import (
	"fmt"
	"io"
	"net/http"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"lovely-pdf-bot/internal/pdf"
)

type session struct {
	mu     sync.Mutex
	images [][]byte
}

type Handler struct {
	bot        *tgbotapi.BotAPI
	sessionsMu sync.RWMutex
	sessions   map[int64]*session
}

func NewHandler(bot *tgbotapi.BotAPI) *Handler {
	return &Handler{
		bot:      bot,
		sessions: make(map[int64]*session),
	}
}

func (h *Handler) Handle(msg *tgbotapi.Message) {
	sess := h.getSession(msg.From.ID)
	chatID := msg.Chat.ID

	if msg.IsCommand() {
		h.handleCommand(msg.Command(), chatID, sess)
		return
	}

	fileID := extractImageFileID(msg)
	if fileID == "" {
		return
	}

	data, err := h.downloadFile(fileID)
	if err != nil {
		h.send(chatID, "Не удалось скачать файл: "+err.Error())
		return
	}

	sess.mu.Lock()
	sess.images = append(sess.images, data)
	n := len(sess.images)
	sess.mu.Unlock()

	h.send(chatID, fmt.Sprintf("Принято (%d шт.). Продолжай или /done.", n))
}

func (h *Handler) handleCommand(cmd string, chatID int64, sess *session) {
	switch cmd {
	case "start":
		h.send(chatID, "Привет! Отправляй картинки по одной, затем /done — получишь PDF.\n/cancel — сбросить очередь.")

	case "done":
		sess.mu.Lock()
		imgs := sess.images
		sess.images = nil
		sess.mu.Unlock()

		if len(imgs) == 0 {
			h.send(chatID, "Нет картинок. Сначала пришли хотя бы одну.")
			return
		}

		h.send(chatID, fmt.Sprintf("Генерирую PDF из %d картинок…", len(imgs)))

		pdfBytes, err := pdf.Generate(imgs)
		if err != nil {
			h.send(chatID, "Ошибка генерации PDF: "+err.Error())
			return
		}

		doc := tgbotapi.NewDocument(chatID, tgbotapi.FileBytes{
			Name:  "photos.pdf",
			Bytes: pdfBytes,
		})
		if _, err := h.bot.Send(doc); err != nil {
			h.send(chatID, "Не удалось отправить PDF: "+err.Error())
		}

	case "cancel":
		sess.mu.Lock()
		n := len(sess.images)
		sess.images = nil
		sess.mu.Unlock()
		h.send(chatID, fmt.Sprintf("Очередь сброшена, удалено %d картинок.", n))
	}
}

func (h *Handler) getSession(uid int64) *session {
	h.sessionsMu.Lock()
	defer h.sessionsMu.Unlock()
	if s, ok := h.sessions[uid]; ok {
		return s
	}
	s := &session{}
	h.sessions[uid] = s
	return s
}

func (h *Handler) downloadFile(fileID string) ([]byte, error) {
	f, err := h.bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		return nil, err
	}
	resp, err := http.Get(f.Link(h.bot.Token))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func (h *Handler) send(chatID int64, text string) {
	_, _ = h.bot.Send(tgbotapi.NewMessage(chatID, text))
}

// extractImageFileID returns the file ID for a photo or image document,
// empty string if the message contains neither.
func extractImageFileID(msg *tgbotapi.Message) string {
	if msg.Photo != nil {
		// Telegram provides multiple sizes; last one is the largest.
		return msg.Photo[len(msg.Photo)-1].FileID
	}
	if msg.Document != nil {
		switch msg.Document.MimeType {
		case "image/jpeg", "image/png", "image/gif":
			return msg.Document.FileID
		}
	}
	return ""
}
