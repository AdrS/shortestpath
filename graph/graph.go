package graph

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
)

type Cord struct {
	Lat  int
	Long int
}

/*
func getDegreesMinutesSeconds(cord int) (degrees int, minutes int, seconds float64) {
	degrees = cord/1e6
	left := cord % 1e6
	minutes = int(float64(left)/1e6*60)
	left -= minutes
} */

func (c Cord) String() string {
	lat := float64(c.Lat) / 1e6
	long := float64(c.Long) / 1e6
	var latDir, longDir string
	if lat < 0 {
		latDir = "S"
	} else {
		latDir = "N"
	}
	if long < 0 {
		longDir = "W"
	} else {
		longDir = "E"
	}
	return fmt.Sprintf("%f %s %f %s", math.Abs(lat), latDir, math.Abs(long), longDir)
}

type Dest struct {
	Dest int
	Dist int
}

type Graph struct {
	Nodes          []Cord
	AdjacencyLists [][]Dest
}

func loadCords(in io.Reader) ([]Cord, error) {
	var idx, lat, long int
	cords := make([]Cord, 0)
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "v ") {
			if _, err := fmt.Sscanf(line, "v %d %d %d", &idx, &long, &lat); err != nil {
				return nil, err
			}
			cords = append(cords, Cord{lat, long})
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return cords, nil
}

func LoadGraph(cordFile, arcFile string) (*Graph, error) {
	cf, err := os.Open(cordFile)
	if err != nil {
		return nil, err
	}
	af, err := os.Open(arcFile)
	if err != nil {
		return nil, err
	}

	// Read cordinates
	cords, err := loadCords(cf)
	if err != nil {
		return nil, err
	}

	// Set up adjacency lists
	adjLists := make([][]Dest, 0)
	for _ = range cords {
		adjLists = append(adjLists, make([]Dest, 0))
	}

	var src, dest, dist int
	scanner := bufio.NewScanner(af)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "a ") {
			if _, err := fmt.Sscanf(line, "a %d %d %d", &src, &dest, &dist); err != nil {
				return nil, err
			}
			// convert to 0 based indexing
			src--
			dest--
			if src >= len(cords) || src < 0 || dest >= len(cords) || dest < 0 {
				return nil, errors.New("invalid index")
			}

			adjLists[src] = append(adjLists[src], Dest{dest, dist})
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return &Graph{Nodes: cords, AdjacencyLists: adjLists}, nil
}
