package graph

import (
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/traverse"

	"github.com/jemoster/icfp2017/src/protocol"
)

func (g *Graph) Score(mines []protocol.SiteID, numPunters int, points func(punter uint64, src, dst protocol.SiteID) int64) []protocol.Score {
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
					s.Score += points(s.Punter, m, protocol.SiteID(dst.ID()))
				},
			}

			bft.Walk(g, m, nil)
		}
	}

	return scores
}
