package main

import (
	"fmt"
	"io"
	"log"
	"net"

	"bufio"

	"github.com/jemoster/icfp2017/src/graph"

	. "github.com/jemoster/icfp2017/src/protocol"
	. "github.com/jemoster/icfp2017/src/protocol/io"
)

type Future struct {
	Source    SiteID `json:"source"`
	Target    SiteID `json:"target"`
	fulfilled bool
}

type Punter struct {
	ID uint64
	Name string
	
	conn net.Conn

	reader *bufio.Reader
	writer io.Writer

	futures  []Future
	splurges int
	options  int
}

type recvHandshake struct {
	Name string `json:"me"`
}

type sendHandshake struct {
	Name string `json:"you"`
}

type sendSetup struct {
	Punter   uint64   `json:"punter"`
	Punters  uint64   `json:"punters"`
	Map      *Map     `json:"map"`
	Settings Settings `json:"settings"`
}

type recvSetup struct {
	Ready   uint64   `json:"ready"`
	Futures []Future `json:"futures"`
}

type sendMove struct {
	Move struct {
		Moves []*Move `json:"moves"`
	} `json:"move"`
}

type recvMove struct {
	Move
}

type stop struct {
	Moves  []*Move `json:"moves"`
	Scores []Score `json:"scores"`
}

type sendStop struct {
	Stop stop `json:"stop"`
}

type Session struct {
	Map        Map
	Settings   Settings
	
	Punters    []Punter
	NumPunters int
	
	Graph *graph.Graph
}

func (s *Session) acceptMove(punter *Punter) (*Move, error) {
	var rM recvMove
	if err := Recv(punter.reader, &rM); err != nil {
		return nil, err
	}

Outer:
	switch {
	case rM.Move.Claim != nil:
		claim := rM.Move.Claim

		ok := s.Graph.HasEdgeBetween(claim.Source, claim.Target)

		if !ok {
			fmt.Printf("[%d] claimed a river that doesn't exist! D:", punter.ID)
			break
		}

		river := s.Graph.EdgeBetween(claim.Source, claim.Target).(*graph.MetadataEdge)

		if river.IsOwned {
			fmt.Printf("[%d] claimed a river that has already been claimed! D:\n", punter.ID)
			break
		}

		river.IsOwned = true
		river.OwnerPunter = punter.ID

		return &Move{Claim: claim}, nil

	case rM.Move.Splurge != nil:
		if !s.Settings.Splurges {
			fmt.Printf("[%d] tried to splurge, but splurging is disabled.\n", punter.ID)
			break
		}

		splurge := rM.Move.Splurge

		if len(splurge.Route) < 2 {
			fmt.Printf("[%d] tried to splurge, but did not specify enough sites.\n", punter.ID)
			break
		}

		if len(splurge.Route) > punter.splurges+1 {
			fmt.Printf("[%d] tried to splurge, but does not have enough lethargy (needs: %d, has: %d).\n", punter.ID, len(splurge.Route)-1, punter.splurges)
			break
		}

		optionsNeeded := 0

		src := splurge.Route[0]
		for _, tgt := range splurge.Route[1:] {
			edge := s.Graph.EdgeBetween(src, tgt).(*graph.MetadataEdge) // should make sure it exists first..

			if edge.IsOwned {
				if edge.IsOptioned {
					fmt.Printf("[%d] tried to splurge, but an edge on its path is owned and already optioned.\n", punter.ID)
					break Outer
				}

				if !s.Settings.Options {
					fmt.Printf("[%d] tried to splurge, but an edge on its path is owned and optioneds is not enabled.\n", punter.ID)
					break Outer
				}

				optionsNeeded += 1
			}

			src = tgt
		}

		if optionsNeeded > punter.options {
			fmt.Printf("[%d] tried to splurge, but does not have enough options (needs: %d, have: %d).\n", punter.ID, optionsNeeded, punter.options)
			break
		}

		src = splurge.Route[0]
		for _, tgt := range splurge.Route[1:] {
			edge := s.Graph.EdgeBetween(src, tgt).(*graph.MetadataEdge) // should make sure it exists first..

			if edge.IsOwned {
				edge.OptionPunter = punter.ID
				punter.options--
			} else {
				edge.IsOwned = true
				edge.OwnerPunter = punter.ID
			}

			src = tgt
		}

		punter.splurges -= len(splurge.Route)

		return &rM.Move, nil
	case rM.Move.Option != nil:
		if !s.Settings.Options {
			fmt.Printf("[%d] tried to option, but options is disabled.\n", punter.ID)
			break
		}

		option := rM.Move.Option

		if punter.options < 1 {
			fmt.Printf("[%d] tried to option, but has no options remaining.\n", punter.ID)
			break
		}

		edge := s.Graph.EdgeBetween(option.Source, option.Target).(*graph.MetadataEdge)

		if !edge.IsOwned {
			fmt.Printf("[%d] tried to option, but it isn't owned.  I don't know what to do here, so I'm going to just go with it.\n", punter.ID)
		}

		if edge.IsOptioned {
			fmt.Printf("[%d] tried to option, but the option has already been assigned.\n", punter.ID)
		}

		punter.options--
		edge.IsOptioned = true
		edge.OptionPunter = punter.ID

		return &rM.Move, nil
	}

	// If it is nothing else, then assume Pass.
	punter.splurges++

	return &Move{Pass: &Pass{punter.ID}}, nil
}

