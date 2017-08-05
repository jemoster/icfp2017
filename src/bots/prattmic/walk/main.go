package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/golang/glog"
	"github.com/jemoster/icfp2017/src/graph"
	"github.com/jemoster/icfp2017/src/protocol"
	"gonum.org/v1/gonum/graph/simple"
)

type state struct {
	Punter  uint64
	Punters uint64
	Map     protocol.Map

	Turn uint64
}

type LongWalk struct{}

func (LongWalk) Name() string {
	return "prattmic-longwalk"
}

// furthestNode returns the mine and site that are furthest apart, and how far
// they are.
func furthestNode(g *simple.UndirectedGraph, s *state) (protocol.SiteID, protocol.SiteID, uint64) {
	distances := graph.ShortestDistances(g, s.Map.Mines)

	var mine protocol.SiteID
	var target protocol.SiteID
	var furthest uint64

	for m, sites := range distances {
		for site, dist := range sites {
			if dist > furthest {
				mine = m
				target = site
				furthest = dist
			}
		}
	}

	return mine, target, furthest
}

func (LongWalk) Setup(setup *protocol.Setup) (*protocol.Ready, error) {
	glog.Infof("Setup")

	s := &state{
		Punter:  setup.Punter,
		Punters: setup.Punters,
		Map:     setup.Map,

		Turn: 0,
	}

	g := graph.Build(&s.Map)
	mine, target, dist := furthestNode(g, s)
	glog.Infof("Furthest site: %d -> %d: %d", mine, target, dist)

	return &protocol.Ready{
		Ready: s.Punter,
		State: s,
	}, nil
}

func (LongWalk) Play(m []protocol.Move, jsonState json.RawMessage) (*protocol.GameplayOutput, error) {
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

func (LongWalk) Stop(stop *protocol.Stop, jsonState json.RawMessage) error {
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

	var s LongWalk
	if err := protocol.Play(os.Stdin, os.Stdout, &s); err != nil {
		glog.Exitf("Play failed: %v", err)
	}
}
