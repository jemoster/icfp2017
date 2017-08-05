package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/jemoster/icfp2017/src/protocol"
)

const TAG string = "Simpleton"

func dbg(s string, a ...interface{}) {
	log.Printf("[%s] %s", TAG, fmt.Sprintf(s, a...))
}

type state struct {
	Punter  uint64
	Punters uint64
	Map     protocol.Map

	Turn uint64
}

type Simpleton struct{}

func (Simpleton) Name() string {
	return "Simpleton"
}

func (Simpleton) Setup(g *protocol.Setup) (*protocol.Ready, error) {
	dbg("Setup")

	s := state{
		Punter:  g.Punter,
		Punters: g.Punters,
		Map:     g.Map,

		Turn: 0,
	}

	return &protocol.Ready{
		Ready: s.Punter,
		State: s,
	}, nil
}

func (Simpleton) Play(m []protocol.Move, jsonState json.RawMessage) (*protocol.GameplayOutput, error) {
	dbg("Play")

	var s state
	if err := json.Unmarshal([]byte(jsonState), &s); err != nil {
		return nil, fmt.Errorf("error unmarshaling state %s: %v", string(jsonState), err)
	}

	s.Turn++
	dbg("Turn: %d", s.Turn)

	return &protocol.GameplayOutput{
		Move: protocol.Move{
			Pass: &protocol.Pass{
				s.Punter,
			},
		},
		State: s,
	}, nil
}

func (Simpleton) Stop(stop *protocol.Stop, jsonState json.RawMessage) error {
	dbg("Stop: %+v", stop)

	var s state
	if err := json.Unmarshal([]byte(jsonState), &s); err != nil {
		return fmt.Errorf("error unmarshaling state %s: %v", string(jsonState), err)
	}

	return nil
}

func main() {
	var s Simpleton
	if err := protocol.Play(os.Stdin, os.Stdout, &s); err != nil {
		log.Fatalf("Play failed: %v", err)
	}
}
