package dcel

import "fmt"

// var (
// 	dcelGraph *Graph
// 	_         graph.Graph = dcelGraph
// )

type Graph struct {
	nodes map[int]Node
	edges map[int]Edge
	faces map[int]Face

	nextNodeID int
	nextEdgeID int
	nextFaceID int

	freeNodes map[int]struct{}
	freeEdges map[int]struct{}
	freeFaces map[int]struct{}
}

func NewGraph() *Graph {
	return &Graph{
		nodes: make(map[int]Node),
		edges: make(map[int]Edge),
		faces: make(map[int]Face),

		freeNodes: make(map[int]struct{}),
		freeEdges: make(map[int]struct{}),
		freeFaces: make(map[int]struct{}),
	}
}

func (g *Graph) Node(id int) Node { return g.nodes[id] }
func (g *Graph) Edge(id int) Edge { return g.edges[id] }
func (g *Graph) Face(id int) Face { return g.faces[id] }

func (g *Graph) has(id int) bool {
	_, exists := g.nodes[id]
	return exists
}
func (g *Graph) Has(n Node) bool { return g.has(n.ID()) }

func (g *Graph) Nodes() []Node {
	var nodes []Node
	for _, n := range g.nodes {
		nodes = append(nodes, n)
	}
	return nodes
}

func (g *Graph) From(n Node) []Node {
	start := n.Halfedge()
	if start == nil {
		// Node n is isolated.
		return nil
	}
	var from []Node
	for iter := start; ; {
		from = append(from, iter.Twin().From())
		iter = iter.Twin().Next()
		if iter == start {
			break
		}
	}
	return from
}

func (g *Graph) Halfedge(x, y Node) Halfedge {
	if !g.Has(x) || !g.Has(y) {
		// At least one of the nodes does not belong to the graph.
		return nil
	}
	if x.Halfedge() == nil || y.Halfedge() == nil {
		// At least one of the nodes is isolated.
		return nil
	}
	start := x.Halfedge()
	for iter := start; ; {
		if iter.Twin().From().ID() == y.ID() {
			return iter
		}
		iter = iter.Twin().Next()
		if iter == start {
			break
		}
	}
	return nil
}

func (g *Graph) HalfedgesFrom(x Node) []Halfedge {
	if !g.Has(x) {
		return nil
	}
	if x.Halfedge() == nil {
		return nil
	}
	var hes []Halfedge
	for iter := x.Halfedge(); ; {
		hes = append(hes, iter)
		iter = iter.Twin().Next()
		if iter == x.Halfedge() {
			break
		}
	}
	return hes
}

func (g *Graph) hasEdge(id int) bool {
	_, exists := g.edges[id]
	return exists
}

func (g *Graph) HasEdge(e Edge) bool {
	return g.hasEdge(e.ID())
}

// func (g *Graph) HasEdge(x, y Node) bool {
// 	return g.Halfedge(x, y) != nil
// }
//
// func (g *Graph) Edge(u, v Node) Edge {
// 	he := g.Halfedge(u, v)
// 	if he == nil {
// 		return nil
// 	}
// 	return he.Edge()
// }

func (g *Graph) EdgeBetween(x, y Node) Edge {
	he := g.Halfedge(x, y)
	if he == nil {
		return nil
	}
	return he.Edge()
}

func (g *Graph) NewNodeID() int {
	if g.nextNodeID != maxInt {
		id := g.nextNodeID
		g.nextNodeID++
		return id
	}
	// All node IDs have already been used. See if at least one has been
	// released.
	for id := range g.freeNodes {
		return id
	}
	if len(g.nodes) == maxInt {
		panic("dcel: graph too large")
	}
	// Resort to checking all positive integers to see if there is at least one
	// unused.
	for id := 0; id < maxInt; id++ {
		if _, exists := g.nodes[id]; !exists {
			return id
		}
	}
	panic("dcel: no free node ID")
}

func (g *Graph) NewEdgeID() int {
	if g.nextEdgeID != maxInt {
		id := g.nextEdgeID
		g.nextEdgeID++
		return id
	}
	// All edge IDs have already been used. See if at least one has been
	// released.
	for id := range g.freeEdges {
		return id
	}
	if len(g.edges) == maxInt {
		panic("dcel: graph too large")
	}
	// Resort to checking all positive integers to see if there is at least one
	// unused.
	for id := 0; id < maxInt; id++ {
		if _, exists := g.edges[id]; !exists {
			return id
		}
	}
	panic("dcel: no free edge ID")
}

