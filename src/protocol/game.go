package protocol

// Game plays the game!
//
// Note that each method is called in a new process. All state must be passed
// via the State fields in the various structs (which must marshal to JSON).
type Game interface {
	// Name returns the name of the player.
	Name() string

	// Setup is called to perform the setup stage.
	Setup(s *GameplayInput) *Ready

	// Play is called for each step in the game.
	Play(g *GameplayInput) *GameplayOutput

	// Stop is called when the game is over.
	Stop(s *GameplayInput)
}
