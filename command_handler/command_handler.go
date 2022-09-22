// Package command_handler provides implementation
// for all handle methods in the bot.
package command_handler

import (
	"log"
	"strconv"

	lcl "github.com/dzendos/Turing/config/locales"
	gs "github.com/dzendos/Turing/game"
	tb "gopkg.in/telebot.v3"
)

type BotHandler struct {
	Bot            *tb.Bot              // Bot contains reference on a main Bot to be able to send c.Messages throygh it.
	Local          *lcl.Localizer       // Local contains dictionary with c.Messages on different languages.
	CurrentPlayers map[int64]*gs.Player // Current players contains all the players that are playing or looking for a game. Key is an id of the player.
}

// CmdStart implements action on '/start' command.// BotHandler provides an interface between bot and commands.
func (handler *BotHandler) CmdStart(c tb.Context) error {
	answer := handler.Local.Get(c.Sender().LanguageCode, "start")
	handler.Bot.Send(c.Sender(), answer)
	return nil
}

// CmdNewGame creates a new instance of a game for a current player
// (if he is not in a game) and puts this Player in currentPlayers
// as Lobby waiter.
func (handler *BotHandler) CmdNewGame(c tb.Context) error {
	_, isPlaying := handler.CurrentPlayers[c.Sender().ID]

	if isPlaying {
		answer := handler.Local.Get(c.Sender().LanguageCode, "NewGameError")
		handler.Bot.Send(c.Sender(), answer)
		return nil
	}

	answer := handler.Local.Get(c.Sender().LanguageCode, "NewGameCreation")
	handler.Bot.Send(c.Sender(), answer)

	player := gs.NewPlayer(c.Sender())
	player.State.HostId = c.Sender().ID

	handler.CurrentPlayers[c.Sender().ID] = player

	if c.Message().Text == "/new_random_game" {
		player.State.IsGameRandom = true
	}

	log.Print(c.Sender())
	return nil
}

// CmdGetMyId sends user his id in telegram
// it can be used to connect to some person's game.
func (handler *BotHandler) CmdGetMyId(c tb.Context) error {
	handler.Bot.Send(c.Sender(), strconv.FormatInt(c.Sender().ID, 10))
	return nil
}

// CmdExitLobby deletes player from lobby if the game have not started yet
// and finishes the game if it has started.
func (handler *BotHandler) CmdExitLobby(c tb.Context) error {
	player, isInGame := handler.CurrentPlayers[c.Sender().ID]

	if !isInGame {
		answer := handler.Local.Get(c.Sender().LanguageCode, "NotInLobby")
		handler.Bot.Send(c.Sender(), answer)
		return nil
	}

	player.State.NumberOfPlayers--

	var host, knight, knave *gs.Player

	// Telling others that someone left the lobby.
	for _, playerF := range handler.CurrentPlayers {
		if playerF.State == player.State {
			if playerF != player {
				answer := player.User.FirstName + handler.Local.Get(playerF.User.LanguageCode, "LeftTheLobby")
				handler.Bot.Send(playerF.User, answer)
			}

			if playerF.Role == gs.Host {
				host = playerF
			}
			if playerF.Role == gs.Knight {
				knight = playerF
			}
			if playerF.Role == gs.Knave {
				knave = playerF
			}
		}
	}

	if player.Role != gs.Lobby {
		host.State.WasGameFinished = true

		gs.PrintStatistics(
			handler.Bot,
			handler.Local,
			host,
			knight,
			knave,
			host.State,
		)

		for user, playerF := range handler.CurrentPlayers {
			if playerF.State == player.State && playerF != player {
				delete(handler.CurrentPlayers, user)
			}
		}
	}

	delete(handler.CurrentPlayers, player.User.ID)

	return nil
}

