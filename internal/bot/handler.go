package bot

import (
	"io"
	"net/http"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"lovely-pdf-bot/internal/pdf"
)

type session struct {
	mu     sync.Mutex
	images [][]byte
	lang   lang
	orient pdf.Orientation
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
	sess := h.getSession(msg.From.ID, msg.From.LanguageCode)
	chatID := msg.Chat.ID

	if msg.IsCommand() {
		h.handleCommand(msg.Command(), msg.CommandArguments(), chatID, sess)
		return
	}
	if cmd := buttonCommand(msg.Text); cmd != "" {
		h.handleCommand(cmd, "", chatID, sess)
		return
	}

	fileID := extractImageFileID(msg)
	if fileID == "" {
		return
	}

	data, err := h.downloadFile(fileID)
	if err != nil {
		h.send(chatID, sess.getLang().t(msgDownloadFailed, err))
		return
	}

	sess.mu.Lock()
	sess.images = append(sess.images, data)
	n := len(sess.images)
	l := sess.lang
	sess.mu.Unlock()

	h.send(chatID, l.t(msgAccepted, n))
}

func (h *Handler) handleCommand(cmd, args string, chatID int64, sess *session) {
	l := sess.getLang()

	switch cmd {
	case "start", "help":
		h.sendWithMenu(chatID, l.t(msgStart), l)

	case "lang":
		next := langEN
		if l == langEN {
			next = langRU
		}
		if args != "" {
			var ok bool
			if next, ok = parseLang(args); !ok {
				h.send(chatID, l.t(msgLangUsage))
				return
			}
		}
		sess.mu.Lock()
		sess.lang = next
		sess.mu.Unlock()
		h.sendWithMenu(chatID, next.t(msgLangSwitched), next)

	case "done":
		sess.mu.Lock()
		imgs := sess.images
		sess.images = nil
		orient := sess.orient
		sess.mu.Unlock()

		if len(imgs) == 0 {
			h.send(chatID, l.t(msgNoImages))
			return
		}

		h.send(chatID, l.t(msgGenerating, len(imgs)))

		pdfBytes, err := pdf.Generate(imgs, orient)
		if err != nil {
			h.send(chatID, l.t(msgGenerateFailed, err))
			return
		}

		doc := tgbotapi.NewDocument(chatID, tgbotapi.FileBytes{
			Name:  "photos.pdf",
			Bytes: pdfBytes,
		})
		if _, err := h.bot.Send(doc); err != nil {
			h.send(chatID, l.t(msgSendFailed, err))
		}

	case "orientation":
		sess.mu.Lock()
		next := pdf.Landscape
		if sess.orient == pdf.Landscape {
			next = pdf.Portrait
		}
		if args != "" {
			var ok bool
			if next, ok = parseOrientation(args); !ok {
				sess.mu.Unlock()
				h.send(chatID, l.t(msgOrientUsage))
				return
			}
		}
		sess.orient = next
		sess.mu.Unlock()
		h.send(chatID, l.t(orientationMsg(next)))

	case "cancel":
		sess.mu.Lock()
		n := len(sess.images)
		sess.images = nil
		sess.mu.Unlock()
		h.send(chatID, l.t(msgCancelled, n))
	}
}

func (h *Handler) getSession(uid int64, langCode string) *session {
	h.sessionsMu.Lock()
	defer h.sessionsMu.Unlock()
	if s, ok := h.sessions[uid]; ok {
		return s
	}
	s := &session{lang: defaultLang(langCode)}
	h.sessions[uid] = s
	return s
}

func (s *session) getLang() lang {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lang
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

// sendWithMenu sends text and (re)attaches the menu keyboard in the given language.
func (h *Handler) sendWithMenu(chatID int64, text string, l lang) {
	m := tgbotapi.NewMessage(chatID, text)
	m.ReplyMarkup = menuKeyboard(l)
	_, _ = h.bot.Send(m)
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
