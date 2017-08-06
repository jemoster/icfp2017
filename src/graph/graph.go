package graph

import (
	"math"

	"github.com/jemoster/icfp2017/src/protocol"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/path"
	"gonum.org/v1/gonum/graph/simple"
)

type MetadataEdge struct {
	F graph.Node
	T graph.Node
	W float64

	IsOwned bool
	Punter  uint64
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

// Build returns a graph.Graph that represents m.
func Build(m *protocol.Map) *simple.UndirectedGraph {
	g := simple.NewUndirectedGraph(0.0, math.Inf(0))

	for _, si := range m.Sites {
		g.AddNode(simple.Node(si.ID))
	}

	var river *MetadataEdge
	for _, r := range m.Rivers {
		river = &MetadataEdge{
			F: g.Node(int64(r.Source)),
			T: g.Node(int64(r.Target)),
			W: 1.0,
		}
		if r.IsOwned {
			river.IsOwned = true
			river.Punter = r.Punter
		}
		g.SetEdge(river)
	}

	return g
}

func SerializeRivers(g *simple.UndirectedGraph) []protocol.River {
	edges := g.Edges()
	rivers := make([]protocol.River, len(edges))
	for i := range edges {
		curEdge := edges[i].(*MetadataEdge)
		rivers[i] = protocol.River{
			Source: protocol.SiteID(curEdge.From().ID()),
			Target: protocol.SiteID(curEdge.To().ID()),
		}

		if curEdge.IsOwned {
			rivers[i].IsOwned = true
			rivers[i].Punter = curEdge.Punter
		}
	}
	return rivers
}

func UpdateGraph(g *simple.UndirectedGraph, m []protocol.Move) {
	for i := range m {
		move := m[i]
		if move.Claim != nil {
			claimedEdge := g.EdgeBetween(
				g.Node(int64(move.Claim.Source)),
				g.Node(int64(move.Claim.Target))).(*MetadataEdge)
			claimedEdge.IsOwned = true
			claimedEdge.Punter = move.Claim.Punter
		}
	}
}

// ShortestFrom returns a path.Shortest for a specific mine.
func ShortestFrom(g *simple.UndirectedGraph, mine protocol.SiteID) path.Shortest {
	return path.DijkstraFrom(g.Node(int64(mine)), g)
}

// Distances is a map from source mine ID to map of target site ID to distance.
type Distances map[protocol.SiteID]map[protocol.SiteID]uint64

// ShortestDistances returns the distances from each mine to every site.
//
// g is a graph created by Build.
func ShortestDistances(g *simple.UndirectedGraph, mines []protocol.SiteID) Distances {
	sites := g.Nodes()
	results := make(Distances, len(mines))
	for _, mine := range mines {
		shortest := ShortestFrom(g, mine)

		results[mine] = make(map[protocol.SiteID]uint64, len(sites))
		for _, site := range sites {
			// All edges are weight 1, so the weight to a node is
			// the distance to that node.
			results[mine][protocol.SiteID(site.ID())] = uint64(shortest.WeightTo(g.Node(site.ID())))
		}
	}

	return results
}
