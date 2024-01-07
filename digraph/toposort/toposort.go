package toposort

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

// storableVertex implements the Vertexer interface.
// It is implemented as a storable structure.
// And it uses short json tag to reduce the number of bytes after serialization.
type storableVertex[T any] struct {
	WrappedID string `json:"i"`
	Value     T      `json:"v"`
}

func (v *storableVertex[T]) Vertex() (id string, value T) {
	return v.WrappedID, v.Value
}

func (v *storableVertex[T]) ID() string {
	return v.WrappedID
}

// Vertexer is the interface that wraps the basic Vertex method.
// Vertex returns an id that identifies this vertex and the value of this vertex.
//
// The reason for defining this new structure is that the vertex id may be
// automatically generated when the caller adds a vertex. At this time, the
// vertex structure added by the user does not contain id information.
type Vertexer[T any] interface {
	Vertex() (id string, value T)
}

// Visitor is the interface that wraps the basic Visit method.
// It can use the Visitor and XXXWalk functions together to traverse the entire DAG.
// And access per-vertex information when traversing.
type Visitor[T any] interface {
	Visit(context.Context, *storableVertex[T])
}

type visitationVertex[T any] struct {
	sv       storableVertex[T]
	visiting bool
	visited  bool
}

func newVisitationVertex[T any](sv storableVertex[T]) *visitationVertex[T] {
	return &visitationVertex[T]{sv: sv}
}

type visitation[T any] struct {
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.RWMutex
	wg     sync.WaitGroup

	graph    *Graph[T]
	vertices map[string]*visitationVertex[T]
}

func newVisitation[T any](g *Graph[T]) *visitation[T] {
	return &visitation[T]{vertices: make(map[string]*visitationVertex[T]), graph: g}
}

type VertexerFunc[T any] func(context.Context, string, T)

func (v *visitation[T]) Visit(id string, vfunc VertexerFunc[T]) {
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

		vfunc(ctx, id, vv.sv.Value)

		v.mu.Lock()
		vv.visited = true

		// find any downstream vertices that are ready to be visited
		children, err := v.graph.dag.GetChildren(id)
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
				parents, err := v.graph.dag.GetParents(childID)
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
				v.Visit(childID, vfunc)
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

func (g *Graph[T]) topoSort(ctx context.Context, vfunc VertexerFunc[T]) {
	visitation := newVisitation[T](g)

	visitation.ctx, visitation.cancel = context.WithCancel(ctx)

	allVertices := g.dag.GetVertices()
	for _, id := range vertexIDs(allVertices) {
		fmt.Println("vertex", id)
		value, ok := g.Vertex(id)
		fmt.Println("vertex", id, value, ok)
		if !ok {
			continue
		}
		sv := storableVertex[T]{WrappedID: id, Value: value.Value}

		visitation.mu.Lock()
		visitation.vertices[id] = newVisitationVertex(sv)
		visitation.mu.Unlock()
	}

	vertices := g.dag.GetRoots()
	for _, id := range vertexIDs(vertices) {
		visitation.mu.Lock()
		visitation.wg.Add(1)
		visitation.mu.Unlock()

		visitation.Visit(id, vfunc)
	}

	visitation.wg.Wait()
}
