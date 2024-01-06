package dag

import (
	"context"
	"sort"
	"sync"

	llq "github.com/emirpasic/gods/queues/linkedlistqueue"
	lls "github.com/emirpasic/gods/stacks/linkedliststack"
)

// Visitor is the interface that wraps the basic Visit method.
// It can use the Visitor and XXXWalk functions together to traverse the entire DAG.
// And access per-vertex information when traversing.
type Visitor interface {
	Visit(context.Context, Vertexer)
}

// DFSWalk implements the Depth-First-Search algorithm to traverse the entire DAG.
// The algorithm starts at the root node and explores as far as possible
// along each branch before backtracking.
func (d *DAG) DFSWalk(ctx context.Context, visitor Visitor) {
	d.muDAG.RLock()
	defer d.muDAG.RUnlock()

	stack := lls.New()

	vertices := d.getRoots()
	for _, id := range reversedVertexIDs(vertices) {
		v := vertices[id]
		sv := storableVertex{WrappedID: id, Value: v}
		stack.Push(sv)
	}

	visited := make(map[string]bool, d.getSize())

	for !stack.Empty() {
		v, _ := stack.Pop()
		sv := v.(storableVertex)

		if !visited[sv.WrappedID] {
			visited[sv.WrappedID] = true
			visitor.Visit(ctx, &sv)
		}

		vertices, _ := d.getChildren(sv.WrappedID)
		for _, id := range reversedVertexIDs(vertices) {
			v := vertices[id]
			sv := storableVertex{WrappedID: id, Value: v}
			stack.Push(sv)
		}
	}
}

// BFSWalk implements the Breadth-First-Search algorithm to traverse the entire DAG.
// It starts at the tree root and explores all nodes at the present depth prior
// to moving on to the nodes at the next depth level.
func (d *DAG) BFSWalk(ctx context.Context, visitor Visitor) {
	d.muDAG.RLock()
	defer d.muDAG.RUnlock()

	queue := llq.New()

	vertices := d.getRoots()
	for _, id := range vertexIDs(vertices) {
		v := vertices[id]
		sv := storableVertex{WrappedID: id, Value: v}
		queue.Enqueue(sv)
	}

	visited := make(map[string]bool, d.getOrder())

	for !queue.Empty() {
		v, _ := queue.Dequeue()
		sv := v.(storableVertex)

		if !visited[sv.WrappedID] {
			visited[sv.WrappedID] = true
			visitor.Visit(ctx, &sv)
		}

		vertices, _ := d.getChildren(sv.WrappedID)
		for _, id := range vertexIDs(vertices) {
			v := vertices[id]
			sv := storableVertex{WrappedID: id, Value: v}
			queue.Enqueue(sv)
		}
	}
}

func vertexIDs(vertices map[string]interface{}) []string {
	ids := make([]string, 0, len(vertices))
	for id := range vertices {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func reversedVertexIDs(vertices map[string]interface{}) []string {
	ids := vertexIDs(vertices)
	i, j := 0, len(ids)-1
	for i < j {
		ids[i], ids[j] = ids[j], ids[i]
		i++
		j--
	}
	return ids
}

type visitationVertex struct {
	sv       storableVertex
	visiting bool
	visited  bool

	done chan struct{}
}

func newVisitationVertex(sv storableVertex) *visitationVertex {
	return &visitationVertex{sv: sv, done: make(chan struct{})}
}

type visitation struct {
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.RWMutex
	wg     sync.WaitGroup

	d        *DAG
	vertices map[string]*visitationVertex
}

func newVisitation(d *DAG) *visitation {
	return &visitation{vertices: make(map[string]*visitationVertex), d: d}
}

func (v *visitation) Visit(id string, visitor Visitor) {
	go func() {
		defer v.wg.Done()

		ctx, cancel := context.WithCancel(v.ctx)
		defer cancel()

		v.mu.Lock()
		vv, ok := v.vertices[id]
		if !ok {
			v.mu.Unlock()
			return
		}

		vv.visiting = true
		v.mu.Unlock()

		visitor.Visit(ctx, &vv.sv)

		v.mu.Lock()
		vv.visited = true

		// find any downstream vertices that are ready to be visited
		children, err := v.d.getChildren(id)
		if err == nil {
			for childID, _ := range children {
				childVertex, ok := v.vertices[childID]
				if !ok {
					continue
				}

				if childVertex.visited || childVertex.visiting {
					continue
				}

				visit := true
				// ensure that all of the child's parents have been visited
				parents, err := v.d.GetParents(childID)
				if err == nil {
					for parentID, _ := range parents {
						parentVertex, ok := v.vertices[parentID]
						if !ok {
							continue
						}

						if !parentVertex.visited {
							visit = false
							continue
						}
					}
				}

				if !visit {
					continue
				}
				v.wg.Add(1)
				v.Visit(childID, visitor)
			}
		}

		v.mu.Unlock()
	}()
}

func (v *visitation) OrderedWalk(ctx context.Context, visitor Visitor) {
	v.d.muDAG.RLock()
	defer v.d.muDAG.RUnlock()

	v.ctx, v.cancel = context.WithCancel(ctx)

	allVertices := v.d.GetVertices()
	for _, id := range vertexIDs(allVertices) {
		vertex := allVertices[id]
		sv := storableVertex{WrappedID: id, Value: vertex}

		v.mu.Lock()
		v.vertices[id] = newVisitationVertex(sv)
		v.mu.Unlock()
	}

	vertices := v.d.getRoots()
	for _, id := range vertexIDs(vertices) {
		v.mu.Lock()
		v.wg.Add(1)
		v.mu.Unlock()

		v.Visit(id, visitor)
	}

	v.wg.Wait()
}

// OrderedWalk implements the Topological Sort algorithm to traverse the entire DAG.
// This means that for any edge a -> b, node a will be visited before node b.
func (d *DAG) OrderedWalk(ctx context.Context, visitor Visitor) {
	d.muDAG.RLock()
	defer d.muDAG.RUnlock()

	visitation := newVisitation(d)
	visitation.OrderedWalk(ctx, visitor)

}
