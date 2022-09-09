// Package command_handler provides implementation
// for all handle methods in the bot.
package command_handler

import (
	"log"
	"strconv"

	lcl "github.com/dzendos/Turing/config/locales"
	gs "github.com/dzendos/Turing/game"
	tb "gopkg.in/tucnak/telebot.v2"
)

type BotHandler struct {
	Bot            *tb.Bot              // Bot contains reference on a main Bot to be able to send messages throygh it.
	Local          *lcl.Localizer       // Local contains dictionary with messages on different languages.
	CurrentPlayers map[int64]*gs.Player // Current players contains all the players that are playing or looking for a game. Key is an id of the player.
}

// CmdStart implements action on '/start' command.// BotHandler provides an interface between bot and commands.
func (handler *BotHandler) CmdStart(message *tb.Message) {
	answer := handler.Local.Get(message.Sender.LanguageCode, "start")
	handler.Bot.Send(message.Sender, answer)
}

// CmdNewGame creates a new instance of a game for a current player
// (if he is not in a game) and puts this Player in currentPlayers
// as Lobby waiter.
func (handler *BotHandler) CmdNewGame(message *tb.Message) {
	_, isPlaying := handler.CurrentPlayers[message.Sender.ID]

	if isPlaying {
		answer := handler.Local.Get(message.Sender.LanguageCode, "NewGameError")
		handler.Bot.Send(message.Sender, answer)
		return
	}

	answer := handler.Local.Get(message.Sender.LanguageCode, "NewGameCreation")
	handler.Bot.Send(message.Sender, answer)

	player := gs.NewPlayer(message.Sender)

	handler.CurrentPlayers[message.Sender.ID] = player

	log.Print(message.Sender)
}

// CmdGetMyId sends user his id in telegram
// it can be used to connect to some person's game.
func (handler *BotHandler) CmdGetMyId(message *tb.Message) {
	handler.Bot.Send(message.Sender, strconv.FormatInt(message.Sender.ID, 10))
}

// CmdExitLobby deletes player from lobby if the game have not started yet
// and finishes the game if it has started.
func (handler *BotHandler) CmdExitLobby(message *tb.Message) {
	player, isInGame := handler.CurrentPlayers[message.Sender.ID]

	if !isInGame {
		answer := handler.Local.Get(message.Sender.LanguageCode, "NotInLobby")
		handler.Bot.Send(message.Sender, answer)
		return
	}

	player.State.NumberOfPlayers--

	// Telling others that someone left the lobby.
	for _, playerF := range handler.CurrentPlayers {
		if playerF.State == player.State && playerF != player {
			answer := player.User.Username + handler.Local.Get(playerF.User.LanguageCode, "LeftTheLobby")
			handler.Bot.Send(playerF.User, answer)
		}
	}

	if player.Role != gs.Lobby {
		printStatistics(handler, player.State)

		for user, playerF := range handler.CurrentPlayers {
			if playerF.State == player.State && playerF != player {
				delete(handler.CurrentPlayers, user)
			}
		}
	}

	delete(handler.CurrentPlayers, player.User.ID)
}

// MessageHandler handles essages sent by the user
// (for example during the game or while inviting people).
func (handler *BotHandler) MessageHandler(message *tb.Message) {
	p, isPlaying := handler.CurrentPlayers[message.Sender.ID]

	log.Print(p)

	// If we are not in a game (we are not playing and we have not created one).
	if !isPlaying {
		// If user wrote some message in this case - it means he tries to connect to some person by his id.
		id, err := strconv.ParseInt(message.Text, 10, 64)
		if err != nil {
			answer := handler.Local.Get(message.Sender.LanguageCode, "IncorrectGameId")
			handler.Bot.Send(message.Sender, answer)
			return
		}

		// If we are here - it means that message sent by the user is the int
		// and it can be some user id.
		doesUserExist := false
		for _, player := range handler.CurrentPlayers {
			// check if player do not play
			if player.User.ID == id {
				if player.Role != gs.Lobby {
					answer := handler.Local.Get(message.Sender.LanguageCode, "UserAlreadyInGame")
					handler.Bot.Send(message.Sender, answer)
					return
				}

				// We connect to this person.
				newPlayer := gs.NewPlayer(message.Sender)
				newPlayer.State = player.State
				handler.CurrentPlayers[message.Sender.ID] = newPlayer

				doesUserExist = true

				// Sending messages to users about what happened
				joinedPlayerAnswer := handler.Local.Get(message.Sender.LanguageCode, "YouJoined") + player.User.Username
				hostKnavenswer := message.Sender.Username + handler.Local.Get(message.Sender.LanguageCode, "SomePlayerJoinedYou")
				handler.Bot.Send(message.Sender, joinedPlayerAnswer)
				handler.Bot.Send(player.User, hostKnavenswer)

				// Changing game state
				player.State.PlayerJoined(handler.Bot, handler.Local, &handler.CurrentPlayers)

				break
			}
		}

		if !doesUserExist {
			answer := handler.Local.Get(message.Sender.LanguageCode, "UserDoesNotExist")
			handler.Bot.Send(message.Sender, answer)
			return
		}

		return
	}

	isInLobby := handler.CurrentPlayers[message.Sender.ID].Role == gs.Lobby

	switch isInLobby {
	case true: // It means that we are waiting for others to join
		answer := handler.Local.Get(message.Sender.LanguageCode, "WaitingForOthers") +
			strconv.Itoa(handler.CurrentPlayers[message.Sender.ID].State.NumberOfPlayers)
		handler.Bot.Send(message.Sender, answer)

	case false: // It means that we are playing and try to do some action.
		player := handler.CurrentPlayers[message.Sender.ID]

		player.State.PerformAction(player, &message.Text, handler.Bot, handler.Local, &handler.CurrentPlayers)
	}
}

// printStatistics sends all the information about the game
// when the game is over.
func printStatistics(handler *BotHandler, state *gs.GameState) {
	for _, playerF := range handler.CurrentPlayers {
		if playerF.State == state {
			// Printing stat of a match.
			result := handler.Local.Get(playerF.User.LanguageCode, "GameOver")
			handler.Bot.Send(playerF.User, result)
		}
	}
}
