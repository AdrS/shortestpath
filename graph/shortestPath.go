package graph

import (
	"log"
	"math"
	"math/rand"
)

type NodeSearchState struct {
	// Length of shortest path found (so far) from s
	Distance int
	// Estimate of distance from vertex to t (only needs to be computed once)
	Potential int
	// Predecessor of vertex on shortest path from s
	Pred int
	// Have all the edges from this vertex been relaxed yet
	Processed bool
	// Index in heap
	Idx int
}

type SearchState struct {
	Nodes []NodeSearchState
	Heap  []int
}

func (s *SearchState) checkInvariants() {
	for i, node := range s.Nodes {
		if node.Processed {
			if node.Idx != -1 {
				log.Fatalf("Processed node %d should not be in heap %v, %v\n", i, node, s.Heap)
			}
			if node.Distance == math.MaxInt64 {
				log.Fatalf("Processed node %d should be reachable %v, %v\n", i, node, s.Heap)
			}
		}
		if node.Idx != -1 {
			if node.Idx < 0 || node.Idx >= len(s.Heap) {
				log.Fatalf("Invalid heap index %d, %v, %v\n", i, node, s.Heap)
			}
			if s.Heap[node.Idx] != i {
				log.Fatalf("Heap element and node %d do not match %d %v, %v\n", i, s.Heap[node.Idx], node, s.Heap)
			}
		}
	}
	for idx, i := range s.Heap {
		if i < 0 || i >= len(s.Nodes) {
			log.Fatalf("Invalid node index %d, %v\n", i, s.Heap)
		}
		if s.Nodes[i].Idx != idx {
			log.Fatalf("Heap element and node do not match %d %d %v, %v\n", idx, i, s.Nodes[i], s.Heap)
		}
	}
}

func (s *SearchState) Relax(u, v, distance int) {
	if distance < s.Nodes[v].Distance {
		s.Nodes[v].Distance = distance
		s.Nodes[v].Pred = u
		if s.Nodes[v].Idx == -1 {
			Push(s, v)
		} else {
			Fix(s, s.Nodes[v].Idx)
		}
	}
}

func (s *SearchState) Len() int {
	return len(s.Heap)
}

func (s *SearchState) Less(a, b int) bool {
	u := s.Nodes[s.Heap[a]]
	v := s.Nodes[s.Heap[b]]
	return u.Distance+u.Potential <= v.Distance+v.Potential
}

func (s *SearchState) Swap(a, b int) {
	s.Heap[a], s.Heap[b] = s.Heap[b], s.Heap[a]
	s.Nodes[s.Heap[a]].Idx = a
	s.Nodes[s.Heap[b]].Idx = b
}

func (s *SearchState) Push(x int) {
	s.Nodes[x].Idx = len(s.Heap)
	s.Heap = append(s.Heap, x)
}

func (s *SearchState) Pop() int {
	n := len(s.Heap)
	last := s.Heap[n-1]
	s.Heap = s.Heap[0 : n-1]
	s.Nodes[last].Idx = -1
	return last
}

// Adapted from standard library "container/heap"
func Push(s *SearchState, x int) {
	s.Push(x)
	up(s, s.Len()-1)
}

// Pop removes the minimum element (according to Less) from the heap
// and returns it. The complexity is O(log(n)) where n = h.Len().
// It is equivalent to Remove(h, 0).
func Pop(s *SearchState) int {
	n := s.Len() - 1
	s.Swap(0, n)
	down(s, 0, n)
	return s.Pop()
}

// Fix re-establishes the heap ordering after the element at index i has changed its value.
// Changing the value of the element at index i and then calling Fix is equivalent to,
// but less expensive than, calling Remove(h, i) followed by a Push of the new value.
// The complexity is O(log(n)) where n = h.Len().
func Fix(s *SearchState, i int) {
	if !down(s, i, s.Len()) {
		up(s, i)
	}
}

func up(s *SearchState, j int) {
	for {
		i := (j - 1) / 2 // parent
		if i == j || !s.Less(j, i) {
			break
		}
		s.Swap(i, j)
		j = i
	}
}

func down(s *SearchState, i0, n int) bool {
	i := i0
	for {
		j1 := 2*i + 1
		if j1 >= n || j1 < 0 { // j1 < 0 after int overflow
			break
		}
		j := j1 // left child
		if j2 := j1 + 1; j2 < n && s.Less(j2, j1) {
			j = j2 // = 2*i + 2  // right child
		}
		if !s.Less(j, i) {
			break
		}
		s.Swap(i, j)
		i = j
	}
	return i > i0
}

func NewSearchState(size int) *SearchState {
	s := SearchState{
		Nodes: make([]NodeSearchState, 0, size),
		Heap:  make([]int, 0, size),
	}
	for i := 0; i < size; i++ {
		s.Nodes = append(s.Nodes, NodeSearchState{
			Distance:  math.MaxInt64,
			Potential: -1,
			Pred:      -1,
			Processed: false,
			Idx:       -1,
		})
	}
	return &s
}

type PotentialFunc func(v int) int

// Potential functions:
// - function pi: V -> R
// - transforms edge weights from w(u, v) -> w_{pi}(u, v) = w(u, v) - pi(u) + pi(v)
// - for any path P from s to t, l_{pi}(P) = l(P) - pi(s) + pi(t) (telescoping sum)
// - => transformation does not change shortest paths (does change lengths)
// - d_pi(s, u) = d(s, u) - pi(s) + pi(u)
// - potential must be feasible: w_{pi}(u, v) >= 0 for all u, v (otherwise we cannot run Dijkstra)
//
// - pick potential that is lower bound on d(u, t)
//
// - A* = Dijkstra where next v is chosen to minimize dist(v) + pi(v) [lower bound of d(v,t)]
//
// Set all potentials to 0 to use Dijkstra's

