package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/jemoster/icfp2017/src/graph"
	"github.com/jemoster/icfp2017/src/protocol"
)

var debug = flag.Bool("vase", false, "verbose logging")

func debugf(format string, v ...interface{}) {
	if *debug {
		log.Printf(format, v...)
	}
}

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

// distances contains the distance (value) to a target (key) for one source.
type distances map[protocol.SiteID]uint64

func FindMax(d graph.Distances) (mine1 protocol.SiteID, site1 protocol.SiteID, maxDist1 uint64, 
                                 mine2 protocol.SiteID, mine3 protocol.SiteID, maxDist2 uint64) {

    for mine, sites := range d {
        for site, distance := range sites {
            if distance > maxDist1 {
                maxDist1 = distance
                mine1 = mine
                site1 = site
            }
            
            if _, ok := d[site]; ok {
                if distance > maxDist2 {
                    maxDist2 = distance
                    mine2 = mine
                    mine3 = site
                }
            }
        }
    }
    
    return
}

func ComputePoints(dist uint64) (uint64) {
    return dist * dist * dist / 3 + dist * dist / 2 + dist / 6
}


func main() {
	flag.Parse()

	m, err := readMap()
	if err != nil {
		log.Fatalf("Failed to read map: %v", err)
	}

	debugf("Parsed map: %+v\n", m)

	g := graph.New(m, func(*graph.MetadataEdge) float64 {
		return 1.0
	})

	debugf("Graph: %+v\n", g)

	results := g.ShortestDistances(m.Mines)
    
	debugf("Results: %+v\n", results)

	b, err := json.Marshal(results)
	if err != nil {
		log.Fatalf("Failed to marshal results: %v", err)
	}

	fmt.Println(string(b))
    
    mine1, site1, maxDist1, mine2, mine3, maxDist2 := FindMax(results)
    
    fmt.Println(mine1, site1, maxDist1, mine2, mine3, maxDist2)
    
    longest := ComputePoints(maxDist1)
    mine2mine := ComputePoints(maxDist2) * 2
    if longest < mine2mine {
        fmt.Printf("Should walk from mine %d to mine %d, worth %d more points\n",
                    mine2, mine3, mine2mine - longest)
    }
}
