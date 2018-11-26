package graph

import (
	//"log"
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
				//log.Printf("Processed node %d should not be in heap %v, %v\n", i, node, s.Heap)
			}
			if node.Distance == math.MaxInt64 {
				//log.Printf("Processed node %d should be reachable %v, %v\n", i, node, s.Heap)
			}
		}
		if node.Idx != -1 {
			if node.Idx < 0 || node.Idx >= len(s.Heap) {
				//log.Printf("Invalid heap index %d, %v, %v\n", i, node, s.Heap)
			}
			if s.Heap[node.Idx] != i {
				//log.Printf("Heap element and node %d do not match %d %v, %v\n", i, s.Heap[node.Idx], node, s.Heap)
			}
		}
	}
	for idx, i := range s.Heap {
		if i < 0 || i >= len(s.Nodes) {
			//log.Printf("Invalid node index %d, %v\n", i, s.Heap)
		}
		if s.Nodes[i].Idx != idx {
			//log.Printf("Heap element and node do not match %d %d %v, %v\n", idx, i, s.Nodes[i], s.Heap)
		}
	}
}

func (s *SearchState) Relax(u, v, distance int) {
	//s.checkInvariants()
	//log.Printf("Relax %d: %v -> %d\n", v, s.Nodes[v], distance)
	if distance < s.Nodes[v].Distance {
		//log.Println("updating distance")
		s.Nodes[v].Distance = distance
		s.Nodes[v].Pred = u
		if s.Nodes[v].Idx == -1 {
			Push(s, v)
		} else {
			Fix(s, s.Nodes[v].Idx)
		}
	}
	//s.checkInvariants()
}

func (s *SearchState) Len() int {
	//s.checkInvariants()
	return len(s.Heap)
}

func (s *SearchState) Less(a, b int) bool {
	//s.checkInvariants()
	return s.Nodes[s.Heap[a]].Distance < s.Nodes[s.Heap[b]].Distance
}

func (s *SearchState) Swap(a, b int) {
	//s.checkInvariants()
	s.Heap[a], s.Heap[b] = s.Heap[b], s.Heap[a]
	s.Nodes[s.Heap[a]].Idx = a
	s.Nodes[s.Heap[b]].Idx = b
	//s.checkInvariants()
}

func (s *SearchState) Push(x int) {
	//s.checkInvariants()
	//log.Println("Push", x)
	s.Nodes[x].Idx = len(s.Heap)
	s.Heap = append(s.Heap, x)
	//s.checkInvariants()
}

func (s *SearchState) Pop() int {
	//s.checkInvariants()
	n := len(s.Heap)
	last := s.Heap[n-1]
	s.Heap = s.Heap[0 : n-1]
	s.Nodes[last].Idx = -1
	//s.checkInvariants()
	return last
}

// Adapted from standard library "container/heap"
func Init(s *SearchState) {
	// heapify
	n := s.Len()
	for i := n/2 - 1; i >= 0; i-- {
		down(s, i, n)
	}
	//s.checkInvariants()
}

func Push(s *SearchState, x int) {
	//s.checkInvariants()
	s.Push(x)
	up(s, s.Len()-1)
	//s.checkInvariants()
}

// Pop removes the minimum element (according to Less) from the heap
// and returns it. The complexity is O(log(n)) where n = h.Len().
// It is equivalent to Remove(h, 0).
func Pop(s *SearchState) int {
	//s.checkInvariants()
	n := s.Len() - 1
	s.Swap(0, n)
	down(s, 0, n)
	//s.checkInvariants()
	return s.Pop()
}

// Remove removes the element at index i from the heap.
// The complexity is O(log(n)) where n = h.Len().
func Remove(s *SearchState, i int) int {
	n := s.Len() - 1
	if n != i {
		s.Swap(i, n)
		if !down(s, i, n) {
			up(s, i)
		}
	}
	return s.Pop()
}

// Fix re-establishes the heap ordering after the element at index i has changed its value.
// Changing the value of the element at index i and then calling Fix is equivalent to,
// but less expensive than, calling Remove(h, i) followed by a Push of the new value.
// The complexity is O(log(n)) where n = h.Len().
func Fix(s *SearchState, i int) {
	//s.checkInvariants()
	if !down(s, i, s.Len()) {
		up(s, i)
	}
	//s.checkInvariants()
}

func up(s *SearchState, j int) {
	//s.checkInvariants()
	for {
		i := (j - 1) / 2 // parent
		if i == j || !s.Less(j, i) {
			break
		}
		s.Swap(i, j)
		j = i
	}
	//s.checkInvariants()
}

func down(s *SearchState, i0, n int) bool {
	//s.checkInvariants()
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
	//s.checkInvariants()
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
	//s.checkInvariants()
	return &s
}

// Runs a shortest path algorithm from source to dest
// and returns the reverse of the shortest path and
// the sequence of vertices visited
func SearchSequence(graph *Graph, src, dest int) ([]int, []int) {
	//log.Printf("Searching for path from %d to %d...\n", src, dest)

	// Use Dijkstra's algorithm
	state := NewSearchState(len(graph.Nodes))

	vistSeq := make([]int, 0)

	state.Relax(-1, src, 0)
	//log.Println(state.Heap, state.Nodes[src], state.Len())

	for state.Len() != 0 {
		//log.Printf("Loop %v\n", state.Heap)
		// Find closest unprocessed reachable vertex
		u := Pop(state)
		//log.Println("Processing", u)

		vistSeq = append(vistSeq, u)

		if u == dest {
			break
		}

		// Relax edges leaving u
		state.Nodes[u].Processed = true
		if state.Nodes[u].Idx != -1 {
			//log.Fatal("should not be here")
		}
		for _, dest := range graph.AdjacencyLists[u] {
			v := dest.Dest
			//log.Print(state.Nodes[v])
			if !state.Nodes[v].Processed {
				state.Relax(u, dest.Dest, dest.Dist+state.Nodes[u].Distance)
			}
		}
	}
	//log.Printf("Done with search\n")

	// Reconstruct shortest path if vertex is destination is reachable
	shortestPath := make([]int, 0)
	if state.Nodes[dest].Distance != -1 {
		cur := dest
		for cur != src {
			//log.Printf("cur: %d %v\n", cur, state.Nodes[cur])
			shortestPath = append(shortestPath, cur)
			cur = state.Nodes[cur].Pred
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
