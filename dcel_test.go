package dcel

import "testing"

func TestTriangle(t *testing.T) {
	g := NewGraph()

	nodes := []Node{NewBaseNode(0), NewBaseNode(1), NewBaseNode(2)}
	for _, n := range nodes {
		g.AddNode(n)
	}

	var edges []Edge
	for i := 0; i < 3; i++ {
		e := MakeEdge(
			NewBaseEdge(i),
			NewBaseHalfedge(nodes[i]),
			NewBaseHalfedge(nodes[(i+1)%3]),
		)
		if err := g.SetEdge(e); err != nil {
			t.Errorf("dcel: failed to set edge %d", i)
		}
		edges = append(edges, e)
	}

	h0, _ := edges[0].Halfedges()
	h1, _ := edges[1].Halfedges()
	h2, _ := edges[2].Halfedges()
	f := NewBaseFace(0)
	if err := g.SetFace(f, h0, h1, h2); err != nil {
		t.Error(err)
	}

	if len(g.Nodes()) != 3 {
		t.Errorf("dcel: triangle returns wrong number of nodes")
	}
}

func TestTwoTriangles(t *testing.T) {
	g := NewGraph()

	for i := 0; i < 4; i++ {
		g.AddNode(NewBaseNode(i))
	}

	for i := 0; i < 4; i++ {
		e := MakeEdge(
			NewBaseEdge(i),
			NewBaseHalfedge(g.Node(i)),
			NewBaseHalfedge(g.Node((i+1)%4)),
		)
		if err := g.SetEdge(e); err != nil {
			t.Errorf("dcel: failed to set edge %d", i)
		}
	}
	e := MakeEdge(
		NewBaseEdge(4),
		NewBaseHalfedge(g.Node(1)),
		NewBaseHalfedge(g.Node(3)),
	)
	if err := g.SetEdge(e); err != nil {
		t.Errorf("dcel: failed to set edge %d", 4)
	}

	h0, _ := g.Edge(0).Halfedges()
	h1, _ := g.Edge(4).Halfedges()
	h2, _ := g.Edge(3).Halfedges()
	f1 := NewBaseFace(0)
	if err := g.SetFace(f1, h0, h1, h2); err != nil {
		t.Error(err)
	}

	h3, _ := g.Edge(1).Halfedges()
	h4, _ := g.Edge(2).Halfedges()
	_, h5 := g.Edge(4).Halfedges()
	f2 := NewBaseFace(1)
	if err := g.SetFace(f2, h3, h4, h5); err != nil {
		t.Error(err)
	}

	if len(g.Nodes()) != 4 {
		t.Errorf("dcel: two triangle graph returns wrong number of nodes")
	}
}

func TestQuad(t *testing.T) {
	g := NewGraph()

	for i := 0; i < 4; i++ {
		g.AddNode(NewBaseNode(i))
	}

	for i := 0; i < 4; i++ {
		e := MakeEdge(
			NewBaseEdge(i),
			NewBaseHalfedge(g.Node(i)),
			NewBaseHalfedge(g.Node((i+1)%4)),
		)
		if err := g.SetEdge(e); err != nil {
			t.Errorf("dcel: failed to set edge %d", i)
		}
	}

	h0, _ := g.Edge(0).Halfedges()
	h1, _ := g.Edge(1).Halfedges()
	h2, _ := g.Edge(2).Halfedges()
	h3, _ := g.Edge(3).Halfedges()
	f := NewBaseFace(0)
	if err := g.SetFace(f, h0, h1, h2, h3); err != nil {
		t.Error(err)
	}

	if len(g.Nodes()) != 4 {
		t.Errorf("dcel: two triangle graph returns wrong number of nodes")
	}
}