// Landmarks:
// Triangle inequality d(a, b) + d(b, c) >= d(a,c)
// For landmark L
// - d(L, u) + d(u, v) >= d(L, v)
//     => d(u, v) >= d(L, v) - d(L, u)
// Improve lower bound:
// - d(u, v) >= max{d(L, v) - d(L, u), d(v, L) - dist(u, L)}
// - take maximum over multiple landmarks
//
// pi(v) = d(L, t) - d(L, v) is feasible
// pf: l(u, v) - pi(u) + pi(v) = l(u, v) - d(L, t) + d(L, u) + d(L, t) - d(L, v)
//      = d(L, u) + l(u, v) - d(L, v) >= 0 bc d(L, v) <= d(L, u) + l(u, v) by triangle inequality

// Runs a shortest path algorithm from source to dest
// and returns the reverse of the shortest path and
// the sequence of vertices visited
func SearchSequence(graph *Graph, src, dest int, potential PotentialFunc) ([]int, []int) {
	state := NewSearchState(len(graph.Nodes))
	vistSeq := make([]int, 0)

	state.Relax(-1, src, 0)

	for state.Len() != 0 {
		// Find closest unprocessed reachable vertex
		u := Pop(state)

		vistSeq = append(vistSeq, u)

		if u == dest {
			// log.Printf("Found: %d, expected: %d\n", state.Nodes[u].Distance, Dijkstra(graph, src)[dest])
			break
		}

		// Relax edges leaving u
		state.Nodes[u].Processed = true
		for _, dest := range graph.AdjacencyLists[u] {
			v := dest.Dest
			if !state.Nodes[v].Processed {
				// Lazily compute potentials
				if state.Nodes[v].Potential == -1 {
					state.Nodes[v].Potential = potential(v)
				}
				state.Relax(u, v, state.Nodes[u].Distance+dest.Dist)
			}
		}
	}

	// Reconstruct shortest path if vertex is destination is reachable
	shortestPath := make([]int, 0)
	if state.Nodes[dest].Distance != -1 {
		cur := dest
		for cur != src {
			shortestPath = append(shortestPath, cur)
			cur = state.Nodes[cur].Pred
		}
		if src != dest {
			shortestPath = append(shortestPath, src)
		}
	}

	return shortestPath, vistSeq
}

// Computes distances from src vertex to every other vertex in graph
func Dijkstra(graph *Graph, src int) []int {
	state := NewSearchState(len(graph.Nodes))
	vistSeq := make([]int, 0)

	state.Relax(-1, src, 0)

	for state.Len() != 0 {
		// Find closest unprocessed reachable vertex
		u := Pop(state)

		vistSeq = append(vistSeq, u)

		// Relax edges leaving u
		state.Nodes[u].Processed = true
		for _, dest := range graph.AdjacencyLists[u] {
			v := dest.Dest
			if !state.Nodes[v].Processed {
				state.Relax(u, dest.Dest, dest.Dist+state.Nodes[u].Distance)
			}
		}
	}
	distances := make([]int, len(graph.Nodes))
	for i, n := range state.Nodes {
		distances[i] = n.Distance
	}
	return distances
}

// TODO: use map[string]->function pattern

// Returns list of landmarks and distances for landmarks to every point
func PickRandomLandmarks(graph *Graph, n int) []int {
	// Start with random selection
	landmarks := make([]int, 0)
	for i := 0; i < n; i++ {
		landmarks = append(landmarks, rand.Intn(len(graph.Nodes)))
	}
	return landmarks
}

// Returns distance from landmark i to each vertex
func DistancesFromLandmarks(graph *Graph, landmarks []int) [][]int {
	distances := make([][]int, len(landmarks))
	for i, landmark := range landmarks {
		distances[i] = Dijkstra(graph, landmark)
	}
	return distances
}

// Find how many hops every element is from src
func bfs(graph *Graph, src int) []int {
	distances := make([]int, len(graph.Nodes))
	for i := range distances {
		distances[i] = math.MaxInt64
	}
	queue := make([]int, 0)
	queue = append(queue, src)
	distances[src] = 0
	for len(queue) > 0 {
		u := queue[0]
		queue = queue[1:]
		for _, dest := range graph.AdjacencyLists[u] {
			v := dest.Dest
			// if undiscovered add to queue
			if distances[v] == math.MaxInt64 {
				distances[v] = distances[u] + 1
				queue = append(queue, v)
			}
		}
	}
	return distances
}

func PickFarthestLandmarks(graph *Graph, n int) []int {
	// Start with random landmark
	landmarks := make([]int, 0)
	landmarks = append(landmarks, rand.Intn(len(graph.Nodes)))

	distanceFromSet := bfs(graph, landmarks[0])

	// Pick node greatest number of hops from previously picked landmarks as next
	for i := 1; i < n; i++ {
		maxDist := 0
		next := 0
		for j, dist := range distanceFromSet {
			if dist > maxDist {
				maxDist = dist
				next = j
			}
		}
		landmarks = append(landmarks, next)
		// Update node distances from set
		for j, dist := range bfs(graph, next) {
			if dist < distanceFromSet[j] {
				distanceFromSet[j] = dist
			}
		}
	}
	return landmarks
}
