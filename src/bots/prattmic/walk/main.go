package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/golang/glog"
	"github.com/jemoster/icfp2017/src/graph"
	"github.com/jemoster/icfp2017/src/protocol"
	gograph "gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/path"
)

// maxMoves is the approximate maximum number of moves to plan ahead.
//
// A value of 1 means that we will plan only a single route, because a route
// can have at minimum 1 move. However, a route with more moves may be computed.
const maxMoves = 1

// TODO(prattmic): This is ugly. We should do work in a goroutine and just wait
// for an alarm on the main goroutine.
var (
	StartTime time.Time
	TimeoutThreshold time.Duration
)

const setupTimeoutThreshold = 9900*time.Millisecond
const playTimeoutThreshold = 900*time.Millisecond

// route is a mine-site route, which includes multiple intermediate sites.
type route struct {
	Mine protocol.SiteID
	Site protocol.SiteID
}

type futureMove struct {
	Move protocol.Move

	// Route is the route that this move is part of.
	Route route
}

type state struct {
	Punter  uint64
	Punters uint64
	Map     protocol.Map

	// Shortest distances for all mines.
	Distances graph.Distances

	Turn uint64

	// Moves are the next moves to take. The next move is the first in the
	// list.
	Moves []futureMove

	// Exhausted indicates we are permanently out of moves.
	Exhausted bool

	// CompletedRoutes are all the routes from mine to sites that have been
	// successfully completed.
	CompletedRoutes map[protocol.SiteID]map[protocol.SiteID]struct{}
}

func (s *state) weightFunc() graph.WeightFunc {
	return func(e *graph.MetadataEdge) float64 {
		if !e.IsOwned {
			return 1.0
		}

		if e.OwnerPunter == s.Punter {
			return 0.0
		}

		return math.Inf(0)
	}
}

type LongWalk struct{}

func (LongWalk) Name() string {
	return "longwalk"
}

