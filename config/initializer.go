// Package config provides basic functions
// to initialize bot before the start.
package config

import (
	"time"

	cmd_handler "github.com/dzendos/Turing/command_handler"
	lcl "github.com/dzendos/Turing/config/locales"
	tb "gopkg.in/tucnak/telebot.v2"
)

// ########################## MOVE TO CONFIG ##########################
var TOKEN string = ""

// InitializeBot tries to connect the bot with
// our token.
func InitializeBot() (*tb.Bot, error) {
	// TODO: add loading token from json
	return tb.NewBot(tb.Settings{
		Token:  TOKEN,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
}

// InitializeBotHandler connects bot with all handle
// methods we have.
func InitializeBotHandler(bot *tb.Bot) {
	// TODO: add handlers
	botHandler := cmd_handler.BotHandler{Bot: bot, Local: lcl.NewLocalizer()}

	bot.Handle("/start", botHandler.CmdStart)
}
