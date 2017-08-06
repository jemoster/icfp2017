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

	// Shortest distances for all mines.
	Distances graph.Distances

	Turn uint64

	// Moves are the next moves to take. The next move is the first in the list.
	Moves []protocol.Move
}

type LongWalk struct{}

func (LongWalk) Name() string {
	return "longwalk"
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
		// TODO(prattmic): furtherNode is called multiple times, but
		// this shoul be done only once.
		shortest := graph.ShortestFrom(g, m)

		// TODO(prattmic): clean this up, do initialization elsewhere.
		var initDistances bool
		if _, ok := s.Distances[m]; !ok {
			s.Distances[m] = make(map[protocol.SiteID]uint64, len(s.Map.Sites))
			initDistances = true
		}

		for _, site := range s.Map.Sites {
			if _, ok := except[site.ID]; ok {
				continue
			}

			n := g.Node(int64(site.ID))

			var dist float64
			if !initDistances {
				// Always use s.Distance for initial distance
				// if available.
				dist = float64(s.Distances[m][site.ID])
			} else {
				dist = shortest.WeightTo(n)
			}

			if initDistances {
				s.Distances[m][site.ID] = uint64(dist)
			}

			if math.IsInf(dist, 0) {
				// Unreachable.
				continue
			}

			if uint64(dist) > furthest {
				npath, _ := shortest.To(n)
				if len(npath) == 0 {
					// Not reachable anymore.
					continue
				}
				path = npath
				mine = m
				target = site.ID
				furthest = uint64(dist)
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
		// FIXME(prattmic): Needs to use s.Distance, not g for
		// non-initial calls.
		mine, target, dist, path := furthestNode(g, s, taken)
		glog.Infof("Furthest site: %d -> %d: %d, path: %+v", mine, target, dist, path)
		if dist == 0 || len(path) == 0 {
			break
		}
		taken[target] = struct{}{}

		for i := 0; i < len(path)-1; i++ {
			source := protocol.SiteID(path[i].ID())
			target := protocol.SiteID(path[i+1].ID())

			edge := g.EdgeBetween(path[i], path[i+1]).(*graph.MetadataEdge)
			if edge.IsOwned {
				if edge.Punter == s.Punter {
					// We own this one, great! We don't
					// need this move again. Check the next
					// one.
					glog.Infof("edge {%v, %v} already taken by us!", source, target)
					continue
				}
				// Someone else owns this. Skip the rest of the moves.
				//
				// TODO(prattmic): We should only get here if
				// there are no other paths available.
				glog.Infof("edge {%v, %v} already taken by %d", source, target, edge.Punter)
				break
			}

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

		Distances: make(graph.Distances),

		Turn: 0,
	}

	g := graph.BuildWithWeight(&s.Map, func(p uint64) float64 {
		if p == s.Punter {
			return 0.0
		}
		return math.Inf(0)
	})

	// There are as many moves as rivers, but divided among all punters.
	n := (len(s.Map.Rivers) + int(s.Punters)) / int(s.Punters)
	s.Moves = pickMoves(g, s, n)

	return &protocol.Ready{
		Ready: s.Punter,
		State: s,
	}, nil
}

// nextMove returns the next move to make.
//
// If a river is already taken, nextMove skips it and returns the next valid
// move.
//
// TODO(prattmic): unfortunately that move might not be useful anymore.
func nextMove(g *simple.UndirectedGraph, s *state) protocol.Move {
	for {
		if len(s.Moves) <= 0 {
			glog.Warningf("Ran out of moves!")
			return protocol.Move{
				Pass: &protocol.Pass{
					s.Punter,
				},
			}
		}

		move := s.Moves[0]
		s.Moves = s.Moves[1:]

		if move.Claim != nil {
			edge := g.EdgeBetween(g.Node(int64(move.Claim.Source)), g.Node(int64(move.Claim.Target))).(*graph.MetadataEdge)
			if edge.IsOwned {
				glog.Warningf("Move %v: river already taken by %d! Recomputing moves.", move, edge.Punter)
				s.Moves = pickMoves(g, s, len(s.Moves))
				continue
			}
		}

		return move
	}
}

func checkMoves(g *simple.UndirectedGraph, s *state, m []protocol.Move)  {
	for _, theirMove := range m {
		if theirMove.Claim == nil {
			continue
		}

		their := makeClaim(theirMove.Claim.Source, theirMove.Claim.Target)

		for _, ourMove := range s.Moves {
			if ourMove.Claim == nil {
				continue
			}

			our := makeClaim(ourMove.Claim.Source, ourMove.Claim.Target)
			if our == their {
				glog.Warningf("Future move %v taken by %d! Recomputing moves.", ourMove, theirMove.Claim.Punter)
				s.Moves = pickMoves(g, s, len(s.Moves))
			}
		}
	}
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
	g := graph.BuildWithWeight(&s.Map, func(p uint64) float64 {
		if p == s.Punter {
			return 0.0
		}
		return math.Inf(0)
	})
	graph.UpdateGraph(g, m)
	s.Map.Rivers = graph.SerializeRivers(g)

	checkMoves(g, &s, m)

	move := nextMove(g, &s)

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
