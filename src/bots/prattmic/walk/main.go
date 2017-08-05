package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"

	"github.com/golang/glog"
	"github.com/jemoster/icfp2017/src/graph"
	"github.com/jemoster/icfp2017/src/protocol"
	gograph "gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
)

type state struct {
	Punter  uint64
	Punters uint64
	Map     protocol.Map

	Turn  uint64
	Moves []protocol.Move
}

type LongWalk struct{}

func (LongWalk) Name() string {
	return "prattmic-longwalk"
}

// furthestNode returns the mine and site that are furthest apart, and how far
// they are.
func furthestNode(g *simple.UndirectedGraph, s *state) (protocol.SiteID, protocol.SiteID, uint64, []gograph.Node) {
	var (
		mine     protocol.SiteID
		target   protocol.SiteID
		furthest uint64
		path     []gograph.Node
	)

	for _, m := range s.Map.Mines {
		shortest := graph.ShortestFrom(g, m)

		for _, site := range s.Map.Sites {
			n := g.Node(int64(site.ID))
			dist := shortest.WeightTo(n)
			if math.IsInf(dist, 0) {
				// Unreachable.
				continue
			}

			if uint64(dist) > furthest {
				mine = m
				target = site.ID
				furthest = uint64(dist)
				path, _ = shortest.To(n)
			}
		}
	}

	return mine, target, furthest, path
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
	mine, target, dist, path := furthestNode(g, s)
	glog.Infof("Furthest site: %d -> %d: %d, path: %+v", mine, target, dist, path)

	s.Moves = make([]protocol.Move, dist)
	for i := range s.Moves {
		s.Moves[i] = protocol.Move{
			Claim: &protocol.Claim{
				Punter: s.Punter,
				Source: protocol.SiteID(path[i].ID()),
				Target: protocol.SiteID(path[i+1].ID()),
			},
		}
		glog.Infof("move %d: %+v", i, s.Moves[i].Claim)
	}

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

	glog.Infof("Turn: %d", s.Turn)

	var move protocol.Move
	if s.Turn < uint64(len(s.Moves)) {
		move = s.Moves[s.Turn]
	} else {
		move = protocol.Move{
			Pass: &protocol.Pass{
				s.Punter,
			},
		}
	}
	glog.Infof("Playing: %+v", move)

	s.Turn++

	return &protocol.GameplayOutput{
		Move:  move,
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
