package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"

	"github.com/golang/glog"
	"github.com/jemoster/icfp2017/src/graph"
	"github.com/jemoster/icfp2017/src/protocol"
	gonumGraph "gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
)

type state struct {
	Punter  uint64
	Punters uint64
	Map     protocol.Map

	RiversToClaim []int

	Turn uint64
}

func (s *state) Update(g *simple.UndirectedGraph, m []protocol.Move) {
	s.Map.Rivers = graph.SerializeRivers(g)
	s.Turn += uint64(len(m))
}

// string representation of river where smaller id comes first
func riverKey(r protocol.River) string {
	if r.Source < r.Target {
		return fmt.Sprintf("%d %d", r.Source, r.Target)
	} else {
		return fmt.Sprintf("%d %d", r.Target, r.Source)
	}
}

type Blob struct{}

func (Blob) Name() string {
	return "blob"
}

func (Blob) Setup(setup *protocol.Setup) (*protocol.Ready, error) {
	glog.Infof("Setup")

	s := state{
		Punter:  setup.Punter,
		Punters: setup.Punters,
		Map:     setup.Map,
		Turn:    0,
	}

	// Pick a random mine.
	blobCenter := protocol.SiteID(rand.Intn(len(setup.Map.Mines)))

	g := graph.Build(&setup.Map)

	riverIdx := map[string]int{}
	for i, r := range setup.Map.Rivers {
		riverIdx[riverKey(r)] = i
	}

	queue := []gonumGraph.Node{g.Node(int64(blobCenter))}
	sitesVisited := map[int64]bool{}
	riversVisited := map[int]bool{}
	for len(queue) > 0 {
		site := queue[0]
		queue = queue[1:]
		sitesVisited[site.ID()] = true
		for _, neighbor := range g.From(site) {
			if !sitesVisited[neighbor.ID()] {
				r := protocol.River{
					Source: protocol.SiteID(site.ID()),
					Target: protocol.SiteID(neighbor.ID()),
				}
				ridx := riverIdx[riverKey(r)]
				if !riversVisited[ridx] {
					s.RiversToClaim = append(s.RiversToClaim, ridx)
					riversVisited[ridx] = true
				}
				queue = append(queue, neighbor)
			}
		}
	}

	glog.Infof("%d rivers starting from blobCenter %d: %v", len(s.RiversToClaim), blobCenter, s.RiversToClaim)

	return &protocol.Ready{
		Ready: s.Punter,
		State: s,
	}, nil
}

func (Blob) Play(m []protocol.Move, jsonState json.RawMessage) (*protocol.GameplayOutput, error) {
	glog.Infof("Play")

	var s state
	if err := json.Unmarshal([]byte(jsonState), &s); err != nil {
		return nil, fmt.Errorf("error unmarshaling state %s: %v", string(jsonState), err)
	}

	g := graph.Build(&s.Map)
	graph.UpdateGraph(g, m)
	s.Update(g, m)

	// Pick an unclaimed river, or pass if all are claimed.
	move := protocol.Move{}
	for _, i := range s.RiversToClaim {
		if !s.Map.Rivers[i].IsOwned {
			move.Claim = &protocol.Claim{
				Punter: s.Punter,
				Source: s.Map.Rivers[i].Source,
				Target: s.Map.Rivers[i].Target,
			}
		}
	}
	if move.Claim == nil {
		for _, r := range s.Map.Rivers {
			if !r.IsOwned {
				move.Claim = &protocol.Claim{
					Punter: s.Punter,
					Source: r.Source,
					Target: r.Target,
				}
			}
		}
	}
	if move.Claim == nil {
		move.Pass = &protocol.Pass{
			Punter: s.Punter,
		}
	}

	glog.Infof("Turn: %d", s.Turn)

	return &protocol.GameplayOutput{
		Move:  move,
		State: s,
	}, nil
}

func (Blob) Stop(stop *protocol.Stop, jsonState json.RawMessage) error {
	glog.Infof("Stop: %+v", stop)

	var s state
	if err := json.Unmarshal([]byte(jsonState), &s); err != nil {
		return fmt.Errorf("error unmarshaling state %s: %v", string(jsonState), err)
	}

	return nil
}

func main() {
	flag.Set("logtostderr", "true")
	flag.Parse()

	var s Blob
	if err := protocol.Play(os.Stdin, os.Stdout, &s); err != nil {
		glog.Exitf("Play failed: %v", err)
	}
}
