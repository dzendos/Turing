// Package game_state implements structure
// and logic of the game.
package game

import (
	"log"
	"math/rand"
	"strconv"
	"time"

	lcl "github.com/dzendos/Turing/config/locales"
	"github.com/goombaio/namegenerator"
	tb "gopkg.in/telebot.v3"
)

// printStatistics sends all the information about the game
// when the game is over.
func PrintStatistics(bot *tb.Bot, local *lcl.Localizer, host, knight, knave *Player, state *GameState) {
	hostResult := local.Get(host.User.LanguageCode, "GameOver")
	knightResult := local.Get(knight.User.LanguageCode, "GameOver")
	knaveResult := local.Get(knave.User.LanguageCode, "GameOver")

	bot.Send(host.User, hostResult)
	bot.Send(knight.User, knightResult)
	bot.Send(knave.User, knaveResult)

	begDate := host.State.BegginingDate
	gameDuration := time.Since(begDate)

	hostNumberOfMessages := local.Get(host.User.LanguageCode, "NumberOfMessages") + strconv.FormatInt(int64(len(host.History)), 10)
	hostBegginingDate := local.Get(host.User.LanguageCode, "BegginingDate") + begDate.Format(time.RFC822)
	hostGameDuration := local.Get(host.User.LanguageCode, "GameDuration") + strconv.FormatInt(int64(gameDuration.Seconds()), 10)
	hostAnswer := hostNumberOfMessages + "\n" + hostBegginingDate + "\n" + hostGameDuration

	knightNumberOfMessages := local.Get(knight.User.LanguageCode, "NumberOfMessages") + strconv.FormatInt(int64(len(knight.History)), 10)
	knightBegginingDate := local.Get(knight.User.LanguageCode, "BegginingDate") + begDate.Format(time.RFC822)
	knightGameDuration := local.Get(knight.User.LanguageCode, "GameDuration") + strconv.FormatInt(int64(gameDuration.Seconds()), 10)
	knightAnswer := knightNumberOfMessages + "\n" + knightBegginingDate + "\n" + knightGameDuration

	knaveNumberOfMessages := local.Get(knave.User.LanguageCode, "NumberOfMessages") + strconv.FormatInt(int64(len(knave.History)), 10)
	knaveBegginingDate := local.Get(knave.User.LanguageCode, "BegginingDate") + begDate.Format(time.RFC822)
	knaveGameDuration := local.Get(knave.User.LanguageCode, "GameDuration") + strconv.FormatInt(int64(gameDuration.Seconds()), 10)
	knaveAnswer := knaveNumberOfMessages + "\n" + knaveBegginingDate + "\n" + knaveGameDuration

	bot.Send(host.User, hostAnswer)
	bot.Send(knight.User, knightAnswer)
	bot.Send(knave.User, knaveAnswer)
}

type answerHandler struct {
	Bot   *tb.Bot        // Bot contains reference on a main Bot to be able to send messages throygh it.
	Local *lcl.Localizer // Local contains dictionary with messages on different languages.

	RightPlayer *Player

	host   *Player
	knave  *Player
	knight *Player
}

func (handler *answerHandler) pressHandle(c tb.Context) error {
	name := c.Callback().Unique

	if name == handler.RightPlayer.User.FirstName {
		// Win case.
		hostAnswer := handler.Local.Get(handler.host.User.LanguageCode, "YouWin")
		knightAnswer := handler.Local.Get(handler.knight.User.LanguageCode, "YouWin")
		knaveAnswer := handler.Local.Get(handler.knave.User.LanguageCode, "YouLoose")
		handler.Bot.Edit(c.Message(), hostAnswer)
		handler.Bot.Send(handler.knight.User, knightAnswer)
		handler.Bot.Send(handler.knave.User, knaveAnswer)
	} else {
		hostAnswer := handler.Local.Get(handler.host.User.LanguageCode, "YouLoose")
		knightAnswer := handler.Local.Get(handler.knight.User.LanguageCode, "YouLoose")
		knaveAnswer := handler.Local.Get(handler.knave.User.LanguageCode, "YouWin")
		handler.Bot.Edit(c.Message(), hostAnswer)
		handler.Bot.Send(handler.knight.User, knightAnswer)
		handler.Bot.Send(handler.knave.User, knaveAnswer)
	}

	handler.host.State.WasGameFinished = true
	handler.host.State.WasGameSuccesfull = true

	PrintStatistics(
		handler.Bot,
		handler.Local,
		handler.host,
		handler.knight,
		handler.knave,
		handler.host.State,
	)

	UploadGame(handler.host, handler.knight, handler.knave)

	return nil
}