func (g *Graph) NewFaceID() int {
	if g.nextFaceID != maxInt {
		id := g.nextFaceID
		g.nextFaceID++
		return id
	}
	// All face IDs have already been used. See if at least one has been
	// released.
	for id := range g.freeFaces {
		return id
	}
	if len(g.faces) == maxInt {
		panic("dcel: graph too large")
	}
	// Resort to checking all positive integers to see if there is at least one
	// unused.
	for id := 0; id < maxInt; id++ {
		if _, exists := g.faces[id]; !exists {
			return id
		}
	}
	panic("dcel: no free face ID")
}

func (g *Graph) AddNode(n Node) {
	if g.Has(n) {
		panic(fmt.Sprintf("dcel: node ID collision: %d", n.ID()))
	}

	g.nodes[n.ID()] = n
	delete(g.freeNodes, n.ID())
	g.nextNodeID = max(g.nextNodeID, n.ID())
}

func (g *Graph) RemoveNode(n Node) {
	if !g.Has(n) {
		return
	}

	for _, h := range g.HalfedgesFrom(n) {
		g.RemoveEdge(h.Edge())
	}
	n.SetHalfedge(nil)

	delete(g.nodes, n.ID())
	if g.nextNodeID != 0 && n.ID() == g.nextNodeID {
		g.nextNodeID--
	}
	g.freeNodes[n.ID()] = struct{}{}
}

// Edge e must be initialized so that its halfedges are paired as given by
// Twin() method and connected to its end nodes.
// Unlike graph.SetEdge, SetEdge returns error.
func (g *Graph) SetEdge(e Edge) error {
	if _, exists := g.edges[e.ID()]; exists {
		return fmt.Errorf("dcel: edge ID collision: %d", e.ID())
	}
	from := e.From()
	to := e.To()
	if from.ID() == to.ID() {
		return fmt.Errorf("dcel: trying to set a loop edge at node %d", from.ID())
	}
	if !g.Has(from) {
		return fmt.Errorf("dcel: node %d not present", from.ID())
	}
	if !g.Has(to) {
		return fmt.Errorf("dcel: node %d not present", to.ID())
	}

	h1, h2 := e.Halfedges()
	if err := attach(h1); err != nil {
		return err
	}
	if err := attach(h2); err != nil {
		detach(h1)
		return err
	}

	g.edges[e.ID()] = e
	delete(g.freeEdges, e.ID())
	g.nextEdgeID = max(g.nextEdgeID, e.ID())
	return nil
}

func attach(h Halfedge) error {
	// h is already connected to its From node, but the node does not know
	// about it, so change that.

	u := h.From()
	if u.Halfedge() == nil {
		// From node is isolated.

		u.SetHalfedge(h)
		h.SetPrev(h.Twin())
		h.Twin().SetNext(h)
		return nil
	}

	// From node is not isolated, so we must update its neighboring halfedges.

	// First find a free (i.e., without an adjacent face) halfedge from u.
	out := u.Halfedge()
	for {
		if out.Face() == nil {
			break
		}
		out = out.Twin().Next()
		if out == u.Halfedge() {
			return fmt.Errorf("dcel: no free halfedge from node %d", u.ID())
		}
	}

	// Adjust the connections.
	in := out.Prev()
	in.SetNext(h)
	h.SetPrev(in)
	h.Twin().SetNext(out)
	out.SetPrev(h.Twin())
	return nil
}

// RemoveEdge removes e and its adjacent faces from g.
func (g *Graph) RemoveEdge(e Edge) {
	if _, exists := g.edges[e.ID()]; !exists {
		return
	}

	h1, h2 := e.Halfedges()
	if h1.Face() != nil {
		g.RemoveFace(h1.Face())
	}
	if h2.Face() != nil {
		g.RemoveFace(h2.Face())
	}

	detach(h1)
	detach(h2)

	delete(g.edges, e.ID())
	if g.nextEdgeID != 0 && e.ID() == g.nextEdgeID {
		g.nextEdgeID--
	}
	g.freeEdges[e.ID()] = struct{}{}
}

func detach(h Halfedge) {
	if h.Face() != nil {
		panic("dcel: face not removed before detaching halfedge")
	}

	out := h.Twin().Next()
	in := h.Prev()
	from := h.From()
	if from.Halfedge() == h {
		// h is the halfedge referenced by its from node.
		if out == h {
			// It is also the only halfedge adjacent to the from node, so it
			// will become isolated.
			from.SetHalfedge(nil)
		} else {
			if out.Face() != nil {
				panic("dcel: outgoing halfedge is not free")
			}
			from.SetHalfedge(out)
		}
	}
	out.SetPrev(in)
	in.SetNext(out)

	h.SetFrom(nil)
	h.SetPrev(h.Twin())
	h.Twin().SetNext(h)
}

