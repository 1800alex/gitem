package main

import (
	"context"
	"fmt"
	"time"

	"1800alex/gitem/digraph/toposort"
)

type Project struct {
	Name      string
	DependsOn []string

	vertexID string
}

func (p *Project) ID() string {
	return p.Name
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

func ProjectVisit(ctx context.Context, id string, v Project) {
	// id, _ := visitor.Vertex()
	// proj := p.ByID(id)
	// if proj == nil {
	// 	return
	// }
	fmt.Println("visiting", v.Name)
	time.Sleep(2000 * time.Millisecond)
	fmt.Println("done", v.Name)
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
	graph := toposort.New[Project]()

	vertices := []string{}

	for i, project := range projects {
		fmt.Println("adding vertex", project.ID())
		v, _ := graph.AddVertex(project.Name, project)
		fmt.Println("added vertex", v)
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

			_ = graph.AddEdge(dep.vertexID, project.vertexID)
		}
	}

	// add the above vertices and connect them with two edges
	_ = graph.AddEdge(vertices[0], vertices[1])
	_ = graph.AddEdge(vertices[0], vertices[2])

	graph.TopologicalSort(context.Background(), ProjectVisit) // TODO implement Visitor with generics

	// describe the graph
	fmt.Print(graph.String())
}
