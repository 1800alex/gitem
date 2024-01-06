package main

import (
	"context"
	"fmt"
	"time"

	"1800alex/gitem/digraph/toposort"
	"1800alex/gitem/digraph/toposort/dag"
)

type Project struct {
	Name      string
	DependsOn []string

	vertexID string
}

type Projects []Project

func (p Projects) Len() int {
	return len(p)
}

func (p Projects) ByID(id string) *Project {
	for _, project := range p {
		if project.vertexID == id {
			return &project
		}
	}
	return nil
}

func (p Projects) ByName(name string) *Project {
	for _, project := range p {
		if project.Name == name {
			return &project
		}
	}
	return nil
}

func (p Projects) Visit(ctx context.Context, visitor toposort.Vertexer) {
	id, _ := visitor.Vertex()
	proj := p.ByID(id)
	if proj == nil {
		return
	}
	fmt.Println("visiting", proj.Name)
	time.Sleep(2000 * time.Millisecond)
	fmt.Println("done", proj.Name)
}

func main() {

	projects := Projects{
		{Name: "alfa"},
		{Name: "bravo", DependsOn: []string{"alfa"}},
		{Name: "charlie", DependsOn: []string{"alfa"}},
		{Name: "delta", DependsOn: []string{"charlie"}},
		{Name: "echo", DependsOn: []string{"delta"}},
		{Name: "foxtrot", DependsOn: []string{"alfa"}},
		{Name: "golf", DependsOn: []string{"echo"}},
		{Name: "hotel"},
	}

	// initialize a new graph
	d := dag.NewDAG()

	vertices := []string{}

	for i, project := range projects {
		v, _ := d.AddVertex(project.Name)
		projects[i].vertexID = v
		vertices = append(vertices, v)
	}

	// TODO add edges
	for _, project := range projects {
		for _, dependency := range project.DependsOn {
			dep := projects.ByName(dependency)
			if dep == nil {
				continue
			}

			_ = d.AddEdge(dep.vertexID, project.vertexID)
		}
	}

	// add the above vertices and connect them with two edges
	_ = d.AddEdge(vertices[0], vertices[1])
	_ = d.AddEdge(vertices[0], vertices[2])

	toposort.TopologicalSort(context.Background(), d, projects)

	// describe the graph
	fmt.Print(d.String())
}
