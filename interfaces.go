package dcel

import "github.com/gonum/graph"

// Node is a graph node in the DCEL data structure.
type Node interface {
	graph.Node

	// Halfedge returns the outgoing halfedge from the node. When the node is
	// isolated, the returned halfedge is nil. When the node is at a boundary,
	// the halfedge's Face is nil.
	Halfedge() Halfedge
	// SetHalfedge sets the outgoing halfedge from the node.
	SetHalfedge(Halfedge)
}

// Halfedge is a an oriented edge in the DCEL data structure.
type Halfedge interface {
	// From returns the origin node.
	From() Node
	// SetFrom sets the origin node.
	SetFrom(Node)

	// Twin returns the twin halfedge in the same edge.
	Twin() Halfedge
	// SetTwin sets the twin halfedge in the same edge.
	SetTwin(Halfedge)

	// Next returns the next halfedge around the adjacent face.
	Next() Halfedge
	// SetNext sets the next halfedge around the adjacent face.
	SetNext(Halfedge)

	// Prev returns the previous halfedge around the adjacent face.
	Prev() Halfedge
	// SetPrev sets the previous halfedge around the adjacent face.
	SetPrev(Halfedge)

	// Edge returns the undirected edge to which the halfedge belongs.
	Edge() Edge
	// SetEdge sets the undirected edge to which the halfedge belongs.
	SetEdge(Edge)

	// Face returns the adjacent face.
	Face() Face
	// SetFace sets the adjacent face.
	SetFace(Face)
}

// Edge is an undirected edge in the DCEL data structure.
type Edge interface {
	graph.Edge

	// ID returns an edge identifier unique within the graph.
	ID() int

	// Halfedges returns the two halfedges that form the edge.
	Halfedges() (Halfedge, Halfedge)
	// SetHalfedges sets the two halfedges that form the edge.
	SetHalfedges(Halfedge, Halfedge)
}

// Face is a face in the DCEL data structure.
type Face interface {
	// ID returns a face identifier unique within the graph.
	ID() int

	// Halfedge returns an adjacent halfedge.
	Halfedge() Halfedge
	// SetHalfedge sets an adjacent halfedge.
	SetHalfedge(Halfedge)
}

// Items wraps methods for allocating graph entities that can be stored in the
// DCEL data structure.
type Items interface {
	// NewNode returns a new node with the given id.
	NewNode(int) Node
	// NewHalfedge returns a new halfedge.
	NewHalfedge() Halfedge
	// NewEdge returns a new edge with the given id.
	NewEdge(int) Edge
	// NewFace returns a new face with the given id.
	NewFace(int) Face
}
