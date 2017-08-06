package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"math/rand"

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
	OwnedPaths          [][]protocol.Site
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

	g := graph.Build(&s.Map)
	graph.UpdateGraph(g, m)
	s.Update(g, m)
	glog.Infof("Turn: %d", s.Turn)

	if len(s.AvailableMineRivers) > 0 {
		// Grab all rivers around mines.
		i := 0
		for i < len(s.AvailableMineRivers) {
			candidate := s.AvailableMineRivers[i]
			edge := g.EdgeBetween(g.Node(int64(candidate.Source)), g.Node(int64(candidate.Target))).(*graph.MetadataEdge)
			i++

			if !edge.IsOwned {
				s.AvailableMineRivers = s.AvailableMineRivers[i:]
				s.ActivePaths = append(s.ActivePaths, []protocol.Site{protocol.Site{ID: candidate.Source}, protocol.Site{ID: candidate.Target}})

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

		// No rivers around mines are still available.
		s.AvailableMineRivers = make([]protocol.River, 0)
	}

	if len(s.AvailableMineRivers) == 0 && len(s.ActivePaths) > 0 {
		glog.Infof("Following an active path")
		// TODO(akesling): Go path by path instead of just following one.

		// Randomly extend our rivers now that mines are covered.
		pathIndex := rand.Intn(len(s.ActivePaths))
		toExtend := s.ActivePaths[pathIndex]
		end := len(toExtend) - 1
		var source *protocol.Site
		var target *protocol.Site
ExtendPath:
		for i := range toExtend {
			// Walk back down path until there's a path available.
			glog.Infof("Evaluating paths %+v", toExtend)
			end = len(toExtend) - 1 - i
			source = &toExtend[end]
			glog.Infof("Evaluating paths from site %d", source.ID)
			neighbors := g.From(g.Node(int64(toExtend[end].ID)))

			for j := range neighbors {
				candidate := neighbors[j]
				edge := g.EdgeBetween(g.Node(int64(source.ID)), candidate).(*graph.MetadataEdge)
				if (!edge.IsOwned) {
					target = &protocol.Site{ID: protocol.SiteID(candidate.ID())}
					break ExtendPath
				}
			}
			glog.Infof("No available rivers for site %d", source.ID)
		}
		if (source != nil && target != nil) {
			s.ActivePaths[pathIndex] = append(toExtend[:end+1], *target)
		} else {
			s.ActivePaths = append(s.ActivePaths[:pathIndex], s.ActivePaths[pathIndex+1:]...)
			glog.Infof("No paths available from active path at index %d of %d, removing path.", pathIndex, len(s.ActivePaths))
		}

		if (source != nil && target != nil) {
			glog.Infof("Path selected, from %d to %d", source.ID, target.ID)
			return &protocol.GameplayOutput{
				Move: protocol.Move{
					Claim: &protocol.Claim{
						Punter: s.Punter,
						Source: source.ID,
						Target: target.ID,
					},
				},
				State: s,
			}, nil
		}
	} else {
		glog.Infof("No active paths available to follow.")
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
