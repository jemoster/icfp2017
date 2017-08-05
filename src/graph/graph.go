package graph

import (
	"math"

	"github.com/jemoster/icfp2017/src/protocol"
	"gonum.org/v1/gonum/graph/path"
	"gonum.org/v1/gonum/graph/simple"
)

// Build returns a graph.Graph that represents m.
func Build(m *protocol.Map) *simple.UndirectedGraph {
	g := simple.NewUndirectedGraph(0.0, math.Inf(0))

	for _, si := range m.Sites {
		g.AddNode(simple.Node(si.ID))
	}

	for _, r := range m.Rivers {
		g.SetEdge(simple.Edge{
			F: g.Node(int64(r.Source)),
			T: g.Node(int64(r.Target)),
			W: 1.0,
		})
	}

	return g
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
