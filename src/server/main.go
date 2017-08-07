package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/jemoster/icfp2017/src/protocol"
	"io/ioutil"
	"log"
	"os"
	"path"
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
	mapPath := flag.String("map", path.Join("..", "..", "maps", "sample.json"), "A file containing a JSON Map object")
	srvPort := flag.Int("port", 9001, "The port to listen on")
	numPunters := flag.Int("punters", 2, "Number of players")

	splurges := flag.Bool("splurges", true, "to disable splurges use --splurges=false")
	options := flag.Bool("options", true, "to disable options use --options=false")

	runOnce := flag.Bool("runonce", false, "to run only one session use --runonce=true")
	resultsDir := flag.String("results", "results", "directory in which to place log files.")

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

	if _, err := os.Stat(*resultsDir); os.IsNotExist(err) {
		os.Mkdir(*resultsDir, os.ModePerm)
	}

	resultsFileName := path.Join(*resultsDir, fmt.Sprintf("%s.log", path.Base(*mapPath)))

	resultsFile, err := os.OpenFile(resultsFileName, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	defer resultsFile.Close()

	resultsWriter := bufio.NewWriter(resultsFile)
	defer resultsWriter.Flush()

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

	serv.run(resultsWriter)
}
