package main

import (
	"encoding/json"
	"flag"
	"fmt"
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
	Punter         uint64
	Punters        uint64
	Map            protocol.Map
	ActiveMine     protocol.Site
	AvailableMines []protocol.Site
	PrevSites      []protocol.Site
	Distances      graph.Distances

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
	return "Punter: 76"
}

func (Brownian) Setup(setup *protocol.Setup) (*protocol.Ready, error) {
	glog.Infof("Setup")

	s := InitializeState(setup)
	g := graph.Build(&s.Map)

	s.Distances = graph.ShortestDistances(g, s.Map.Mines)
	s.ActiveMine = protocol.Site{s.Map.Mines[0]}
	s.AvailableMines = make([]protocol.Site, len(s.Map.Mines[1:]))
	for i, id := range s.Map.Mines[1:] {
		s.AvailableMines[i] = protocol.Site{id}
	}
	s.PrevSites = make([]protocol.Site, 1)
	s.PrevSites[0] = s.ActiveMine

	glog.Infof("Setup complete")
	return &protocol.Ready{
		Ready: s.Punter,
		State: s,
	}, nil
}

func getUnownedAdjacent(g *simple.UndirectedGraph, siteID protocol.SiteID) []protocol.SiteID {
	currNode := g.Node(int64(siteID))
	reachable := g.From(currNode)
	unowned := make([]protocol.SiteID, 0, len(reachable))
	for _, node := range reachable {
		edge := g.EdgeBetween(currNode, node)
		meta := edge.(*graph.MetadataEdge)
		if meta.IsOwned {
			continue
		}
		unowned = append(unowned, protocol.SiteID(node.ID()))
	}
	return unowned

}

type searchState struct {
	ActiveMine     protocol.Site
	AvailableMines []protocol.Site
	PrevSites      []protocol.Site
}

func newSearchState(s *state) searchState {
	var ss searchState
	ss.ActiveMine = s.ActiveMine
	ss.AvailableMines = make([]protocol.Site, len(s.AvailableMines))
	copy(ss.AvailableMines, s.AvailableMines)
	ss.PrevSites = make([]protocol.Site, len(s.PrevSites))
	copy(ss.PrevSites, s.PrevSites)
	return ss
}

func copySearchState(s *searchState) searchState {
	var ss searchState
	ss.ActiveMine = s.ActiveMine
	ss.AvailableMines = make([]protocol.Site, len(s.AvailableMines))
	copy(ss.AvailableMines, s.AvailableMines)
	ss.PrevSites = make([]protocol.Site, len(s.PrevSites))
	copy(ss.PrevSites, s.PrevSites)
	return ss
}

func findNextBest(prefix string, depth int, g *simple.UndirectedGraph, s *state, ss searchState, siteID protocol.SiteID) (bool, uint64, protocol.SiteID) {
	reachable := getUnownedAdjacent(g, siteID)

	for len(reachable) == 0 && len(ss.PrevSites) > 1 {
		glog.Infof("%sbacktrack %v -> %v", prefix, siteID, ss.PrevSites[len(ss.PrevSites)-1].ID)
		ss.PrevSites = ss.PrevSites[:len(ss.PrevSites)-1]
		glog.Infof("%sprev sites: %v len: %v", prefix, ss.PrevSites, len(ss.PrevSites))
		siteID = ss.PrevSites[len(ss.PrevSites)-1].ID
		reachable = getUnownedAdjacent(g, siteID)
	}

	// Shit, went back through all prev sites and none had more reachable
	// nodes, time to try new mines.
	if len(reachable) == 0 && len(ss.PrevSites) == 0 {
		for len(reachable) == 0 && len(ss.AvailableMines) > 0 {
			ss.ActiveMine = ss.AvailableMines[len(ss.AvailableMines)-1]
			ss.AvailableMines = ss.AvailableMines[:len(ss.AvailableMines)-1]
			ss.PrevSites = make([]protocol.Site, 1)
			ss.PrevSites[0] = ss.ActiveMine
			reachable = getUnownedAdjacent(g, ss.ActiveMine.ID)
		}
	}

	// No mines, no moves, couldn't find anything
	if len(reachable) == 0 {
		return false, 0, protocol.SiteID(0)
	}

	if depth > 1 {
		// Find best move that makes path longer
		ss.PrevSites = append(ss.PrevSites, protocol.Site{siteID})
		found, score, bestSite := findNextBest(prefix+"    ", depth-1, g, s, copySearchState(&ss), reachable[0])
		max := score
		best := protocol.SiteID(reachable[0].ID())
		glog.Infof("%sreachable: %v", prefix, reachable)
		for _, site := range reachable[1:] {
			var newFound bool
			newFound, score, bestSite = findNextBest(prefix+"    ", depth-1, g, s, copySearchState(&ss), site)
			if !found {
				found = newFound
			}
			glog.Infof("%sbestSite: %v max: %v score: %v at: %v", prefix, bestSite, max, score, ss.PrevSites)
			if score > max {
				max = score
				best = protocol.SiteID(site.ID())
			}
		}
		return found, max, best
	}

	// Find best move that makes path longer
	max := s.Distances[ss.ActiveMine.ID][protocol.SiteID(reachable[0].ID())]
	best := protocol.SiteID(reachable[0].ID())
	glog.Infof("%sreachable: %v", prefix, reachable)
	for _, site := range reachable {
		score := s.Distances[ss.ActiveMine.ID][protocol.SiteID(site.ID())]
		glog.Infof("%sbest: %v max: %v score: %v at: %v", prefix, best, max, score, ss.PrevSites)
		if score > max {
			max = s.Distances[ss.ActiveMine.ID][protocol.SiteID(site.ID())]
			best = protocol.SiteID(site.ID())
		}
	}
	return true, max, best
}

func (Brownian) Play(m []protocol.Move, jsonState json.RawMessage) (*protocol.GameplayOutput, error) {
	s, err := ParseState(jsonState)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling state %s: %v", string(jsonState), err)
	}

	g := graph.Build(&s.Map)
	graph.UpdateGraph(g, m)
	s.Update(g, m)

	currentSiteID := s.PrevSites[len(s.PrevSites)-1].ID
	found, score, site := findNextBest("", 3, g, s, newSearchState(s), currentSiteID)
	reachable := getUnownedAdjacent(g, currentSiteID)
	glog.Infof("Turn: %d found: %v score: %v site: %v reachable:%v at: %v", s.Turn, found, score, site, reachable, s.PrevSites)
	if !found {
		glog.Infof("Turn: %d PASS PASS PASS")
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

	s.PrevSites = append(s.PrevSites, protocol.Site{site})
	return &protocol.GameplayOutput{
		Move: protocol.Move{
			Claim: &protocol.Claim{
				Punter: s.Punter,
				Source: currentSiteID,
				Target: site,
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
