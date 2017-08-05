package main

import (
	"fmt"
	"log"
	"os"

	. "github.com/jemoster/icfp2017/src/protocol"
)

const TAG string = "Simpleton"

func dbg(s string, a ...interface{}) {
	log.Printf("[%s] %s", TAG, fmt.Sprintf(s, a...))
}

type state struct {
	Punter  uint64 `json:"punter"`
	Punters uint64 `json:"punters"`
	Map     Map    `json:"map"`

	Turn uint64 `json:"turn"`
}

type Simpleton struct {
}

func (self Simpleton) Name() string {
	return "Simpleton"
}

func (self Simpleton) Setup(g *GameplayInput) *Ready {
	dbg("Setup")

	s := state{
		Punter:  g.Punter,
		Punters: g.Punters,
		Map:     g.Map,

		Turn: 0,
	}

	return &Ready{
		s.Punter,
		EncodeState(s),
	}
}

func (self Simpleton) Play(g *GameplayInput) *GameplayOutput {
	dbg("Play")

	var state state
	DecodeState(g, &state)

	state.Turn += 1
	dbg("Turn: %d\n", state.Turn)

	return &GameplayOutput{
		Move{
			Pass: &Pass{
				state.Punter,
			},
		},
		EncodeState(state),
	}
}

func (self Simpleton) Stop(g *GameplayInput) {
	dbg("Stop", g)

	var state state
	DecodeState(g, &state)
}

func main() {
	Play(os.Stdin, os.Stdout, new(Simpleton))
}
