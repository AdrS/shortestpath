package graph

import (
	"log"
	"math"
)

type NodeSearchState struct {
	// Length of shortest path (so far) to node
	Distance int
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
	return s.Nodes[s.Heap[a]].Distance < s.Nodes[s.Heap[b]].Distance
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
			Pred:      -1,
			Processed: false,
			Idx:       -1,
		})
	}
	return &s
}

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

// Runs a shortest path algorithm from source to dest
// and returns the reverse of the shortest path and
// the sequence of vertices visited
func SearchSequence(graph *Graph, src, dest int, potential []int) ([]int, []int) {
	state := NewSearchState(len(graph.Nodes))
	vistSeq := make([]int, 0)

	state.Relax(-1, src, 0)

	for state.Len() != 0 {
		// Find closest unprocessed reachable vertex
		u := Pop(state)

		vistSeq = append(vistSeq, u)

		if u == dest {
			break
		}

		// Relax edges leaving u
		state.Nodes[u].Processed = true
		for _, dest := range graph.AdjacencyLists[u] {
			v := dest.Dest
			if !state.Nodes[v].Processed {
				state.Relax(u, dest.Dest, dest.Dist+state.Nodes[u].Distance+potential[u])
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
