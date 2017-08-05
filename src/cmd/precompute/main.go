package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
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

	distance uint64

	neighbors neighbors
}

type graph map[protocol.SiteID]*site
type neighbors map[protocol.SiteID]*site

type unvisited map[protocol.SiteID]*site

// Minimum returns the site with the minimum distance.
func (uv unvisited) Minimum() *site {
	// TODO(prattmic): Less horrific runtime.
	var min *site
	for _, s := range uv {
		if min == nil || s.distance < min.distance {
			min = s
		}
	}
	return min
}

// Copy returns a new graph with cleared visited and distance.
func (g graph) Copy() (graph, unvisited) {
	ng := make(graph)
	nuv := make(unvisited)

	for i, s := range g {
		ng[i] = &site{
			id:        s.id,
			distance:  math.MaxUint64,
			neighbors: make(neighbors),
		}
		nuv[i] = ng[i]
	}

	for i, s := range g {
		for ni := range s.neighbors {
			ng[i].neighbors[ni] = ng[ni]
		}
	}

	return ng, nuv
}

// String implements fmt.Stringer.
//
// It shows every node, but not nicely with their connections.
func (g graph) String() string {
	var buf bytes.Buffer
	for i, s := range g {
		fmt.Fprintf(&buf, "%d: %+v\n", i, s)
	}

	return buf.String()
}

func buildGraph(m *protocol.Map) (graph, unvisited, error) {
	g := make(graph)
	uv := make(unvisited)

	for _, si := range m.Sites {
		if s, ok := g[si.ID]; ok {
			return nil, nil, fmt.Errorf("found duplicate site %+v for %+v", s, si)
		}

		g[si.ID] = &site{
			id:        si.ID,
			distance:  math.MaxUint64,
			neighbors: make(map[protocol.SiteID]*site),
		}
		uv[si.ID] = g[si.ID]
	}

	for _, r := range m.Rivers {
		source, ok := g[r.Source]
		if !ok {
			return nil, nil, fmt.Errorf("source missing for river %+v", r)
		}

		target, ok := g[r.Target]
		if !ok {
			return nil, nil, fmt.Errorf("target missing for river %+v", r)
		}

		source.neighbors[target.id] = target
		target.neighbors[source.id] = source
	}

	for _, mine := range m.Mines {
		if _, ok := g[mine]; !ok {
			return nil, nil, fmt.Errorf("mine %+v missing in graph %+v", mine, g)
		}
	}

	return g, uv, nil
}

// fillGraph fills in distance from start to every site in the graph using Dijkstra's algorithm.
func fillGraph(g graph, uv unvisited, start protocol.SiteID) error {
	curr, ok := g[start]
	if !ok {
		return fmt.Errorf("start %v missing in %+v", start, g)
	}

	curr.distance = 0

	for _, n := range curr.neighbors {
		dist := curr.distance + 1
		if dist < n.distance {
			n.distance = dist
		}
	}

	delete(uv, curr.id)

	for {
		curr := uv.Minimum()
		if curr == nil {
			break
		}

		for _, n := range curr.neighbors {
			dist := curr.distance + 1
			if dist < n.distance {
				n.distance = dist
			}
		}

		delete(uv, curr.id)
	}

	return nil
}

func main() {
	m, err := readMap()
	if err != nil {
		log.Fatalf("Failed to read map: %v", err)
	}

	fmt.Printf("Parsed map: %+v\n", m)

	g, uv, err := buildGraph(m)
	if err != nil {
		log.Fatalf("Failed to build graph: %v", err)
	}

	fmt.Printf("Built graph: %+v unvisited: %+v\n", g, uv)

	// TODO(prattmic): fill a graph for each mine and serialize out.
	if err := fillGraph(g, uv, m.Mines[0]); err != nil {
		log.Fatalf("Failed to fill graph: %v", err)
	}

	fmt.Printf("Filled graph: %+v", g)
}
