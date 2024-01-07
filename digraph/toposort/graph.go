package toposort

import (
	"context"
	"fmt"
	"sync"

	"1800alex/gitem/digraph/toposort/dag"
)

type Vertex[T any] struct {
	ID    string
	Value T
}

type Graph[T any] struct {
	mu  sync.Mutex
	dag *dag.DAG

	vertices map[string]Vertex[T]
}

func New[T any]() *Graph[T] {
	return &Graph[T]{dag: dag.NewDAG(), vertices: make(map[string]Vertex[T])}
}

func (g *Graph[T]) Vertices() map[string]Vertex[T] {
	g.mu.Lock()
	defer g.mu.Unlock()

	return g.vertices
}

func (g *Graph[T]) Vertex(id string) (Vertex[T], bool) {
	g.mu.Lock()
	defer g.mu.Unlock()

	dagVertex, err := g.dag.GetVertex(id)
	if err != nil {
		return Vertex[T]{}, false
	}

	v, ok := g.vertices[dagVertex.(string)]
	return v, ok
}

func (g *Graph[T]) AddVertex(id string, v T) (string, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	vID, err := g.dag.AddVertex(id)
	if err != nil {
		return "", err
	}

	g.vertices[id] = Vertex[T]{ID: vID, Value: v}
	return vID, nil
}

func (g *Graph[T]) AddEdge(from, to string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	return g.dag.AddEdge(from, to)
}

func (g *Graph[T]) String() string {
	g.mu.Lock()
	defer g.mu.Unlock()

	return g.dag.String()
}

func (g *Graph[T]) VertexerFunc(vfunc VertexerFunc[T]) VertexerFunc[T] {
	return func(ctx context.Context, id string, v T) {
		fmt.Println("visiting", id)
		vertex, ok := g.Vertex(id)
		if !ok {
			return
		}

		vfunc(ctx, id, vertex.Value)
	}
}

func (g *Graph[T]) TopologicalSort(ctx context.Context, vfunc VertexerFunc[T]) {
	g.topoSort(ctx, g.VertexerFunc(vfunc))
}