// SetFace reconnects given halfedges so that they form a loop and sets f's
// halfedge to one of them.
func (g *Graph) SetFace(f Face, halfedges ...Halfedge) error {
	// Alternatively, this could take a slice of Nodes, but it seems more
	// consistent to create graph entities from entities of one dimension less,
	// i.e., edges from nodes, faces from edges. On the other hand, setting it
	// from nodes and finding the halfedges automatically is more convenient
	// (but also slower and the edges must be already set).

	// There must be at least three halfedges.
	if len(halfedges) < 3 {
		panic(fmt.Sprintf("dcel: cannot set a face %d from only %d halfedges", f.ID(), len(halfedges)))
	}

	// Check that the halfedges form a chain as given by their From nodes.
	for i, h := range halfedges {
		j := (i + 1) % len(halfedges)
		if h.Twin().From() != halfedges[j].From() {
			return fmt.Errorf("dcel: cannot set face %d, halfedges do not form a chain", f.ID())
		}
	}
	// Check that they are all free.
	for _, h := range halfedges {
		if h.Face() != nil {
			return fmt.Errorf("dcel: cannot set face %d, halfedge from %d to %d already has a face",
				f.ID(), h.From().ID(), h.Twin().From().ID())
		}
	}

	// Add any missing edges. If adding or the reconnection in the next step
	// fail, we have already modified the graph.
	for _, h := range halfedges {
		if !g.HasEdge(h.Edge()) {
			if err := g.SetEdge(h.Edge()); err != nil {
				return err
			}
		}
	}

	// The main action: reconnect the halfedges so the Next and Prev point to
	// consecutive neighbors.
	for i, h1 := range halfedges {
		h2 := halfedges[(i+1)%len(halfedges)]
		if err := reconnect(h1, h2); err != nil {
			return err
		}
	}

	g.faces[f.ID()] = f
	delete(g.freeFaces, f.ID())
	g.nextFaceID = max(g.nextFaceID, f.ID())
	return nil
}

func reconnect(in, out Halfedge) error {
	// We know that in.Twin().From() == out.From().

	if in.Next() == out {
		if out.Prev() != in {
			// This would be our bug.
			panic(fmt.Sprintf("dcel: halfedges around node %d are inconsistently connected",
				out.From().ID()))
		}
		// in and out are already adjacent.
		return nil
	}

	// Find a free incoming halfedge adjacent to the common node between
	// out.Twin() and in.
	var b Halfedge
	for iter := out.Twin(); ; {
		if iter.Face() == nil {
			b = iter
			break
		}
		iter = iter.Next().Twin()
		if iter == in {
			break
		}
	}
	if b == nil {
		return fmt.Errorf("dcel: halfedge reconnection failed around node %d", out.From())
	}

	// Reconnect the halfedges.
	inNext := in.Next()
	outPrev := out.Prev()
	bNext := b.Next()

	in.SetNext(out)
	out.SetPrev(in)

	b.SetNext(inNext)
	inNext.SetPrev(b)

	outPrev.SetNext(bNext)
	bNext.SetPrev(outPrev)

	return nil
}

// RemoveFace disconnects f from g and sets its Halfedge to nil.
func (g *Graph) RemoveFace(f Face) {
	if _, exists := g.faces[f.ID()]; !exists {
		return
	}

	start := f.Halfedge()
	if start != nil {
		// Set the face of all adjacent halfedges to nil.
		for iter := start; ; {
			iter.SetFace(nil)
			iter = iter.Next()
			if iter == start {
				break
			}
		}
	}
	f.SetHalfedge(nil)

	delete(g.faces, f.ID())
	if g.nextFaceID != 0 && f.ID() == g.nextFaceID {
		g.nextFaceID--
	}
	g.freeFaces[f.ID()] = struct{}{}
}

type Node interface {
	ID() int
	Halfedge() Halfedge
	SetHalfedge(Halfedge)
}

type Halfedge interface {
	From() Node
	SetFrom(Node)

	Twin() Halfedge
	SetTwin(Halfedge)

	Next() Halfedge
	SetNext(Halfedge)

	// Holding a reference to the previous halfedge could be optional.
	Prev() Halfedge
	SetPrev(Halfedge)

	Edge() Edge
	SetEdge(Edge)

	Face() Face
	SetFace(Face)
}

type Edge interface {
	ID() int
	From() Node
	To() Node
	// Weight() float64

	Halfedges() (Halfedge, Halfedge)
	SetHalfedges(Halfedge, Halfedge)
}

func MakeEdge(e Edge, h1, h2 Halfedge) Edge {
	e.SetHalfedges(h1, h2)
	h1.SetTwin(h2)
	h2.SetTwin(h1)
	h1.SetEdge(e)
	h2.SetEdge(e)
	h1.SetNext(h2)
	h2.SetNext(h1)
	h1.SetPrev(h2)
	h2.SetPrev(h1)
	return e
}

type Face interface {
	ID() int
	Halfedge() Halfedge
	SetHalfedge(Halfedge)
}

func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}

const maxInt int = int(^uint(0) >> 1)
