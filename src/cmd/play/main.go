package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jemoster/icfp2017/src/protocol"
)

type state struct {
	// Our punter id.
	Punter uint64
}

// rebuildState recreates the state struct from what comes out of the json
// unmarshalling.
func rebuildState(v interface{}) *state {
	// """
	// To unmarshal JSON into an interface value, Unmarshal stores one of
	// these in the interface value:
	//
	// ...
	//
	// map[string]interface{}, for JSON objects
	// """ - encoding/json
	s, ok := v.(map[string]interface{})
	if !ok {
		panic(fmt.Sprintf("state is not map to interface (%T): %+v", v, v))
	}

	punter, ok := s["Punter"]
	if !ok {
		panic(fmt.Sprintf("punter missing: %+v", v))
	}

	punterFloat, ok := s["Punter"].(float64)
	if !ok {
		panic(fmt.Sprintf("punter isn't a number (%T): %+v", punter, punter))
	}

	return &state{
		Punter: uint64(punterFloat),
	}
}

// game implements protocol.Game.
type game struct{}

func (g *game) Name() string {
	return "passbot"
}

func (g *game) Setup(s *protocol.Setup) *protocol.Ready {
	log.Printf("Setup: %+v", s)

	// Always ready.
	return &protocol.Ready{
		Ready: s.Punter,
		State: &state{
			Punter: s.Punter,
		},
	}
}

func (g *game) Play(gi *protocol.GameplayInput) *protocol.GameplayOutput {
	log.Printf("Play: %+v", gi)

	s := rebuildState(gi.State)

	return &protocol.GameplayOutput{
		// Just pass.
		Move: protocol.Move{
			Pass: &protocol.Pass{
				Punter: s.Punter,
			},
		},
		State: s,
	}
}

func (g *game) Stop(s *protocol.StopInput) {
	log.Printf("Stop: %+v", s)
}

func main() {
	var g game
	if err := protocol.Play(os.Stdin, os.Stdout, &g); err != nil {
		log.Fatalf("Play failed: %v", err)
	}
}
