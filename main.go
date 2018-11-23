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

func drawMap(out io.Writer) {
	width := 100
	height := 100
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{0, 0, 255, 255})
		}
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

func main() {
	setup()
	http.HandleFunc("/map", func(w http.ResponseWriter, r *http.Request) {
		drawMap(w)
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
		// TODO: show outbound edges
		//fmt.Fprintf(w, "Edges:")
	})
	log.Print("Starting server...")
	log.Fatal(http.ListenAndServe("localhost:8888", nil))
}
