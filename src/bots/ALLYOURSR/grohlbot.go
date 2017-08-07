package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/golang/glog"
	"github.com/jemoster/icfp2017/src/protocol"
	"github.com/jemoster/icfp2017/src/graph"
	"sort"
)

type state struct {
	Punter  uint64
	Punters uint64
	Map     protocol.Map

	Turn uint64
}

type Scorekeeper struct{

}

func sortByValue(mp map[int64]int64) PairList{
	pl := make(PairList, len(mp))
	i := 0
	for k, v := range mp {
		pl[i] = Pair{k, v}
		i++
	}
	sort.Sort(sort.Reverse(pl))
	return pl
}

type Pair struct {
	Key int64
	Value int64
}

type PairList []Pair

func (p PairList) Len() int { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p PairList) Swap(i, j int){ p[i], p[j] = p[j], p[i] }

type Grohlbot struct{}

func (Grohlbot) Name() string {
	return "THEBEST!THEBEST!THEBEST!"
}




func (Grohlbot) Setup(setup *protocol.Setup) (*protocol.Ready, error) {
	glog.Infof("Setup")

	s := state{
		Punter:  setup.Punter,
		Punters: setup.Punters,
		Map:     setup.Map,

		Turn: 0,
	}

	graph.ShortestDistances()

	return &protocol.Ready{
		Ready: s.Punter,
		State: s,
	}, nil
}

func (Grohlbot) Play(m []protocol.Move, jsonState json.RawMessage) (*protocol.GameplayOutput, error) {
	glog.Infof("Play")

	var s state
	if err := json.Unmarshal([]byte(jsonState), &s); err != nil {
		return nil, fmt.Errorf("error unmarshaling state %s: %v", string(jsonState), err)
	}




	s.Turn++
	glog.Infof("Turn: %d", s.Turn)

	shortestPaths := graph.ShortestPaths(graph.Build(&s.Map), s.Map.Mines)
	hitCounts := graph.SPHitCount(&shortestPaths)

	sorted := sortByValue(hitCounts)

	for _, kvp := range sorted{
		if(s.Map.Sites[kvp.Key].)
		//Search sorted list for high scoring unowned node.
	}

	//Move is a struct which consists of pointers to structs Claim, Pass, etc
	return &protocol.GameplayOutput{
		Move: protocol.Move{
			Pass: &protocol.Pass{
				s.Punter,
			},
		},
		State: s,
	}, nil
}

func (Grohlbot) Stop(stop *protocol.Stop, jsonState json.RawMessage) error {
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

	var s Grohlbot
	if err := protocol.Play(os.Stdin, os.Stdout, &s); err != nil {
		glog.Exitf("Play failed: %v", err)
	}
}
