package main

import (
	"fmt"
	"github.com/adrs/shortestpath/graph"
	"image"
	"image/color"
	"image/gif"
	"io"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"strconv"
)

func inRange(cord graph.Cord, minLat, maxLat, minLong, maxLong int) bool {
	return cord.Lat >= minLat && cord.Lat <= maxLat && cord.Long >= minLong && cord.Long <= maxLong
}

// Finds vertex in graph closest to cordinate
func closestVertex(cord graph.Cord) int {
	closest := 0
	var minDist int64
	minDist = math.MaxInt64
	for i, v := range roadNetwork.Nodes {
		dist := graph.DistanceSquared(cord, v)
		if dist < minDist {
			minDist = dist
			closest = i
		}
	}
	return closest
}

func pixelLocation(cord graph.Cord, minLat, minLong, radius, size int) (x, y int) {
	x = int(float64((cord.Long-minLong)*size) / float64(2*radius))
	y = int(float64((cord.Lat-minLat)*size) / float64(2*radius))
	return
}

var palette = []color.Color{
	color.RGBA{255, 255, 255, 255}, // White
	color.RGBA{128, 128, 128, 255}, // Grey
	color.RGBA{0, 255, 0, 255},     // Green
	color.RGBA{255, 0, 0, 255},     // Red
	//color.RGBA{255, 128, 128, 255}, // Red
	//color.RGBA{0, 128, 0, 255},     // Green
}

const (
	backgroundColor = 0
	unvisitedColor  = 1
	visitedColor    = 2
	pathColor       = 3
)

func makeMap(centerx, centery, radius, size int) *image.Paletted {
	img := image.NewPaletted(image.Rect(0, 0, size, size), palette)
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			img.SetColorIndex(x, y, backgroundColor)
		}
	}
	minLat := centery - radius
	maxLat := centery + radius
	minLong := centerx - radius
	maxLong := centerx + radius
	for _, cord := range roadNetwork.Nodes {
		if !inRange(cord, minLat, maxLat, minLong, maxLong) {
			continue
		}
		x, y := pixelLocation(cord, minLat, minLong, radius, size)
		img.SetColorIndex(x, size-y, unvisitedColor)
	}
	return img
}

