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
func furthestNode(g *simple.UndirectedGraph, s *state, except map[protocol.SiteID]struct{}) (protocol.SiteID, protocol.SiteID, uint64, []gograph.Node) {
	var (
		mine     protocol.SiteID
		target   protocol.SiteID
		furthest uint64
		path     []gograph.Node
	)

	for _, m := range s.Map.Mines {
		shortest := graph.ShortestFrom(g, m)

		for _, site := range s.Map.Sites {
			if _, ok := except[site.ID]; ok {
				continue
			}

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

// claim is equivalent to protocol.Claim, but a must always be greater than b,
// so equivalent claims hash as equal.
type claim struct {
	a protocol.SiteID
	b protocol.SiteID
}

func makeClaim(source, target protocol.SiteID) claim {
	if source < target {
		return claim{source, target}
	}
	return claim{target, source}
}

func pickMoves(g *simple.UndirectedGraph, s *state, n int) []protocol.Move {
	taken := make(map[protocol.SiteID]struct{})
	claims := make(map[claim]struct{}, n)
	moves := make([]protocol.Move, 0, n)

	for len(moves) < n {
		mine, target, dist, path := furthestNode(g, s, taken)
		glog.Infof("Furthest site: %d -> %d: %d, path: %+v", mine, target, dist, path)
		if dist == 0 {
			break
		}
		taken[target] = struct{}{}

		for i := uint64(0); i < dist; i++ {
			source := protocol.SiteID(path[i].ID())
			target := protocol.SiteID(path[i+1].ID())

			c := makeClaim(source, target)
			if _, ok := claims[c]; ok {
				glog.Infof("edge %v already taken by previous move", c)
				continue
			}
			claims[c] = struct{}{}

			moves = append(moves, protocol.Move{
				Claim: &protocol.Claim{
					Punter: s.Punter,
					Source: source,
					Target: target,
				},
			})
			glog.Infof("move %d: %v", i, moves[len(moves)-1])
		}
	}

	return moves
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

	// There are as many moves as rivers, but divided among all punters.
	n := (len(s.Map.Rivers) + int(s.Punters)) / int(s.Punters)
	s.Moves = pickMoves(g, s, n)

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

	// Add the most recent moves to the graph, and update state's copy of
	// the owned rivers.
	g := graph.Build(&s.Map)
	graph.UpdateGraph(g, m)
	s.Map.Rivers = graph.SerializeRivers(g)

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

	if move.Claim != nil {
		edge := g.EdgeBetween(g.Node(int64(move.Claim.Source)), g.Node(int64(move.Claim.Target))).(*graph.MetadataEdge)
		if edge.IsOwned {
			glog.Warningf("River already taken by %d!", edge.Punter)
		}
	}

	glog.Infof("Playing: %v", move)

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
