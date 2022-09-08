package game_state
<<<<<<< HEAD

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
=======
>>>>>>> 960281ef16d1fbcf83d0bb64d64338666352b71a
