package graph

import (
	"math"
)

type nodeState struct {
	// Length of shortest path yet to node
	distance int
	// Predecessor of vertex on shortest path
	pred      int
	processed bool
}

// Runs a shortest path algorithm from source to dest
// and returns the reverse of the shortest path and
// the sequence of vertices visited
func SearchSequence(graph *Graph, src, dest int) ([]int, []int) {
	// Use Dijkstra's algorithm
	// TODO: use heap for faster implementation
	states := make([]nodeState, 0, len(graph.Nodes))
	for _ = range graph.Nodes {
		states = append(states, nodeState{math.MaxInt64, -1, false})
	}
	states[src].distance = 0

	shortestPath := make([]int, 0)
	vistSeq := make([]int, 0)
	for {
		// Find closest unprocessed vertex
		minDist := math.MaxInt64
		u := -1
		for i, state := range states {
			if !state.processed && state.distance < minDist {
				minDist = state.distance
				u = i
			}
		}
		// No more unprocessed vertices
		if u == -1 || u == dest {
			break
		}

		// Relax edges leaving u
		states[u].processed = true
		vistSeq = append(vistSeq, u)
		for _, dest := range graph.AdjacencyLists[u] {
			dist := dest.Dist + states[u].distance
			if dist < states[dest.Dest].distance {
				states[dest.Dest].distance = dist
				states[dest.Dest].pred = u
			}
		}
	}
	// Reconstruct shortest path if vertex is destination is reachable
	if states[dest].distance != -1 {
		cur := dest
		for cur != src {
			shortestPath = append(shortestPath, cur)
			cur = states[cur].pred
		}
		if src != dest {
			shortestPath = append(shortestPath, src)
		}
	}
	return shortestPath, vistSeq

	// TODO: use A* and ALT

	// TODO:
	// create map: vertex -> vertexState (only store state for vertices we've reached)
	// heap: discovered vertices
}