// CmdAnswer calls a c.Message with keyboard with 2 keys - names of the players
// So the host can make a decision about the personality and finish the game.
func (handler *BotHandler) CmdAnswer(c tb.Context) error {
	player, isPlaying := handler.CurrentPlayers[c.Sender().ID]

	if !isPlaying {
		answer := handler.Local.Get(c.Sender().LanguageCode, "AnswerError")
		handler.Bot.Send(c.Sender(), answer)
		return nil
	}

	if player.Role != gs.Host {
		answer := handler.Local.Get(c.Sender().LanguageCode, "NotAHostAnswer")
		handler.Bot.Send(c.Sender(), answer)
		return nil
	}

	var host, knight, knave *gs.Player

	for _, playerF := range handler.CurrentPlayers {
		if playerF.State == player.State && playerF.Role == gs.Host {
			host = playerF
		}
		if playerF.State == player.State && playerF.Role == gs.Knight {
			knight = playerF
		}
		if playerF.State == player.State && playerF.Role == gs.Knave {
			knave = playerF
		}
	}

	hostAnswer := handler.Local.Get(host.User.LanguageCode, "WhoIsKnave") + host.State.AnswerHandler.RightPlayer.NickName
	knightAnswer := handler.Local.Get(knight.User.LanguageCode, "HostMakingDecision")
	knaveAnswer := handler.Local.Get(knave.User.LanguageCode, "HostMakingDecision")
	handler.Bot.Send(host.User, hostAnswer, host.State.Selector)
	handler.Bot.Send(knight.User, knightAnswer)
	handler.Bot.Send(knave.User, knaveAnswer)

	delete(handler.CurrentPlayers, host.User.ID)
	delete(handler.CurrentPlayers, knight.User.ID)
	delete(handler.CurrentPlayers, knave.User.ID)

	return nil
}

// c.MessageHandler handles c.Messages sent by the user
// (for example during the game or while inviting people).
func (handler *BotHandler) MessageHandler(c tb.Context) error {
	p, isPlaying := handler.CurrentPlayers[c.Sender().ID]

	log.Print(p)

	// If we are not in a game (we are not playing and we have not created one).
	if !isPlaying {
		// If user wrote some c.Message in this case - it means he tries to connect to some person by his id.
		id, err := strconv.ParseInt(c.Message().Text, 10, 64)
		if err != nil {
			answer := handler.Local.Get(c.Sender().LanguageCode, "IncorrectGameId")
			handler.Bot.Send(c.Sender(), answer)
			return nil
		}

		// If we are here - it means that c.Message sent by the user is the int
		// and it can be some user id.
		doesUserExist := false
		for _, player := range handler.CurrentPlayers {
			// check if player do not play
			if player.User.ID == id {
				if player.Role != gs.Lobby {
					answer := handler.Local.Get(c.Sender().LanguageCode, "UserAlreadyInGame")
					handler.Bot.Send(c.Sender(), answer)
					return nil
				}

				if player.User.ID == c.Sender().ID {
					answer := handler.Local.Get(c.Sender().LanguageCode, "JoiningYourOwnGame")
					handler.Bot.Send(c.Sender(), answer)
					return nil
				}

				// We connect to this person.
				newPlayer := gs.NewPlayer(c.Sender())
				newPlayer.State = player.State
				handler.CurrentPlayers[c.Sender().ID] = newPlayer

				doesUserExist = true

				// Sending c.Messages to users about what happened
				joinedPlayerAnswer := handler.Local.Get(c.Sender().LanguageCode, "YouJoined") + player.User.FirstName
				hostKnavenswer := c.Sender().FirstName + handler.Local.Get(c.Sender().LanguageCode, "SomePlayerJoinedYou")
				handler.Bot.Send(c.Sender(), joinedPlayerAnswer)
				handler.Bot.Send(player.User, hostKnavenswer)

				// Changing game state
				player.State.PlayerJoined(handler.Bot, handler.Local, &handler.CurrentPlayers)

				break
			}
		}

		if !doesUserExist {
			answer := handler.Local.Get(c.Sender().LanguageCode, "UserDoesNotExist")
			handler.Bot.Send(c.Sender(), answer)
			return nil
		}

		return nil
	}

	isInLobby := handler.CurrentPlayers[c.Sender().ID].Role == gs.Lobby

	switch isInLobby {
	case true: // It means that we are waiting for others to join
		answer := handler.Local.Get(c.Sender().LanguageCode, "WaitingForOthers") +
			strconv.Itoa(handler.CurrentPlayers[c.Sender().ID].State.NumberOfPlayers)
		handler.Bot.Send(c.Sender(), answer)

	case false: // It means that we are playing and try to do some action.
		player := handler.CurrentPlayers[c.Sender().ID]

		player.State.PerformAction(player, &c.Message().Text, handler.Bot, handler.Local, &handler.CurrentPlayers)
	}

	return nil
}
