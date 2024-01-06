package dag

type testVertex struct {
	WID string `json:"i"`
	Val string `json:"v"`

	visited bool `json:"-"`
}

func (tv *testVertex) ID() string {
	return tv.WID
}

func (tv *testVertex) Vertex() (id string, value interface{}) {
	return tv.WID, tv.Val
}

func (tv *testVertex) WasVisited() bool {
	return tv.visited
}

func (tv *testVertex) Visited() {
	tv.visited = true
}

type testStorableDAG struct {
	StorableVertices []testVertex   `json:"vs"`
	StorableEdges    []storableEdge `json:"es"`
}

func (g testStorableDAG) Vertices() []Vertexer {
	l := make([]Vertexer, 0, len(g.StorableVertices))
	for i, _ := range g.StorableVertices {
		l = append(l, &g.StorableVertices[i])
	}
	return l
}

func (g testStorableDAG) Edges() []Edger {
	l := make([]Edger, 0, len(g.StorableEdges))
	for _, v := range g.StorableEdges {
		l = append(l, v)
	}
	return l
}
