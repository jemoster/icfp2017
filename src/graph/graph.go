package graph

import (
	"math"

	"github.com/jemoster/icfp2017/src/protocol"
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
