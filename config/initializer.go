// Package config provides basic functions
// to initialize bot before the start.
package config

import (
	"encoding/json"
	"log"
	"os"
	"time"

	cmd_handler "github.com/dzendos/Turing/command_handler"
	lcl "github.com/dzendos/Turing/config/locales"
	gs "github.com/dzendos/Turing/game"
	tb "gopkg.in/tucnak/telebot.v2"
)

// InitializeBot tries to connect the bot with
// our token.
func InitializeBot() (*tb.Bot, error) {
	jsonDict, _ := os.ReadFile("config/config.json")

	var configs map[string]map[string]string
	err := json.Unmarshal(jsonDict, &configs)

	if err != nil {
		log.Print(err.Error())
		return nil, err
	}

	var token string = string(configs["bot"]["token"])
	// TODO: add loading token from json
	return tb.NewBot(tb.Settings{
		Token:  token,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})
}

// InitializeBotHandler connects bot with all handle
// methods we have.
func InitializeBotHandler(bot *tb.Bot) {
	botHandler := cmd_handler.BotHandler{Bot: bot, Local: lcl.NewLocalizer(), CurrentPlayers: make(map[int64]*gs.Player)}

	bot.Handle("/start", botHandler.CmdStart)
	bot.Handle("/get_my_id", botHandler.CmdGetMyId)
	bot.Handle("/new_game", botHandler.CmdNewGame)
	bot.Handle("/exit_lobby", botHandler.CmdExitLobby)
	bot.Handle("/answer", botHandler.CmdAnswer)
	bot.Handle(tb.OnText, botHandler.MessageHandler)
}
