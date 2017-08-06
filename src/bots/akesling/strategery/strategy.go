package main

import (
	"math/rand"

	"github.com/golang/glog"
	"github.com/jemoster/icfp2017/src/graph"
	"github.com/jemoster/icfp2017/src/protocol"
	"gonum.org/v1/gonum/graph/simple"
)

type Strategy interface {
	Name() string
	SetUp(s *state, g *simple.UndirectedGraph) error

	IsApplicable(s *state, g *simple.UndirectedGraph) bool
	Run(s *state, g *simple.UndirectedGraph) (*protocol.GameplayOutput, error)
}

func AllStrategies(s *state, g *simple.UndirectedGraph) []Strategy {
	return []Strategy{CaptureMineAdjacentRivers{}, ConnectRivers{}, RandomWalkPaths{}}
}

func DetermineStrategies(s *state, g *simple.UndirectedGraph) []Strategy {
	return []Strategy{CaptureMineAdjacentRivers{}, ConnectRivers{}, RandomWalkPaths{}}
}

type StrategyStateRegistry struct {
	ActivePaths         [][]protocol.Site
	UnconnectedOrigins  []protocol.River
	AvailableMineRivers []protocol.River
}

type CaptureMineAdjacentRivers struct {}

func (CaptureMineAdjacentRivers) Name() string {
	return "CaptureMineAdjacentRivers"
}

func (CaptureMineAdjacentRivers) SetUp(s *state, g *simple.UndirectedGraph) error {
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
		}
	}
	ShuffleRivers(s.AvailableMineRivers)

	return nil
}

func (CaptureMineAdjacentRivers) IsApplicable(s *state, g *simple.UndirectedGraph) bool {
	if len(s.AvailableMineRivers) == 0 {
		return false
	}
	return true
}

func (CaptureMineAdjacentRivers) Run(s *state, g *simple.UndirectedGraph) (*protocol.GameplayOutput, error) {
	glog.Info("Surrounding mines")
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

	return nil, nil
}

type ConnectRivers struct {}

func (ConnectRivers) Name() string {
	return "ConnectRivers"
}

func (ConnectRivers) SetUp(s *state, g *simple.UndirectedGraph) error {
	for i := range s.Map.Mines {
		mine := s.Map.Mines[i]
		neighbors := g.From(g.Node(int64(s.Map.Mines[i])))
		for j := range neighbors {
			n := neighbors[j]
			s.UnconnectedOrigins = append(s.UnconnectedOrigins, protocol.River{
				Source: mine,
				Target: protocol.SiteID(n.ID()),
			})
		}
	}
	ShuffleRivers(s.UnconnectedOrigins)

	return nil
}

func (ConnectRivers) IsApplicable(s *state, g *simple.UndirectedGraph) bool {
	if len(s.UnconnectedOrigins) <= 1 {
		return false
	}
	return true
}

func (ConnectRivers) Run(s *state, g *simple.UndirectedGraph) (*protocol.GameplayOutput, error) {
	// Connect our rivers if possible.
	// TODO(akesling): Make sure we connect all paths through edges that aren't
	// _through_ a mine (i.e. foo -> mine1 -> bar -> mine2 won't count foo as
	// attached to the mine2).
	glog.Info("Connecting mines")
	var processed int
	for i := range s.UnconnectedOrigins {
		processed = i
		start := s.UnconnectedOrigins[i].Target
		for j := range s.UnconnectedOrigins {
			if s.UnconnectedOrigins[i].Source == s.UnconnectedOrigins[j].Source {
				// Same mine
				continue
			}

			end := s.UnconnectedOrigins[j].Target
			shortTree := graph.ShortestFrom(g, start)
			weight := shortTree.WeightTo(g.Node(int64(end)))
			if weight == 0 {
				// We're already connected
				continue
			}
			if weight >= float64(int64(len(s.Map.Rivers))-int64(s.Turn)) {
				// Not possible to reach
				continue
			}

			path, weight := shortTree.To(g.Node(int64(end)))
			for k := range path {
				if k == 0 {
					continue
				}

				edge := g.EdgeBetween(path[k-1], path[k]).(*graph.MetadataEdge)

				if edge.IsOwned && edge.OwnerPunter != s.Punter {
					glog.Errorf("Path between %d and %d runs through an opponent's river (%+v)", start, end, edge)
					break
				}

				if edge.IsOwned {
					// We must own it.
					continue
				}

				return &protocol.GameplayOutput{
					Move: protocol.Move{
						Claim: &protocol.Claim{
							Punter: s.Punter,
							Source: protocol.SiteID(edge.F.ID()),
							Target: protocol.SiteID(edge.T.ID()),
						},
					},
					State: s,
				}, nil
			}
		}
	}
	s.UnconnectedOrigins = s.UnconnectedOrigins[processed+1:]

	return nil, nil
}


type RandomWalkPaths struct {}

func (RandomWalkPaths) Name() string {
	return "RandomWalkPaths"
}

func (RandomWalkPaths) SetUp(s *state, g *simple.UndirectedGraph) error {
	return nil
}

func (RandomWalkPaths) IsApplicable(s *state, g *simple.UndirectedGraph) bool {
	if len(s.ActivePaths) == 0 {
		glog.Infof("No active paths available to follow.")
		return false
	}
	return true
}

func (RandomWalkPaths) Run(s *state, g *simple.UndirectedGraph) (*protocol.GameplayOutput, error) {
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
			if !edge.IsOwned {
				target = &protocol.Site{ID: protocol.SiteID(candidate.ID())}
				break ExtendPath
			}
		}
		glog.Infof("No available rivers for site %d", source.ID)
	}
	if source != nil && target != nil {
		s.ActivePaths[pathIndex] = append(toExtend[:end+1], *target)
	} else {
		s.ActivePaths = append(s.ActivePaths[:pathIndex], s.ActivePaths[pathIndex+1:]...)
		glog.Infof("No paths available from active path at index %d of %d, removing path.", pathIndex, len(s.ActivePaths))
	}

	if source != nil && target != nil {
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

	return nil, nil
}
