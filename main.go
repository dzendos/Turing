package main

import (
	"log"

	"github.com/dzendos/Turing/config"
)

func main() {
	bot, err := config.InitializeBot()

	if err != nil {
		log.Fatal(err)
		return
	}

	config.InitializeBotHandler(bot)

	bot.Start()
}