func newAnswerHandler(bot *tb.Bot, local *lcl.Localizer, rightPlayer, host, knave, knight *Player) *answerHandler {
	return &answerHandler{
		bot,
		local,
		rightPlayer,
		host,
		knave,
		knight,
	}
}

// Type GameState contains all the information about
// the current game.
type GameState struct {
	HasHostFinished   bool
	HasKnaveFinished  bool
	HasKnightFinished bool
	IsHostTurn        bool
	IsGameRandom      bool

	NumberOfPlayers int

	WasGameSuccesfull bool
	WasGameFinished   bool

	HostId int64

	BegginingDate time.Time

	Selector *tb.ReplyMarkup
	Btn1     tb.Btn
	Btn2     tb.Btn

	AnswerHandler *answerHandler
}

func (gs *GameState) randomDistribution(bot *tb.Bot, local *lcl.Localizer, currentPlayers *map[int64]*Player) (*Player, *Player, *Player) {
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

	return players[0], players[1], players[2]
}

func (gs *GameState) creatorIsAHost(bot *tb.Bot, local *lcl.Localizer, currentPlayers *map[int64]*Player) (*Player, *Player, *Player) {
	var host *Player

	for _, player := range *currentPlayers {
		if player.State == gs && player.User.ID == gs.HostId {
			host = player
			break
		}
	}

	var players []*Player

	for _, player := range *currentPlayers {
		if player.State == gs && player != host {
			players = append(players, player)
		}
	}

	host.Role = Host
	players[0].Role = Knave
	players[1].Role = Knight

	return host, players[0], players[1]
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
	var host, knight, knave *Player
	if gs.IsGameRandom {
		host, knave, knight = gs.randomDistribution(bot, local, currentPlayers)
	} else {
		host, knave, knight = gs.creatorIsAHost(bot, local, currentPlayers)
	}

	knave.NickName = getRandomNickName()
	time.Sleep(8 * time.Millisecond)
	knight.NickName = getRandomNickName()

	// Sending messages
	hostAnswer := local.Get(host.User.LanguageCode, "HostGreetingMessage") + "\n" +
		knave.User.FirstName + "\n" +
		knight.User.FirstName

	KnaveAnswer := local.Get(knave.User.LanguageCode, "KnaveGreetingMessage") + knight.User.FirstName
	KnightAnswer := local.Get(knight.User.LanguageCode, "KnightGreetingMessage") + knave.User.FirstName

	bot.Send(host.User, hostAnswer)
	bot.Send(knave.User, KnaveAnswer)
	bot.Send(knight.User, KnightAnswer)

	gs.IsHostTurn = true

	// Creating buttons for a host.
	rand.Seed(time.Now().UnixNano())
	randomPlayer := rand.Intn(2)
	log.Print(randomPlayer)

	var players [2]*Player
	players[0] = knight
	players[1] = knave

	// TODO: review
	gs.Btn1 = gs.Selector.Data(knave.User.FirstName, knave.User.FirstName)
	gs.Btn2 = gs.Selector.Data(knight.User.FirstName, knight.User.FirstName)

	gs.AnswerHandler = newAnswerHandler(
		bot,
		local,
		players[randomPlayer],
		host,
		knave,
		knight,
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

	player.History = append(player.History, MessageHistory{
		*message,
		uint64(time.Since(player.State.BegginingDate).Seconds()),
	})
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
		false,
		1,
		false,
		false,
		0,
		time.Now(),
		&tb.ReplyMarkup{},
		tb.Btn{},
		tb.Btn{},
		nil,
	}
}

// shufflePlayers is used to give random roles for players.
func shufflePlayers(players []*Player) {
	rand.Seed(time.Now().UnixNano())
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
