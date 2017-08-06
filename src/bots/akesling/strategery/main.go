package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"os"

	"github.com/golang/glog"
	"github.com/jemoster/icfp2017/src/graph"
	"github.com/jemoster/icfp2017/src/protocol"
	"gonum.org/v1/gonum/graph/simple"
)

func ShuffleRivers(r []protocol.River) {
	for i := len(r) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		r[i], r[j] = r[j], r[i]
	}
}

type state struct {
	Punter              uint64
	Punters             uint64
	Map                 protocol.Map
	ActivePaths         [][]protocol.Site
	UnconnectedOrigins  []protocol.River
	AvailableMineRivers []protocol.River

	Turn uint64
}

func InitializeState(setup *protocol.Setup) *state {
	return &state{
		Punter:  setup.Punter,
		Punters: setup.Punters,
		Map:     setup.Map,

		Turn: 0,
	}
}

func ParseState(jsonState json.RawMessage) (*state, error) {
	var s *state
	if err := json.Unmarshal([]byte(jsonState), &s); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *state) Update(g *simple.UndirectedGraph, m []protocol.Move) {
	s.Map.Rivers = graph.SerializeRivers(g)
	s.Turn += uint64(len(m))
}

type Brownian struct{}

func (Brownian) Name() string {
	return "Brownian"
}

func (Brownian) Setup(setup *protocol.Setup) (*protocol.Ready, error) {
	glog.Infof("Setup")

	s := InitializeState(setup)
	g := graph.BuildWithWeight(&s.Map, func(p uint64) float64 {
		if p == s.Punter {
			return 0.0
		}
		return math.Inf(0)
	})

	// TODO(akesling): Prioritize claiming rivers for mines with fewer owned rivers.
	// Add all mine-neighboring rivers to AvailableMineRivers
	for i := range s.Map.Mines {
		mine := s.Map.Mines[i]
		neighbors := g.From(g.Node(int64(s.Map.Mines[i])))
		for j := range neighbors {
			n := neighbors[j]
			newRiver := protocol.River{
				Source: mine,
				Target: protocol.SiteID(n.ID()),
			}
			s.AvailableMineRivers = append(s.AvailableMineRivers, newRiver)
			s.UnconnectedOrigins = append(s.UnconnectedOrigins, newRiver)
		}
	}
	ShuffleRivers(s.AvailableMineRivers)

	glog.Infof("Setup complete with available mine rivers %+v", s.AvailableMineRivers)
	return &protocol.Ready{
		Ready: s.Punter,
		State: s,
	}, nil
}

func (Brownian) Play(m []protocol.Move, jsonState json.RawMessage) (*protocol.GameplayOutput, error) {
	glog.Infof("Play")

	s, err := ParseState(jsonState)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling state %s: %v", string(jsonState), err)
	}

	weight := func(p uint64) float64 {
		if p == s.Punter {
			return 0.0
		}
		return math.Inf(0)
	}
	g := graph.BuildWithWeight(&s.Map, weight)
	graph.UpdateGraph(g, m, weight)
	s.Update(g, m)
	glog.Infof("Turn: %d", s.Turn)

	strategies := DetermineStrategies(s, g)
	for i := range strategies {
		output, err := strategies[i](s, g)
		if err != nil {
			return nil, err
		}
		if output != nil {
			return output, nil
		}
	}

	// No other moves were made, pass.
	return &protocol.GameplayOutput{
		Move: protocol.Move{
			Pass: &protocol.Pass{
				s.Punter,
			},
		},
		State: s,
	}, nil
}

func (Brownian) Stop(stop *protocol.Stop, jsonState json.RawMessage) error {
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

	var s Brownian
	if err := protocol.Play(os.Stdin, os.Stdout, &s); err != nil {
		glog.Exitf("Play failed: %v", err)
	}
}
