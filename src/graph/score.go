package graph

import (
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/traverse"

	"github.com/jemoster/icfp2017/src/protocol"
)

func (g *Graph) Score(mines []protocol.SiteID, numPunters int) []protocol.Score {
	dist := g.ShortestDistances(mines)

	scores := make([]protocol.Score, numPunters)

	for i := 0; i < numPunters; i++ {
		s := &scores[i]

		s.Punter = uint64(i)
		s.Score = 0

		for _, m := range mines {
			bft := traverse.BreadthFirst{
				EdgeFilter: func(e graph.Edge) bool {
					d := e.(*MetadataEdge)
					if d.IsOwned && d.OwnerPunter == uint64(i) {
						return true
					} else if d.IsOptioned && d.OptionPunter == uint64(i) {
						return true
					}
					return false
				},
				Visit: func(src, dst graph.Node) {
					d := dist[m][protocol.SiteID(dst.ID())]
					s.Score += int64(d * d)
				},
			}

			bft.Walk(g, m, nil)
		}
	}

	return scores
}