func (s *Session) play(srv net.Listener) ([]Score, error) {
	s.Graph = graph.New(&s.Map, func(e *graph.MetadataEdge) float64 { return 1.0 })

	s.Punters = make([]Punter, s.NumPunters)

	fmt.Printf("-\n")
	fmt.Printf("Waiting on clients...\n")

	for i := 0; i < s.NumPunters; i++ {
		conn, err := srv.Accept()
		if err != nil {
			return nil, err
		}
		defer conn.Close()

		s.Punters[i].ID = uint64(i)
		s.Punters[i].conn = conn
		s.Punters[i].reader = bufio.NewReader(conn)
		s.Punters[i].writer = conn
		s.Punters[i].splurges = 0
		s.Punters[i].options = len(s.Map.Mines)

		fmt.Printf("  [%d/%d] Client connected.\n", i+1, s.NumPunters)
	}

	for i := 0; i < s.NumPunters; i++ {
		punter := &s.Punters[i]

		var rH recvHandshake
		if err := Recv(punter.reader, &rH); err != nil {
			return nil, err
		}
		
		punter.Name = rH.Name

		fmt.Printf("Welcome, %s!\n", rH.Name)

		if err := Send(punter.writer, sendHandshake{rH.Name}); err != nil {
			return nil, err
		}
	}

	for i := 0; i < s.NumPunters; i++ {
		punter := &s.Punters[i]

		setup := sendSetup{
			uint64(i), uint64(s.NumPunters), &s.Map, Settings{Futures: false, Splurges: false, Options: false},
		}

		if err := Send(punter.writer, setup); err != nil {
			return nil, err
		}

		var rS recvSetup
		if err := Recv(punter.reader, &rS); err != nil {
			return nil, err
		}

		if rS.Ready != uint64(i) {
			log.Printf("[WARNING] Punter %d is very confused about it's identity.")
		}
	}

	sM := new(sendMove)
	sM.Move.Moves = make([]*Move, s.NumPunters)

	for i, _ := range s.Punters {
		sM.Move.Moves[i] = &Move{Pass: &Pass{uint64(i)}}
	}

	for curTurn := 0; curTurn < len(s.Map.Rivers); curTurn++ {
		punter := &s.Punters[curTurn%s.NumPunters]

		if err := Send(punter.writer, sM); err != nil {
			return nil, err
		}

		move, err := s.acceptMove(punter)
		if err != nil {
			return nil, err
		}

		sM.Move.Moves[punter.ID] = move

	}

	sv := s.Graph.Score(s.Map.Mines, s.NumPunters)

	sS := sendStop{
		stop{
			Moves:  sM.Move.Moves,
			Scores: sv,
		},
	}

	for i := 0; i < s.NumPunters; i++ {
		punter := &s.Punters[i]

		if err := Send(punter.writer, sS); err != nil {
			return nil, err
		}
	}

	return sv, nil
}

type Server struct {
	Map        Map
	Port       int
	NumPunters int
	Settings   Settings
	RunOnce    bool
}

func (s *Server) run(results *bufio.Writer) {
	laddr := fmt.Sprintf(":%d", s.Port)

	srv, err := net.Listen("tcp", laddr)
	if err != nil {
		log.Fatal(err)
	}
	defer srv.Close()

	fmt.Printf("-\n")
	fmt.Printf("Listening at %s\n", laddr)
	
	for {
		session := Session{
			Map:        s.Map,
			NumPunters: s.NumPunters,
			Settings:   s.Settings,
		}

		scores, err := session.play(srv)
		if err != nil {
			fmt.Printf("[ERROR] %+v\n", err)
			continue
		}
		
		fmt.Printf("Score: %+v\n", scores)

		results.WriteString(fmt.Sprintf("%d\n", session.NumPunters))
		for _, score := range scores {
			id := score.Punter
			name := session.Punters[id].Name
			
			results.WriteString(fmt.Sprintf("%s\n", name))
			results.WriteString(fmt.Sprintf("%d\n", score.Score))
		}
		results.Flush()
		
		if s.RunOnce {
			return
		}
	}
}
