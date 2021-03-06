package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"os"

	"github.com/golang/glog"
	"github.com/jemoster/icfp2017/src/graph"
	"github.com/jemoster/icfp2017/src/protocol"
)

func ShuffleRivers(r []protocol.River) {
	for i := len(r) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		r[i], r[j] = r[j], r[i]
	}
}

type state struct {
	StrategyStateRegistry

	Punter              uint64
	Punters             uint64
	Map                 protocol.Map

	Turn uint64
}

func InitializeState(setup *protocol.Setup) *state {
	return &state{
		Punter:  setup.Punter,
		Punters: setup.Punters,
		Map:     setup.Map,

		Turn: 0,
	}
}

func ParseState(jsonState json.RawMessage) (*state, error) {
	var s *state
	if err := json.Unmarshal([]byte(jsonState), &s); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *state) Update(g *graph.Graph, m []protocol.Move) {
	s.Map.Rivers = g.SerializeRivers()
	s.Turn += uint64(len(m))
}

type Strategery struct{}

func (s *state) weightFunc() graph.WeightFunc {
	return func(e *graph.MetadataEdge) float64 {
		if !e.IsOwned {
			return 1.0
		}

		if e.OwnerPunter == s.Punter {
			return 0.0
		}

		return math.Inf(0)
	}
}

func (Strategery) Name() string {
	return "Strategery"
}

func (Strategery) Setup(setup *protocol.Setup) (*protocol.Ready, error) {
	glog.Infof("Setup")

	s := InitializeState(setup)
	g := graph.New(&s.Map, s.weightFunc())

	strategies := AllStrategies(s, g)
	for i := range strategies {
		err := strategies[i].SetUp(s, g)
		if err != nil {
			glog.Errorf("Error occurred in the setup of %s: %s", strategies[i].Name(), err)
		}
	}

	glog.Infof("Setup complete with available mine rivers %+v", s.AvailableMineRivers)
	return &protocol.Ready{
		Ready: s.Punter,
		State: s,
	}, nil
}

func (Strategery) Play(m []protocol.Move, jsonState json.RawMessage) (*protocol.GameplayOutput, error) {
	glog.Infof("Play")

	s, err := ParseState(jsonState)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling state %s: %v", string(jsonState), err)
	}

	g := graph.New(&s.Map, s.weightFunc())
	g.Update(m)
	s.Update(g, m)
	glog.Infof("Turn: %d", s.Turn)

	strategies := DetermineStrategies(s, g)
	for i := range strategies {
		strat := strategies[i]
		if !strat.IsApplicable(s, g) {
			continue
		}

		output, err := strat.Run(s, g)
		if err != nil {
			return nil, err
		}
		if output != nil {
			return output, nil
		}
	}

	// No other moves were made, pass.
	return &protocol.GameplayOutput{
		Move: protocol.Move{
			Pass: &protocol.Pass{
				s.Punter,
			},
		},
		State: s,
	}, nil
}

func (Strategery) Stop(stop *protocol.Stop, jsonState json.RawMessage) error {
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

	var s Strategery
	if err := protocol.Play(os.Stdin, os.Stdout, &s); err != nil {
		glog.Exitf("Play failed: %v", err)
	}
}
