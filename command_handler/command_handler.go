// Package command_handler provides implementation
// for all handle methods in the bot.
package command_handler

import (
	tb "gopkg.in/tucnak/telebot.v2"
)

// BotHandler provides an interface between bot and commands.
type BotHandler struct {
	*tb.Bot
}

// CmdStart implements action on '/start' command.
func (bot *BotHandler) CmdStart(message *tb.Message) {

}
