package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/golang/glog"
	"github.com/jemoster/icfp2017/src/protocol"
)

type state struct {
	Punter  uint64
	Punters uint64
	Map     protocol.Map

	Turn uint64
}

type Punter2 struct{}

func (Punter2) Name() string {
	return "punter2"
}

func (Punter2) Setup(setup *protocol.Setup) (*protocol.Ready, error) {
	glog.Infof("Setup")

	s := state{
		Punter:  setup.Punter,
		Punters: setup.Punters,
		Map:     setup.Map,

		Turn: 0,
	}

	return &protocol.Ready{
		Ready: s.Punter,
		State: s,
	}, nil
}

func (Punter2) Play(m []protocol.Move, jsonState json.RawMessage) (*protocol.GameplayOutput, error) {
	glog.Infof("Play")

	var s state
	if err := json.Unmarshal([]byte(jsonState), &s); err != nil {
		return nil, fmt.Errorf("error unmarshaling state %s: %v", string(jsonState), err)
	}

	s.Turn++
	glog.Infof("Turn: %d", s.Turn)

	return &protocol.GameplayOutput{
		Move: protocol.Move{
			Pass: &protocol.Pass{
				s.Punter,
			},
		},
		State: s,
	}, nil
}

func (Punter2) Stop(stop *protocol.Stop, jsonState json.RawMessage) error {
	glog.Infof("Stop: %+v", stop)

	var s state
	if err := json.Unmarshal([]byte(jsonState), &s); err != nil {
		return fmt.Errorf("error unmarshaling state %s: %v", string(jsonState), err)
	}

	return nil
}

func main() {
	flag.Set("logtostderr", "true")
	flag.Parse()

	if err := protocol.Play(os.Stdin, os.Stdout, &Punter2{}); err != nil {
		glog.Exitf("Play failed: %v", err)
	}
}
