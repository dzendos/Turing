package game_state

import (
	tb "gopkg.in/tucnak/telebot.v2"
)

type PlayerRole int

const (
	Host PlayerRole = iota + 1
	PlayerA
	PlayerB
)

type Player struct {
	UserId tb.User
	Role   PlayerRole
}

type GameState struct {
	Players [3]Player
}
