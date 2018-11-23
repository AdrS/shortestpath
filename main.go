package main

import (
	"fmt"
	"image"
	"image/color"
	"io"
	"strconv"
	//"image/gif"
	"github.com/adrs/shortestpath/graph"
	"image/png"
	"log"
	"net/http"
)

func drawMap(out io.Writer, centerx, centery, radius, size int) {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			img.Set(x, y, color.RGBA{255, 255, 255, 255})
		}
	}
	minLat := centery - radius
	maxLat := centery + radius
	minLong := centerx - radius
	maxLong := centerx + radius
	for _, cord := range roadNetwork.Nodes {
		if cord.Lat < minLat || cord.Lat > maxLat || cord.Long < minLong || cord.Long > maxLong {
			continue
		}
		x := int(float64((cord.Long-minLong)*size) / float64(2*radius))
		y := int(float64((cord.Lat-minLat)*size) / float64(2*radius))
		img.Set(x, size-y, color.RGBA{128, 128, 128, 255})
	}
	png.Encode(out, img)
}

var roadNetwork *graph.Graph

func setup() {
	log.Print("Loading graph...")
	base := "/mnt/c/Users/adrian/Documents/EECS477/project/data/"
	g, err := graph.LoadGraph(base+"USA-road-d.LKS.co", base+"USA-road-t.LKS.gr")
	if err != nil {
		log.Fatal(err)
	}
	roadNetwork = g
}

// Parses integer, ensuring result is in [min, max]
// Provides default value of s is not an integer
func parseInt(s string, min, max, defaultValue int) int {
	x, err := strconv.Atoi(s)
	if err != nil {
		x = defaultValue
	}
	if x < min {
		x = min
	}
	if x > max {
		x = max
	}
	return x
}

// takes part of a lat long cordinate and parses it
func parseCordPart(s string, min, max, defaultValue float64) int {
	x, err := strconv.ParseFloat(s, 64)
	if err != nil {
		x = defaultValue
	}
	if x < min {
		x = min
	}
	if x > max {
		x = max
	}
	return int(x * 1e6)
}

func main() {
	setup()
	http.Handle("/", http.FileServer(http.Dir("static")))
	http.HandleFunc("/map", func(w http.ResponseWriter, r *http.Request) {
		centerx := parseCordPart(r.FormValue("centerx"), -180, 180, -85)
		centery := parseCordPart(r.FormValue("centery"), -180, 180, 44)
		radius := parseCordPart(r.FormValue("radius"), 0.01, 90, 5)
		size := parseInt(r.FormValue("size"), 24, 2000, 400)
		drawMap(w, centerx, centery, radius, size)
	})
	http.HandleFunc("/vertex", func(w http.ResponseWriter, r *http.Request) {
		i, err := strconv.Atoi(r.FormValue("i"))
		if err != nil {
			fmt.Fprintf(w, "error: invalid \"i\" parameter")
			return
		}
		i--
		if i < 0 || i >= len(roadNetwork.Nodes) {
			fmt.Fprintf(w, "error: index out of range")
			return
		}
		fmt.Fprintf(w, "Cordinates: %s\n", roadNetwork.Nodes[i])
		fmt.Fprintf(w, "Edges:")
		for _, dest := range roadNetwork.AdjacencyLists[i] {
			fmt.Fprintf(w, "\tDestination: %d at %s, Distance: %d\n", dest.Dest+1, roadNetwork.Nodes[dest.Dest], dest.Dist)
		}
	})
	log.Print("Starting server...")
	log.Fatal(http.ListenAndServe("localhost:8888", nil))
}
