package graph

import (
	"math"

	"github.com/golang/glog"
	"github.com/jemoster/icfp2017/src/protocol"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/path"
	"gonum.org/v1/gonum/graph/simple"
)

type MetadataEdge struct {
	F graph.Node
	T graph.Node
	W float64

	IsOwned     bool
	OwnerPunter uint64

	// IsOptioned means that OptionPunter holds the option on this edge.
	IsOptioned   bool
	OptionPunter uint64
}

func (m MetadataEdge) From() graph.Node {
	return m.F
}

func (m MetadataEdge) To() graph.Node {
	return m.T
}

func (m MetadataEdge) Weight() float64 {
	return m.W
}

// WeightFunc returns the weight of an edge.
type WeightFunc func(e *MetadataEdge) float64

// Graph is a simple.UndirectedGraph representing a protocol.Map.
type Graph struct {
	*simple.UndirectedGraph

	weight WeightFunc
}

// BuildWithWeight returns a graph.Graph that represents m.
//
// ownedWeight returns the weight of an edge owned by p.
func New(m *protocol.Map, weight WeightFunc) *Graph {
	g := &Graph{
		UndirectedGraph: simple.NewUndirectedGraph(0.0, math.Inf(0)),
		weight:          weight,
	}

	for _, si := range m.Sites {
		g.AddNode(simple.Node(si.ID))
	}

	var river *MetadataEdge
	for _, r := range m.Rivers {
		river = &MetadataEdge{
			F:          g.Node(int64(r.Source)),
			T:          g.Node(int64(r.Target)),
			IsOwned:    r.IsOwned,
			IsOptioned: r.IsOptioned,
		}
		if r.IsOwned {
			river.OwnerPunter = r.OwnerPunter
		}
		if r.IsOptioned {
			river.OptionPunter = r.OptionPunter
		}
		river.W = g.weight(river)
		g.SetEdge(river)
	}

	return g
}

// SerialRivers returns a slice of rivers containing edge metadata.
func (g *Graph) SerializeRivers() []protocol.River {
	edges := g.Edges()
	rivers := make([]protocol.River, len(edges))
	for i := range edges {
		curEdge := edges[i].(*MetadataEdge)
		rivers[i] = protocol.River{
			Source:     protocol.SiteID(curEdge.From().ID()),
			Target:     protocol.SiteID(curEdge.To().ID()),
			IsOwned:    curEdge.IsOwned,
			IsOptioned: curEdge.IsOptioned,
		}

		if curEdge.IsOwned {
			rivers[i].OwnerPunter = curEdge.OwnerPunter
		}
		if curEdge.IsOptioned {
			rivers[i].OptionPunter = curEdge.OptionPunter
		}
	}
	return rivers
}

// Update adds the effect of the passed moves to the graph.
func (g *Graph) Update(m []protocol.Move) {
	for i := range m {
		move := m[i]
		var claim bool // true for claim, false for option.
		var route []protocol.SiteID
		var punter uint64
		switch {
		case move.Claim != nil:
			claim = true
			route = []protocol.SiteID{move.Claim.Source, move.Claim.Target}
			punter = move.Claim.Punter
		case move.Splurge != nil:
			claim = true
			route = move.Splurge.Route
			punter = move.Splurge.Punter
		case move.Option != nil:
			route = []protocol.SiteID{move.Option.Source, move.Option.Target}
			punter = move.Option.Punter
		}

		if len(route) < 2 {
			continue
		}

		for i := 0; i < len(route)-1; i++ {
			source := route[i]
			target := route[i+1]

			e := g.EdgeBetween(g.Node(int64(source)), g.Node(int64(target)))
			if e == nil {
				glog.Warningf("Invalid river {%d, %d} in move %v", source, target, move)
				continue
			}
			edge := e.(*MetadataEdge)
			if claim {
				edge.IsOwned = true
				edge.OwnerPunter = punter
			} else {
				edge.IsOptioned = true
				edge.OptionPunter = punter
			}
			edge.W = g.weight(edge)
		}
	}
}

// ShortestFrom returns a path.Shortest for a specific mine.
func (g *Graph) ShortestFrom(mine protocol.SiteID) path.Shortest {
	return path.DijkstraFrom(g.Node(int64(mine)), g.UndirectedGraph)
}

// Distances is a map from source mine ID to map of target site ID to distance.
type Distances map[protocol.SiteID]map[protocol.SiteID]uint64

// ShortestDistances returns the distances from each mine to every site.
//
// g is a graph created by Build.
func (g *Graph) ShortestDistances(mines []protocol.SiteID) Distances {
	sites := g.Nodes()
	results := make(Distances, len(mines))
	for _, mine := range mines {
		shortest := g.ShortestFrom(mine)

		results[mine] = make(map[protocol.SiteID]uint64, len(sites))
		for _, site := range sites {
			// All edges are weight 1, so the weight to a node is
			// the distance to that node.
			results[mine][protocol.SiteID(site.ID())] = uint64(shortest.WeightTo(g.Node(site.ID())))
		}
	}

	return results
}
