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
		return nil, fmt.Errorf("Failed to read map: %v", err)
	}

	m := &protocol.Map{}
	if err := json.Unmarshal(b, m); err != nil {
		return nil, fmt.Errorf("Failed to unmarshal map: %v", err)
	}

	return m, nil
}

func main() {
	m, err := readMap()
	if err != nil {
		log.Printf("Failed to read map: %v", err)
		os.Exit(1)
	}

	fmt.Printf("Map: %+v\n", m)
}