// furthestNode returns the mine and site that are furthest apart, and how far
// they are.
//
// shortest is the set of shortest path structures for each mine.
// except is the set of routes not to consider.
func furthestNode(
	g *graph.Graph,
	s *state,
	shortest map[protocol.SiteID]*path.Shortest,
	except map[route]struct{}) (protocol.SiteID, protocol.SiteID, uint64, []gograph.Node) {
	var (
		mine     protocol.SiteID
		target   protocol.SiteID
		furthest uint64
		path     []gograph.Node
	)

	for _, m := range s.Map.Mines {
		// TODO(prattmic): clean this up, do initialization elsewhere.
		var initDistances bool
		if _, ok := s.Distances[m]; !ok {
			s.Distances[m] = make(map[protocol.SiteID]uint64, len(s.Map.Sites))
			initDistances = true
		}

		for _, site := range s.Map.Sites {
			r := route{m, site.ID}
			if _, ok := except[r]; ok {
				continue
			}

			n := g.Node(int64(site.ID))

			var dist float64
			if !initDistances {
				// Always use s.Distance for initial distance
				// if available.
				dist = float64(s.Distances[m][site.ID])
			} else {
				dist = shortest[m].WeightTo(n)
			}

			if initDistances {
				s.Distances[m][site.ID] = uint64(dist)
			}

			if math.IsInf(dist, 0) {
				// Unreachable.
				continue
			}

			if uint64(dist) > furthest {
				npath, _ := shortest[m].To(n)
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

func pickMoves(g *graph.Graph, s *state) []futureMove {
	// There are as many moves as rivers, but divided among all punters.
	n := ((len(s.Map.Rivers) + int(s.Punters)) / int(s.Punters)) - int(s.Turn)
	if n > maxMoves {
		n = maxMoves
	}

	// All of the routes taken by previous moves, plus new routes that will
	// be taken by future moves.
	taken := make(map[route]struct{})
	claims := make(map[claim]struct{}, n)
	moves := make([]futureMove, 0, n)

	for mine, m := range s.CompletedRoutes {
		for site, _ := range m {
			taken[route{mine, site}] = struct{}{}
		}
	}

	shortest := make(map[protocol.SiteID]*path.Shortest)
	for _, m := range s.Map.Mines {
		short := g.ShortestFrom(m)
		shortest[m] = &short
	}

	for len(moves) < n {
		if time.Since(StartTime) > TimeoutThreshold {
			glog.Infof("Out of time, got to go!")
			break
		}

		mine, target, dist, path := furthestNode(g, s, shortest, taken)
		glog.Infof("Furthest site: %d -> %d: %d, path: %+v", mine, target, dist, path)
		if dist == 0 || len(path) == 0 {
			break
		}
		r := route{mine, target}
		taken[r] = struct{}{}

		for i := 0; i < len(path)-1; i++ {
			if time.Since(StartTime) > TimeoutThreshold {
				glog.Infof("Out of time, got to go!")
				break
			}

			source := protocol.SiteID(path[i].ID())
			target := protocol.SiteID(path[i+1].ID())

			edge := g.EdgeBetween(path[i], path[i+1]).(*graph.MetadataEdge)
			if edge.IsOwned {
				if edge.OwnerPunter == s.Punter {
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
				glog.Infof("edge {%v, %v} already taken by %d", source, target, edge.OwnerPunter)
				break
			}

			c := makeClaim(source, target)
			if _, ok := claims[c]; ok {
				glog.Infof("edge %v already taken by previous move", c)
				continue
			}
			claims[c] = struct{}{}

			moves = append(moves, futureMove{
				Move: protocol.Move{
					Claim: &protocol.Claim{
						Punter: s.Punter,
						Source: source,
						Target: target,
					},
				},
				Route: r,
			})
			glog.Infof("move %d: %v", i, moves[len(moves)-1])
		}
	}

	return moves
}

func (LongWalk) Setup(setup *protocol.Setup) (*protocol.Ready, error) {
	StartTime = time.Now()
	TimeoutThreshold = setupTimeoutThreshold

	glog.Infof("Setup: game settings: %+v", setup.Settings)

	s := &state{
		Punter:  setup.Punter,
		Punters: setup.Punters,
		Map:     setup.Map,

		Distances: make(graph.Distances),

		Turn: 0,

		CompletedRoutes: make(map[protocol.SiteID]map[protocol.SiteID]struct{}),
	}

	g := graph.New(&s.Map, s.weightFunc())

	s.Moves = pickMoves(g, s)

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
func nextMove(g *graph.Graph, s *state) protocol.Move {
	for {
		if s.Exhausted {
			return protocol.Move{
				Pass: &protocol.Pass{
					s.Punter,
				},
			}
		}

		if len(s.Moves) <= 0 {
			glog.Warningf("Ran out of moves!")
			s.Moves = pickMoves(g, s)
			if len(s.Moves) <= 0 {
				if time.Since(StartTime) > TimeoutThreshold {
					// Moves not actually exhausted, we just ran out of time.
					return protocol.Move{
						Pass: &protocol.Pass{
							s.Punter,
						},
					}
				}

				glog.Warningf("Moves completely exhausted")
				s.Exhausted = true
				continue
			}
		}

		move := s.Moves[0]
		s.Moves = s.Moves[1:]

		if move.Move.Claim != nil {
			edge := g.EdgeBetween(g.Node(int64(move.Move.Claim.Source)), g.Node(int64(move.Move.Claim.Target))).(*graph.MetadataEdge)
			if edge.IsOwned {
				glog.Warningf("Move %v: river already taken by %d! Recomputing moves.", move, edge.OwnerPunter)
				s.Moves = pickMoves(g, s)
				continue
			}

			// TODO(prattmic): If we've really screwed up and this move is
			// invalid, CompletedRoutes will be out of sync.
			//
			// We've always completed at least a partial route. From the
			// mine to this move's target site.
			r := move.Route
			if s.CompletedRoutes[r.Mine] == nil {
				s.CompletedRoutes[r.Mine] = make(map[protocol.SiteID]struct{})
			}
			s.CompletedRoutes[r.Mine][move.Move.Claim.Target] = struct{}{}
		}

		return move.Move
	}
}

func checkMoves(g *graph.Graph, s *state, m []protocol.Move) {
	for _, theirMove := range m {
		if theirMove.Claim == nil {
			continue
		}

		their := makeClaim(theirMove.Claim.Source, theirMove.Claim.Target)

		for _, ourMove := range s.Moves {
			if ourMove.Move.Claim == nil {
				continue
			}

			our := makeClaim(ourMove.Move.Claim.Source, ourMove.Move.Claim.Target)
			if our == their {
				glog.Warningf("Future move %v taken by %d! Recomputing moves.", ourMove.Move, theirMove.Claim.Punter)
				s.Moves = pickMoves(g, s)
			}
		}
	}
}

func (LongWalk) Play(m []protocol.Move, jsonState json.RawMessage) (*protocol.GameplayOutput, error) {
	StartTime = time.Now()
	TimeoutThreshold = playTimeoutThreshold

	glog.Infof("Play")

	var s state
	if err := json.Unmarshal([]byte(jsonState), &s); err != nil {
		return nil, fmt.Errorf("error unmarshaling state %s: %v", string(jsonState), err)
	}

	glog.Infof("Turn: %d", s.Turn)

	// Add the most recent moves to the graph, and update state's copy of
	// the owned rivers.
	g := graph.New(&s.Map, s.weightFunc())
	g.Update(m)
	s.Map.Rivers = g.SerializeRivers()

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

	start := time.Now()

	var s LongWalk
	if err := protocol.Play(os.Stdin, os.Stdout, &s); err != nil {
		glog.Exitf("Play failed: %v", err)
	}

	glog.Infof("Run time: %v", time.Since(start))
}
