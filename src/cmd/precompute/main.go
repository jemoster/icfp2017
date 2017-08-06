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

	g := graph.New(m, func(*graph.MetadataEdge) float64 {
		return 1.0
	})

	debugf("Graph: %+v\n", g)

	results := g.ShortestDistances(m.Mines)

	debugf("Results: %+v\n", results)

	b, err := json.Marshal(results)
	if err != nil {
		log.Fatalf("Failed to marshal results: %v", err)
	}

	fmt.Println(string(b))
}
