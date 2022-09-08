// Package game_state implements structure
// and logic of the game.
package game

// Type GameState contains all the information about
// the current game.
type GameState struct {
	HasHostFinished    bool
	HasPlayerAFinished bool
	HasPlayerBFinished bool
	IsHostTurn         bool
}

// NewGameState creates new empty game state.
func NewGameState() *GameState {
	return &GameState{
		false,
		false,
		false,
		false,
	}
}