func drawMap(out io.Writer, centerx, centery, radius, size int) {
	img := gif.GIF{}
	img.Delay = append(img.Delay, 0)
	img.Image = append(img.Image, makeMap(centerx, centery, radius, size))
	gif.EncodeAll(out, &img)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Finds range of coordinate values specified by indices
func findCordinateRange(indices []int, cords []graph.Cord) (minLat, maxLat, minLong, maxLong int) {
	minLat, maxLat, minLong, maxLong = 180000000, -180000000, 180000000, -180000000
	for _, i := range indices {
		if cords[i].Lat > maxLat {
			maxLat = cords[i].Lat
		}
		if cords[i].Lat < minLat {
			minLat = cords[i].Lat
		}
		if cords[i].Long > maxLong {
			maxLong = cords[i].Long
		}
		if cords[i].Long < minLong {
			minLong = cords[i].Long
		}
	}
	return minLat, maxLat, minLong, maxLong
}

func copyImage(org *image.Paletted) *image.Paletted {
	copy := image.NewPaletted(org.Rect, org.Palette)
	for i, v := range org.Pix {
		copy.Pix[i] = v
	}
	return copy
}

// From: https://stackoverflow.com/questions/51626905/drawing-circles-with-two-radius-in-golang
func drawCircle(img *image.Paletted, x0, y0, r int, c uint8) {
	x, y, dx, dy := r-1, 0, 1, 1
	err := dx - (r * 2)

	for x > y {
		img.SetColorIndex(x0+x, y0+y, c)
		img.SetColorIndex(x0+y, y0+x, c)
		img.SetColorIndex(x0-y, y0+x, c)
		img.SetColorIndex(x0-x, y0+y, c)
		img.SetColorIndex(x0-x, y0-y, c)
		img.SetColorIndex(x0-y, y0-x, c)
		img.SetColorIndex(x0+y, y0-x, c)
		img.SetColorIndex(x0+x, y0-y, c)

		if err <= 0 {
			y++
			err += dy
			dy += 2
		}
		if err > 0 {
			x--
			dx += 2
			err += dx - (r * 2)
		}
	}
}

func drawShortestPath(out io.Writer, src, dest, size, frames, delay int) {
	// TODO: validate src and dest
	// TODO: cache searches
	landmarkPotential := func(v int) int {
		// max{d(L, t) - dist(L, v) for L in landmarks}
		maxDist := 0
		for _, distances := range landmarkDistances {
			// Unreachable -> dont want to deal with underflow
			if distances[dest] == math.MaxInt64 || distances[v] == math.MaxInt64 {
				continue
			}
			dist := distances[dest] - distances[v]
			if dist > maxDist {
				maxDist = dist
			}
		}
		return maxDist
	}
	shortestPath, searchSeq := graph.SearchSequence(roadNetwork, src, dest, landmarkPotential)

	//dijkstraPotential := func(int) int { return 0 }
	//shortestPath, searchSeq := graph.SearchSequence(roadNetwork, src, dest, dijkstraPotential)

	// Determine bounds from search sequence
	minLat, maxLat, minLong, maxLong := findCordinateRange(searchSeq, roadNetwork.Nodes)
	centerx := (minLong + maxLong) / 2
	centery := (minLat + maxLat) / 2
	dx := maxLong - minLong
	dy := maxLat - minLat
	radius := max(max(dx, dy)*11/20, 5e4)
	// TODO: add map boundaries
	// Resize image to remove space beyond map boundaries
	minLat = centery - radius
	maxLat = centery + radius
	minLong = centerx - radius
	maxLong = centerx + radius

	// Generate animation
	anim := gif.GIF{}
	img := makeMap(centerx, centery, radius, size)
	stepsPerFrame := len(searchSeq)
	if frames > 1 {
		stepsPerFrame /= (frames - 1)
	}

	// Draws circles at src and dest
	drawEndpoint := func(img *image.Paletted, v, r int) {
		cord := roadNetwork.Nodes[v]
		x, y := pixelLocation(cord, minLat, minLong, radius, size)
		drawCircle(img, x, size-y, r, pathColor)
	}

	// Show search sequence
	for i, v := range searchSeq {
		cord := roadNetwork.Nodes[v]
		if !inRange(cord, minLat, maxLat, minLong, maxLong) {
			continue
		}
		x, y := pixelLocation(cord, minLat, minLong, radius, size)
		img.SetColorIndex(x, size-y, visitedColor)

		if i%stepsPerFrame == 0 && frames > 1 {
			drawEndpoint(img, src, 3)
			drawEndpoint(img, dest, 3)
			anim.Image = append(anim.Image, copyImage(img))
			anim.Delay = append(anim.Delay, delay)
		}
	}

	// Show shortest path
	for _, v := range shortestPath {
		cord := roadNetwork.Nodes[v]
		if !inRange(cord, minLat, maxLat, minLong, maxLong) {
			continue
		}
		x, y := pixelLocation(cord, minLat, minLong, radius, size)
		img.SetColorIndex(x, size-y, pathColor)
	}
	drawEndpoint(img, src, 5)
	drawEndpoint(img, dest, 5)
	anim.Image = append(anim.Image, img)
	anim.Delay = append(anim.Delay, delay*7)
	gif.EncodeAll(out, &anim)
}

var roadNetwork *graph.Graph
var landmarks []int
var landmarkDistances [][]int

func setup(nodeFilePath, vertexFilePath string) {
	log.Print("Loading graph...")
	g, err := graph.LoadGraph(nodeFilePath, vertexFilePath)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Picking landmarks...")
	landmarks = graph.PickLandmarks(g, 1)
	log.Print("Computing distances to landmarks...")
	landmarkDistances = graph.DistancesFromLandmarks(g, landmarks)
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
		return int(defaultValue * 1e6)
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
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "usage: %s <node file> <vertex file>\n", os.Args[0])
		os.Exit(1)
	}
	rand.Seed(42)
	setup(os.Args[1], os.Args[2])
	http.Handle("/", http.FileServer(http.Dir("static")))
	http.HandleFunc("/map", func(w http.ResponseWriter, r *http.Request) {
		centerx := parseCordPart(r.FormValue("centerx"), -180, 180, -85)
		centery := parseCordPart(r.FormValue("centery"), -180, 180, 44)
		radius := parseCordPart(r.FormValue("radius"), 0.01, 90, 5)
		size := parseInt(r.FormValue("size"), 24, 2000, 400)
		drawMap(w, centerx, centery, radius, size)
	})
	http.HandleFunc("/shortest-path", func(w http.ResponseWriter, r *http.Request) {
		maxIdx := len(roadNetwork.Nodes)
		src := parseInt(r.FormValue("src"), 1, maxIdx, rand.Intn(maxIdx)+1) - 1
		dest := parseInt(r.FormValue("dest"), 1, maxIdx, rand.Intn(maxIdx)+1) - 1
		size := parseInt(r.FormValue("size"), 24, 2000, 400)
		frames := parseInt(r.FormValue("frames"), 1, 120, 15)
		delay := parseInt(r.FormValue("delay"), 0, 2000, 500) / 10
		// Browsers ignore loop count field in gifs :(
		drawShortestPath(w, src, dest, size, frames, delay)
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
	http.HandleFunc("/closest-vertex", func(w http.ResponseWriter, r *http.Request) {
		x := parseCordPart(r.FormValue("x"), -180, 180, -85)
		y := parseCordPart(r.FormValue("y"), -180, 180, 44)
		id := closestVertex(graph.Cord{Lat: y, Long: x})
		lat, long := roadNetwork.Nodes[id].Lat, roadNetwork.Nodes[id].Long
		w.Header().Add("Content-Type", "application/json")
		fmt.Fprintf(w, "{\"Lat\": %d, \"Long\": %d, \"NodeId\": %d}", lat, long, id)
	})
	log.Print("Starting server...")
	log.Fatal(http.ListenAndServe("localhost:8888", nil))
}
