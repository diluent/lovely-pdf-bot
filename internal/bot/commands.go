package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// RegisterCommands publishes the command menu: English as the default for all
// users, Russian for users whose Telegram client language is Russian.
func RegisterCommands(api *tgbotapi.BotAPI) error {
	for _, c := range []struct {
		l    lang
		code string
	}{
		{langEN, ""},
		{langRU, "ru"},
	} {
		var cmds []tgbotapi.BotCommand
		for _, d := range commandDescriptions[c.l] {
			cmds = append(cmds, tgbotapi.BotCommand{Command: d.cmd, Description: d.desc})
		}
		cfg := tgbotapi.NewSetMyCommandsWithScopeAndLanguage(tgbotapi.NewBotCommandScopeDefault(), c.code, cmds...)
		if _, err := api.Request(cfg); err != nil {
			return err
		}
	}
	return nil
}

// menuKeyboard builds the reply keyboard shown under the input field.
func menuKeyboard(l lang) tgbotapi.ReplyKeyboardMarkup {
	var rows [][]tgbotapi.KeyboardButton
	for i, b := range menuButtons[l] {
		if i%2 == 0 {
			rows = append(rows, nil)
		}
		rows[len(rows)-1] = append(rows[len(rows)-1], tgbotapi.NewKeyboardButton(b.label))
	}
	kb := tgbotapi.NewReplyKeyboard(rows...)
	kb.ResizeKeyboard = true
	return kb
}
