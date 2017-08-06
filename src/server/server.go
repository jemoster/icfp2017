package main

import (
	"io"
	"io/ioutil"
	"os"

	"flag"
	"fmt"
	"log"
	"net"

	"bufio"
	"encoding/json"

	"github.com/jemoster/icfp2017/src/graph"

	. "github.com/jemoster/icfp2017/src/protocol"
	. "github.com/jemoster/icfp2017/src/protocol/io"
)

func loadMap(path string) (*Map, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	m := new(Map)

	err = json.Unmarshal(data, m)
	if err != nil {
		return nil, err
	}

	return m, nil
}

type Punter struct {
	conn net.Conn

	reader *bufio.Reader
	writer io.Writer
}

type recvHandshake struct {
	Name string `json:"me"`
}

type sendHandshake struct {
	Name string `json:"you"`
}

type sendSetup struct {
	Punter  uint64 `json:"punter"`
	Punters uint64 `json:"punters"`
	Map     *Map   `json:"map"`
}

type recvSetup struct {
	Ready uint64 `json:"ready"`
}

type sendMove struct {
	Move struct {
		Moves []Move `json:"moves"`
	} `json:"move"`
}

type recvMove struct {
	Move
}

type sendStop struct {
	Stop Stop `json:"stop"`
}

func main() {
	mapPath := flag.String("map", "", "[required] A JSON file")
	srvPort := flag.Int("port", 9001, "The port to listen on")
	numPunters := flag.Int("punters", 2, "Number of players")

	flag.Parse()

	if len(*mapPath) < 1 {
		log.Fatal("map can not be undefined")
	}

	mapData, err := loadMap(*mapPath)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Map Loaded (from %q)\n", *mapPath)
	fmt.Printf("  Sites:  %d\n", len(mapData.Sites))
	fmt.Printf("  Rivers: %d\n", len(mapData.Rivers))
	fmt.Printf("  Mines:  %d\n", len(mapData.Mines))

	gMap := graph.Build(mapData)

	laddr := fmt.Sprintf(":%d", *srvPort)
	srv, err := net.Listen("tcp", laddr)
	if err != nil {
		log.Fatal(err)
	}
	defer srv.Close()

	fmt.Printf("-\n")
	fmt.Printf("Listening at %s\n", laddr)
	fmt.Printf("-\n")
	fmt.Printf("Waiting on clients...\n")

	punters := make([]Punter, *numPunters)

	for i := 0; i < *numPunters; i++ {
		conn, err := srv.Accept()
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()

		punters[i].conn = conn
		punters[i].reader = bufio.NewReader(conn)
		punters[i].writer = conn

		fmt.Printf("  [%d/%d] Client connected.\n", i+1, *numPunters)
	}

	fmt.Printf("<handshake>\n")

	for i := 0; i < *numPunters; i++ {
		punter := &punters[i]

		var rH recvHandshake
		if err := Recv(punter.reader, &rH); err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Welcome, %s!\n", rH.Name)

		if err := Send(punter.writer, sendHandshake{rH.Name}); err != nil {
			log.Fatal(err)
		}
	}

	fmt.Printf("</handshake>\n<setup>\n")

	for i := 0; i < *numPunters; i++ {
		punter := &punters[i]

		setup := sendSetup{
			uint64(i), uint64(*numPunters), mapData,
		}

		if err := Send(punter.writer, setup); err != nil {
			log.Fatal(err)
		}

		var rS recvSetup
		if err := Recv(punter.reader, &rS); err != nil {
			log.Fatal(err)
		}

		if rS.Ready != uint64(i) {
			log.Printf("[WARNING] Punter %d is very confused about it's identity.")
		}
	}

	log.Printf("</setup>\n")

	curPunter := 0
	sM := new(sendMove)
	sM.Move.Moves = make([]Move, *numPunters)

	for i, _ := range punters {
		sM.Move.Moves[i].Pass = &Pass{uint64(i)}
	}

	for curTurn := 0; curTurn < len(mapData.Rivers); curTurn++ {
		punter := &punters[curPunter]

		if err := Send(punter.writer, sM); err != nil {
			log.Fatal(err)
		}

		var rM recvMove
		if err := Recv(punter.reader, &rM); err != nil {
			log.Fatal(err)
		}

		fmt.Printf("[%d] received Move From %d: %+v\n", curTurn, curPunter, rM.Move)

		switch {
		case rM.Move.Claim != nil:
			claim := rM.Move.Claim

			ok := gMap.HasEdgeBetween(claim.Source, claim.Target)

			if !ok {
				fmt.Printf("[%d] claimed a river that doesn't exist! D:")
				goto failure // what is this shit
			} else {
				river := gMap.EdgeBetween(claim.Source, claim.Target).(*graph.MetadataEdge)

				if river.IsOwned {
					fmt.Printf("[%d] claimed a river that has already been claimed! D:")
					goto failure // this is what happens when you try to be fancy
				}

				river.IsOwned = true
				river.OwnerPunter = uint64(curPunter)

				sM.Move.Moves[curPunter].Claim = rM.Move.Claim
			}

			break
		failure:
			fallthrough
		default:
			sM.Move.Moves[curPunter].Pass = &Pass{uint64(curPunter)}

			break
		}
		sM.Move.Moves[curPunter] = rM.Move

		curPunter = (curPunter + 1) % *numPunters
	}

	sv := graph.Score(gMap, mapData.Mines, *numPunters)

	fmt.Printf("TALLIED %+v\n", sv)

	sS := sendStop{
		Stop{
			Moves:  sM.Move.Moves,
			Scores: sv,
		},
	}

	for i := 0; i < *numPunters; i++ {
		punter := &punters[i]

		if err := Send(punter.writer, sS); err != nil {
			log.Fatal(err)
		}
	}
}
