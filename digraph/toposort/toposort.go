package toposort

import (
	"context"
	"sort"
	"sync"

	"1800alex/gitem/digraph/toposort/dag"
)

var (
	_ Vertexer = (*storableVertex)(nil)
)

// storableVertex implements the Vertexer interface.
// It is implemented as a storable structure.
// And it uses short json tag to reduce the number of bytes after serialization.
type storableVertex struct {
	WrappedID string      `json:"i"`
	Value     interface{} `json:"v"`
}

func (v storableVertex) Vertex() (id string, value interface{}) {
	return v.WrappedID, v.Value
}

func (v storableVertex) ID() string {
	return v.WrappedID
}

// Vertexer is the interface that wraps the basic Vertex method.
// Vertex returns an id that identifies this vertex and the value of this vertex.
//
// The reason for defining this new structure is that the vertex id may be
// automatically generated when the caller adds a vertex. At this time, the
// vertex structure added by the user does not contain id information.
type Vertexer interface {
	Vertex() (id string, value interface{})
}

// Visitor is the interface that wraps the basic Visit method.
// It can use the Visitor and XXXWalk functions together to traverse the entire DAG.
// And access per-vertex information when traversing.
type Visitor interface {
	Visit(context.Context, Vertexer)
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

	d        *dag.DAG
	vertices map[string]*visitationVertex
}

func newVisitation(d *dag.DAG) *visitation {
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
		children, err := v.d.GetChildren(id)
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

func vertexIDs(vertices map[string]interface{}) []string {
	ids := make([]string, 0, len(vertices))
	for id := range vertices {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func (v *visitation) topoSort(ctx context.Context, visitor Visitor) {
	// v.d.muDAG.RLock()
	// defer v.d.muDAG.RUnlock()

	v.ctx, v.cancel = context.WithCancel(ctx)

	allVertices := v.d.GetVertices()
	for _, id := range vertexIDs(allVertices) {
		vertex := allVertices[id]
		sv := storableVertex{WrappedID: id, Value: vertex}

		v.mu.Lock()
		v.vertices[id] = newVisitationVertex(sv)
		v.mu.Unlock()
	}

	vertices := v.d.GetRoots()
	for _, id := range vertexIDs(vertices) {
		v.mu.Lock()
		v.wg.Add(1)
		v.mu.Unlock()

		v.Visit(id, visitor)
	}

	v.wg.Wait()
}

// TopologicalSort implements the Topological Sort algorithm to traverse the entire DAG.
// This means that for any edge a -> b, node a will be visited before node b.
func TopologicalSort(ctx context.Context, d *dag.DAG, visitor Visitor) {
	// d.muDAG.RLock()
	// defer d.muDAG.RUnlock()

	visitation := newVisitation(d)
	visitation.topoSort(ctx, visitor)
}
