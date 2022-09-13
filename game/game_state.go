// Package game_state implements structure
// and logic of the game.
package game

import (
	"math/rand"
	"time"

	lcl "github.com/dzendos/Turing/config/locales"
	"github.com/goombaio/namegenerator"
	tb "gopkg.in/tucnak/telebot.v2"
)

type answerHandler struct {
	Bot   *tb.Bot        // Bot contains reference on a main Bot to be able to send messages throygh it.
	Local *lcl.Localizer // Local contains dictionary with messages on different languages.

	RightPlayer *Player

	host   *Player
	knight *Player
	knave  *Player
}

func (handler *answerHandler) pressHandle(c *tb.Callback) {
	if c.Message.Text == handler.RightPlayer.NickName {
		// Win case.
		handler.Bot.Send(handler.host.User, "Tadam!")
	} else {
		handler.Bot.Send(handler.host.User, "Oops!")
	}
}

func newAnswerHandler(bot *tb.Bot, local *lcl.Localizer, rightPlayer, host, knight, knave *Player) *answerHandler {
	return &answerHandler{
		bot,
		local,
		rightPlayer,
		host,
		knight,
		knave,
	}
}

// Type GameState contains all the information about
// the current game.
type GameState struct {
	HasHostFinished   bool
	HasKnaveFinished  bool
	HasKnightFinished bool
	IsHostTurn        bool

	NumberOfPlayers int

	WasGameSuccesfull bool
	WasGameFinished   bool

	BegginingDate time.Time

	Selector *tb.ReplyMarkup
	Btn1     tb.Btn
	Btn2     tb.Btn

	AnswerHandler *answerHandler
}

// PlayerJoined changes the state of the current game
// (increases the number of players in the game and
// if all the players have already connected -> starts the game)
func (gs *GameState) PlayerJoined(bot *tb.Bot, local *lcl.Localizer, currentPlayers *map[int64]*Player) {
	gs.NumberOfPlayers++

	if gs.NumberOfPlayers != 3 {
		return
	}

	// Then we need to change state of people to DistributingRoles state.
	var players []*Player

	for _, player := range *currentPlayers {
		if player.State == gs {
			players = append(players, player)
		}
	}

	shufflePlayers(players)

	players[0].Role = Host
	players[1].Role = Knave
	players[2].Role = Knight

	players[1].NickName = getRandomNickName()
	time.Sleep(8 * time.Millisecond)
	players[2].NickName = getRandomNickName()

	// Sending messages
	hostAnswer := local.Get(players[0].User.LanguageCode, "HostGreetingMessage") + "\n" +
		players[1].User.Username + "\n" +
		players[2].User.Username

	KnaveAnswer := local.Get(players[1].User.LanguageCode, "KnaveGreetingMessage") + players[2].User.Username
	KnightAnswer := local.Get(players[2].User.LanguageCode, "KnightGreetingMessage") + players[1].User.Username

	bot.Send(players[0].User, hostAnswer)
	bot.Send(players[1].User, KnaveAnswer)
	bot.Send(players[2].User, KnightAnswer)

	gs.IsHostTurn = true

	// Creating buttons for a host.
	randomPlayer := rand.Intn(1) + 1

	player2 := 3 - randomPlayer

	gs.Btn1 = gs.Selector.Data(players[1].User.Username, players[randomPlayer].User.Username)
	gs.Btn2 = gs.Selector.Data(players[2].User.Username, players[player2].User.Username)

	gs.AnswerHandler = newAnswerHandler(
		bot,
		local,
		players[randomPlayer],
		players[0],
		players[1],
		players[2],
	)

	gs.Selector.Inline(
		gs.Selector.Row(gs.Btn1, gs.Btn2),
	)

	bot.Handle(&gs.Btn1, gs.AnswerHandler.pressHandle)
	bot.Handle(&gs.Btn2, gs.AnswerHandler.pressHandle)
}

// Perform action checks if player can do some action on the current
// state of the game, and if yes - changes the state of the game.
func (gs *GameState) PerformAction(player *Player, message *string, bot *tb.Bot, local *lcl.Localizer,
	currentPlayers *map[int64]*Player) {

	if !player.CanPerformAction() {
		answer := local.Get(player.User.LanguageCode, "NotYourTurn")
		bot.Send(player.User, answer)
		return
	}

	var host, knight, knave *Player

	for _, playerF := range *currentPlayers {
		if playerF.Role == Host {
			host = playerF
		}

		if playerF.Role == Knight {
			knight = playerF
		}

		if playerF.Role == Knave {
			knave = playerF
		}
	}

	if player.Role == Host {
		toKnave := local.Get(knave.User.LanguageCode, "host") + ":\n" + *message
		toKnight := local.Get(knight.User.LanguageCode, "host") + ":\n" + *message
		bot.Send(knave.User, toKnave)
		bot.Send(knight.User, toKnight)

		player.State.HasHostFinished = true
		player.State.HasKnaveFinished = false
		player.State.HasKnightFinished = false
		player.State.IsHostTurn = false

		toKnave = local.Get(knave.User.LanguageCode, "YourTurn")
		toKnight = local.Get(knight.User.LanguageCode, "YourTurn")
		bot.Send(knave.User, toKnave)
		bot.Send(knight.User, toKnight)
	} else {
		if player.Role == Knight {
			player.State.HasKnightFinished = true
			playerMessage := knight.NickName + ":\n" + *message
			bot.Send(host.User, playerMessage)
		}
		if player.Role == Knave {
			player.State.HasKnaveFinished = true
			playerMessage := knave.NickName + ":\n" + *message
			bot.Send(host.User, playerMessage)
		}

		if player.State.HasKnightFinished && player.State.HasKnaveFinished {
			player.State.HasHostFinished = false
			player.State.IsHostTurn = true

			toHost := local.Get(host.User.LanguageCode, "YourTurn")
			bot.Send(host.User, toHost)
		}
	}
}

// NewGameState creates new empty game state.
// It is performing only when some user creates a game,
// that is why number of users by default is 1.
func NewGameState() *GameState {
	return &GameState{
		false,
		false,
		false,
		false,
		1,
		false,
		false,
		time.Now(),
		&tb.ReplyMarkup{},
		tb.Btn{},
		tb.Btn{},
		nil,
	}
}

// shufflePlayers is used to give random roles for players.
func shufflePlayers(players []*Player) {
	for i := range players {
		j := rand.Intn(i + 1)
		players[i], players[j] = players[j], players[i]
	}
}

// getRandomNickName generates nickname for player in order to hide his real name from the host
func getRandomNickName() string {
	seed := time.Now().UTC().UnixNano()
	nameGenerator := namegenerator.NewNameGenerator(seed)

	return nameGenerator.Generate()
}
