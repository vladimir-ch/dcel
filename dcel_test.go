package dcel

import "testing"

type NodeID int

func (n NodeID) ID() int { return int(n) }

func TestTriangle(t *testing.T) {
	g := New(nil)

	err := g.AddFace(0, NodeID(0), NodeID(1), NodeID(2))
	if err != nil {
		t.Error(err)
	}

	hedges := g.HalfedgesAround(g.Face(0))
	if len(hedges) != 3 {
		t.Error("dcel: wrong number of halfedges")
	}

	for i := 0; i < 3; i++ {
		h := g.Halfedge(NodeID(i), NodeID((i+1)%3))
		if h == nil {
			t.Error("dcel: halfedge does not exist")
		}
	}

	if len(g.Nodes()) != 3 {
		t.Error("dcel: graph with one triangle has wrong number of nodes")
	}
}

func TestTwoTriangles(t *testing.T) {
	g := New(nil)

	err := g.AddFace(0, NodeID(0), NodeID(1), NodeID(2))
	if err != nil {
		t.Fatal(err)
	}
	err = g.AddFace(1, NodeID(2), NodeID(1), NodeID(3))
	if err != nil {
		t.Fatal(err)
	}

	hedges := g.HalfedgesAround(g.Face(0))
	if len(hedges) != 3 {
		t.Error("dcel: wrong number of halfedges")
	}
	hedges = g.HalfedgesAround(g.Face(1))
	if len(hedges) != 3 {
		t.Error("dcel: wrong number of halfedges")
	}

	for i := 0; i < 3; i++ {
		h := g.Halfedge(NodeID(i), NodeID((i+1)%3))
		if h == nil {
			t.Error("dcel: halfedge does not exist")
		}
	}

	if len(g.Nodes()) != 4 {
		t.Error("dcel: graph with two triangles has wrong number of nodes")
	}
}

func TestTriangleSquare(t *testing.T) {
	g := New(nil)

	err := g.AddFace(0, NodeID(0), NodeID(1), NodeID(2))
	if err != nil {
		t.Fatal(err)
	}
	err = g.AddFace(1, NodeID(2), NodeID(1), NodeID(3), NodeID(4))
	if err != nil {
		t.Fatal(err)
	}

	hedges := g.HalfedgesAround(g.Face(0))
	if len(hedges) != 3 {
		t.Error("dcel: wrong number of halfedges")
	}
	hedges = g.HalfedgesAround(g.Face(1))
	if len(hedges) != 4 {
		t.Error("dcel: wrong number of halfedges")
	}

	for i := 0; i < 3; i++ {
		h := g.Halfedge(NodeID(i), NodeID((i+1)%3))
		if h == nil {
			t.Error("dcel: halfedge does not exist")
		}
	}

	if len(g.Nodes()) != 5 {
		t.Error("dcel: graph with triangle and square has wrong number of nodes")
	}
}
