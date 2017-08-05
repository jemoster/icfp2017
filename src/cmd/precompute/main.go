package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/jemoster/icfp2017/src/graph"
	"github.com/jemoster/icfp2017/src/protocol"
	"gonum.org/v1/gonum/graph/path"
)

var debug = flag.Bool("v", false, "verbose logging")

func debugf(format string, v ...interface{}) {
	if *debug {
		log.Printf(format, v...)
	}
}

func readMap() (*protocol.Map, error) {
	b, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("failed to read map: %v", err)
	}

	m := &protocol.Map{}
	if err := json.Unmarshal(b, m); err != nil {
		return nil, fmt.Errorf("failed to unmarshal map: %v", err)
	}

	return m, nil
}

// distances contains the distance (value) to a target (key) for one source.
type distances map[protocol.SiteID]uint64

func main() {
	flag.Parse()

	m, err := readMap()
	if err != nil {
		log.Fatalf("Failed to read map: %v", err)
	}

	debugf("Parsed map: %+v\n", m)

	g := graph.Build(m)

	debugf("Graph: %+v\n", g)

	results := make(map[protocol.SiteID]distances, len(m.Mines))
	for _, mine := range m.Mines {
		shortest := path.DijkstraFrom(g.Node(int64(mine)), g)

		results[mine] = make(distances, len(m.Sites))
		for _, si := range m.Sites {
			// Add edges are weight 1, so the weight to a node is
			// the distance to that node.
			results[mine][si.ID] = uint64(shortest.WeightTo(g.Node(int64(si.ID))))
		}
	}

	debugf("Results: %+v\n", results)

	b, err := json.Marshal(results)
	if err != nil {
		log.Fatalf("Failed to marshal results: %v", err)
	}

	fmt.Println(string(b))
}
