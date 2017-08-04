package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/jemoster/icfp2017/src/protocol"
)

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

type site struct {
	id protocol.SiteID

	visited  bool
	distance uint64

	neighbors map[protocol.SiteID]*site
}

type graph map[protocol.SiteID]*site

func buildGraph(m *protocol.Map) (graph, error) {
	g := make(graph)

	for _, si := range m.Sites {
		if s, ok := g[si.ID]; ok {
			return nil, fmt.Errorf("found duplicate site %+v for %+v", s, si)
		}

		g[si.ID] = &site{
			id:        si.ID,
			neighbors: make(map[protocol.SiteID]*site),
		}
	}

	for _, r := range m.Rivers {
		source, ok := g[r.Source]
		if !ok {
			return nil, fmt.Errorf("source missing for river %+v", r)
		}

		target, ok := g[r.Target]
		if !ok {
			return nil, fmt.Errorf("target missing for river %+v", r)
		}

		source.neighbors[target.id] = target
		target.neighbors[source.id] = source
	}

	for _, mine := range m.Mines {
		if _, ok := g[mine]; !ok {
			return nil, fmt.Errorf("mine %+v missing in graph %+v", mine, g)
		}
	}

	return g, nil
}

func main() {
	m, err := readMap()
	if err != nil {
		log.Fatalf("Failed to read map: %v", err)
	}

	fmt.Printf("Parsed map: %+v\n", m)

	g, err := buildGraph(m)
	if err != nil {
		log.Fatalf("Failed to build graph: %v", err)
	}

	fmt.Printf("Built graph: %+v\n", g)
}
