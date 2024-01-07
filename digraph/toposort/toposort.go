package toposort

import (
	"context"
	"sort"
	"sync"
)

type visitationVertex struct {
	id       string
	visiting bool
	visited  bool
}

func newVisitationVertex(id string) *visitationVertex {
	return &visitationVertex{id: id}
}

type visitation[T any] struct {
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.RWMutex
	wg     sync.WaitGroup

	graph    *Graph[T]
	vertices map[string]*visitationVertex
}

func newVisitation[T any](g *Graph[T]) *visitation[T] {
	return &visitation[T]{vertices: make(map[string]*visitationVertex), graph: g}
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

		vertex, ok := v.graph.Vertex(id)
		if !ok {
			return
		}

		vfunc(ctx, vertex.ID, vertex.Value)

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
		visitation.mu.Lock()
		visitation.vertices[id] = newVisitationVertex(id)
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
