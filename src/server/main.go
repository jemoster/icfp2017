package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/jemoster/icfp2017/src/protocol"
	"io/ioutil"
	"log"
	"os"
)

func loadMap(path string) (*protocol.Map, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	m := new(protocol.Map)

	err = json.Unmarshal(data, m)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func main() {
	mapPath := flag.String("map", "", "[required] A JSON file")
	srvPort := flag.Int("port", 9001, "The port to listen on")
	numPunters := flag.Int("punters", 2, "Number of players")

	splurges := flag.Bool("splurges", true, "to disable splurges use --splurges=false")
	options := flag.Bool("options", true, "to disable options use --options=false")
	
	runOnce := flag.Bool("runonce", false, "to run only one session use --runonce=true")

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

	serv := Server{
		Map:        *mapData,
		Port:       *srvPort,
		NumPunters: *numPunters,
		Settings: protocol.Settings{
			Futures:  false,
			Splurges: *splurges,
			Options:  *options,
		},
		RunOnce: *runOnce,
	}

	serv.run()
}
