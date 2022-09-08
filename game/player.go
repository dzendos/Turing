package game

import (
	tb "gopkg.in/tucnak/telebot.v2"
)

// Type PlayerRole is used to identify
// the role of the player in the game.
type PlayerRole int

// In the game we have three roles that are described here.
const (
	Lobby   PlayerRole = iota + 1 // Lobby - player created his game and eaits for others to join.
	Host                          // Host is the player who asks the question.
	PlayerA                       // PlayerA is the player who tries to help to the Host.
	PlayerB                       // PlayerB is the player who tries to confuse the Host.
)

// Type Player struct contains all neccessary information
// about the player, that is needed during the game.
type Player struct {
	User  *tb.User
	Role  PlayerRole
	State *GameState
}

// NewPlayer creates new player with the role Lobby initially,
// because we create new player only during we look for a game.
func NewPlayer(user *tb.User) *Player {
	return &Player{
		User:  user,
		Role:  Lobby,
		State: NewGameState(),
	}
}
