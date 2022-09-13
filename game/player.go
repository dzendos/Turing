package game

import (
	tb "gopkg.in/tucnak/telebot.v2"
)

type messageHistory struct {
	message        string
	timeFromTheBeg uint64
}

// Type PlayerRole is used to identify
// the role of the player in the game.
type PlayerRole int

// In the game we have three roles that are described here.
const (
	Lobby             PlayerRole = iota + 1 // Lobby - player created his game and eaits for others to join.
	DistributingRoles                       // DistributingRoles - state of the player when roles are assigning.
	Host                                    // Host is the player who asks the question.
	Knave                                   // Knave is the player who tries to confuse the Host.
	Knight                                  // Knight is the player who tries to help to the Host.
)

// Type Player struct contains all neccessary information
// about the player, that is needed during the game.
type Player struct {
	User     *tb.User
	Role     PlayerRole
	NickName string
	State    *GameState

	history []messageHistory
}

// CanPerformAction checks if the player with his role can
// perform some action at current state of the game.
func (player *Player) CanPerformAction() bool {
	switch player.Role {
	case Host:
		return player.State.IsHostTurn && !player.State.HasHostFinished
	case Knave:
		return !player.State.IsHostTurn && !player.State.HasKnaveFinished
	case Knight:
		return !player.State.IsHostTurn && !player.State.HasKnightFinished
	}

	return false
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
