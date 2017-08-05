package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/golang/glog"
	"github.com/jemoster/icfp2017/src/protocol"
	"github.com/jemoster/icfp2017/src/graph"
	"gonum.org/v1/gonum/graph/simple"
)

type state struct {
	Punter  uint64
	Punters uint64
	Map     protocol.Map
	OwnedPaths [][]protocol.Site
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

func UpdateGraph(g *simple.UndirectedGraph, m []protocol.Move) {
	for i := range m {
		move := m[i]
		if move.Claim != nil {
			claimedEdge := g.EdgeBetween(
				g.Node(int64(move.Claim.Source)),
				g.Node(int64(move.Claim.Target))).(*graph.MetadataEdge)
			claimedEdge.IsOwned = true
			claimedEdge.Punter = move.Claim.Punter
		}
	}
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
	g := graph.Build(&s.Map)

	// Add all mine-neighboring rivers to AvailableMineRivers
	for i := range s.Map.Mines {
		mine := s.Map.Mines[i]
		neighbors := g.From(g.Node(int64(s.Map.Mines[i])))
		for j := range neighbors {
			n := neighbors[j]
			s.AvailableMineRivers = append(
				s.AvailableMineRivers,
				protocol.River{
					Source: mine,
					Target: protocol.SiteID(n.ID()),
				})
		}
	}

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

	g := graph.Build(&s.Map)
	UpdateGraph(g, m)
	s.Update(g, m)
	glog.Infof("Turn: %d", s.Turn)

	// OwnedPaths [][]protocol.Site
	// AvailableMineRivers []protocol.River
	if len(s.AvailableMineRivers) != 0 {
		// Grab all rivers around mines.
		i := 0
		for i < len(s.AvailableMineRivers) {
			candidate := s.AvailableMineRivers[i]
			edge := g.EdgeBetween(g.Node(int64(candidate.Source)), g.Node(int64(candidate.Target))).(*graph.MetadataEdge)
			i++

			if !edge.IsOwned {
				s.AvailableMineRivers = s.AvailableMineRivers[i:]
				s.OwnedPaths = append(s.OwnedPaths, []protocol.Site{protocol.Site{ID: candidate.Source}, protocol.Site{ID: candidate.Source}})

				return &protocol.GameplayOutput{
					Move: protocol.Move{
						Claim: &protocol.Claim{
							Punter: s.Punter,
							Source: candidate.Source,
							Target: candidate.Target,
						},
					},
					State: s,
				}, nil
			}
		}

		s.AvailableMineRivers = make([]protocol.River, 0)
	}

	if len(s.AvailableMineRivers) == 0 {
		// Randomly extend our rivers.
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
