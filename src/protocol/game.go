package protocol

import "encoding/json"

// Game plays the game!
//
// Note that each method is called in a new process. All state must be passed
// via the State fields in the various structs (which must marshal to JSON).
type Game interface {
	// Name returns the name of the player.
	Name() string

	// Setup is called to perform the setup stage.
	//
	// Returning a non-nil error aborts the game.
	Setup(s *Setup) (*Ready, error)

	// Play is called for each step in the game.
	//
	// Returning a non-nil error aborts the game.
	Play(m []Move, state json.RawMessage) (*GameplayOutput, error)

	// Stop is called when the game is over.
	Stop(s *Stop, state json.RawMessage) error
}
