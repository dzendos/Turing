// Package command_handler provides implementation
// for all handle methods in the bot.
package command_handler

import (
	lcl "github.com/dzendos/Turing/config/locales"
	tb "gopkg.in/tucnak/telebot.v2"
)

// CmdStart implements action on '/start' command.// BotHandler provides an interface between bot and commands.
type BotHandler struct {
	Bot   *tb.Bot
	Local *lcl.Localizer
}

func (handler *BotHandler) CmdStart(message *tb.Message) {
	answer := handler.Local.Get(message.Sender.LanguageCode, "start")
	handler.Bot.Send(message.Sender, answer)
}
