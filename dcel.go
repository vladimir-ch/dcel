package dcel

import (
	"fmt"

	"github.com/gonum/graph"
)

var (
	dcelGraph *Graph
	_         graph.Undirected = dcelGraph
)

// Graph implements the doubly-connected edge list data structure.
type Graph struct {
	items Items

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

// New returns a new Graph. If items is nil, Base will be used.
func New(items Items) *Graph {
	if items == nil {
		items = Base{}
	}
	return &Graph{
		items: items,

		nodes: make(map[int]Node),
		edges: make(map[int]Edge),
		faces: make(map[int]Face),

		freeNodes: make(map[int]struct{}),
		freeEdges: make(map[int]struct{}),
		freeFaces: make(map[int]struct{}),
	}
}

// Node returns the node with the given id or nil if it does not exist within
// the graph.
func (g *Graph) Node(id int) Node { return g.nodes[id] }

// Face returns the face with the given id or nil if it does not exist within
// the graph.
func (g *Graph) Face(id int) Face { return g.faces[id] }

// Edge returns the edge with the given id or nil if it does not exist within
// the graph.
// func (g *Graph) Edge(id int) Edge { return g.edges[id] }

// Has returns whether a node with the id given by x.ID() exists within the graph.
func (g *Graph) Has(x graph.Node) bool {
	return g.has(x.ID())
}

// has returns whether a node with the given id exists within the graph.
func (g *Graph) has(id int) bool {
	_, exists := g.nodes[id]
	return exists
}

// Nodes returns all the nodes in the graph.
func (g *Graph) Nodes() []graph.Node {
	var nodes []graph.Node
	for _, u := range g.nodes {
		nodes = append(nodes, u)
	}
	return nodes
}

// From returns all neighbors of the node x.
func (g *Graph) From(x graph.Node) []graph.Node {
	u := g.Node(x.ID())
	if u == nil {
		return nil
	}
	if u.Halfedge() == nil {
		// Node n is isolated, so there are no neighbors.
		return nil
	}
	var (
		from  []graph.Node
		start = u.Halfedge().Twin() // An incoming halfedge to u.
	)
	for iter := start; ; {
		from = append(from, iter.From())
		iter = iter.Next().Twin()
		if iter == start {
			break
		}
	}
	return from
}

// HasEdge returns whether an edge exists between nodes x and y.
func (g *Graph) HasEdge(x, y graph.Node) bool {
	return g.Halfedge(x, y) != nil
}

// Edge returns the edge between x and y or nil if the nodes are not connected.
func (g *Graph) Edge(x, y graph.Node) graph.Edge {
	return g.EdgeBetween(x, y)
}

// EdgeBetween returns the edge between x and y or nil if the nodes are not
// connected.
func (g *Graph) EdgeBetween(x, y graph.Node) graph.Edge {
	he := g.Halfedge(x, y)
	if he == nil {
		return nil
	}
	return he.Edge()
}

// Halfedge returns the halfedge from x to y, or nil if the nodes are not
// connected by an edge or at least one is isolated.
func (g *Graph) Halfedge(x, y graph.Node) Halfedge {
	u := g.Node(x.ID())
	v := g.Node(y.ID())
	if u == nil || v == nil {
		// One of the nodes does not belong to the graph.
		return nil
	}
	if u.Halfedge() == nil || v.Halfedge() == nil {
		// At least one of the nodes is isolated.
		return nil
	}
	start := u.Halfedge() // An outgoing halfedge from u.
	for iter := start; ; {
		if iter.Twin().From() == v {
			return iter
		}
		iter = iter.Twin().Next()
		if iter == start {
			break
		}
	}
	// If we get here, the nodes are not connected.
	return nil
}

// NewNodeID returns a new node id unique within the graph.
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

// AddNode adds a new, isolated node with the given id to the graph and returns it.
// AddNode panics if a node with same id already exists in the graph.
func (g *Graph) AddNode(id int) Node {
	if g.has(id) {
		panic(fmt.Sprintf("dcel: node ID collision: %d", id))
	}

	u := g.items.NewNode(id)
	u.SetHalfedge(nil)

	g.nodes[id] = u
	delete(g.freeNodes, id)
	g.nextNodeID = max(g.nextNodeID, id)

	return u
}

// RemoveNode removes the node with ID given by x.ID() from the graph as well
// as any edges attached to it.
func (g *Graph) RemoveNode(x graph.Node) {
	id := x.ID()
	if !g.has(id) {
		// Nothing to do.
		return
	}

	// Remove any attached edges.
	for _, h := range g.HalfedgesFrom(x) {
		g.RemoveEdge(h.Edge())
	}
	g.Node(id).SetHalfedge(nil) // Avoid memory leaks.

	delete(g.nodes, id)
	if g.nextNodeID != 0 && id == g.nextNodeID {
		g.nextNodeID--
	}
	g.freeNodes[id] = struct{}{}
}

// addEdge adds a new edge between nodes identified by x.ID() and y.ID() and
// returns its halfedge from x to y. If the nodes are not in the graph, they
// are added.
//
// addEdge panics if x.ID() == y.ID().
//
// It is not exported because edges cannot be added individualy, they are added
// only when adding faces.
func (g *Graph) addEdge(x, y graph.Node) (Halfedge, error) {
	if x.ID() == y.ID() {
		panic(fmt.Sprintf("dcel: trying to set a loop edge at node %d", x.ID()))
	}

	h := g.Halfedge(x, y)
	if h != nil {
		// Edge between x and y already exists, so return the halfedge.
		return h, nil
	}

	// Add any missing node.
	var u, v Node
	if !g.Has(x) {
		u = g.AddNode(x.ID())
	} else {
		u = g.Node(x.ID())
	}
	if !g.Has(y) {
		v = g.AddNode(y.ID())
	} else {
		v = g.Node(y.ID())
	}

	// Allocate a new edge and attach it to the graph.
	e := g.newEdge()
	h1, h2 := e.Halfedges()
	if err := attach(h1, u); err != nil {
		return nil, err
	}
	if err := attach(h2, v); err != nil {
		detach(h1)
		return nil, err
	}

	id := e.ID()
	g.edges[id] = e
	delete(g.freeEdges, id)
	g.nextEdgeID = max(g.nextEdgeID, id)

	return h1, nil
}

// newEdge allocates a new, properly initialized Edge not connected to any
// node.
func (g *Graph) newEdge() Edge {
	h1 := g.items.NewHalfedge()
	h2 := g.items.NewHalfedge()
	e := g.items.NewEdge(g.newEdgeID())

	h1.SetFrom(nil)
	h2.SetFrom(nil)

	h1.SetTwin(h2)
	h2.SetTwin(h1)

	h1.SetNext(h2)
	h2.SetNext(h1)

	h1.SetPrev(h2)
	h2.SetPrev(h1)

	h1.SetFace(nil)
	h2.SetFace(nil)

	h1.SetEdge(e)
	h2.SetEdge(e)
	e.SetHalfedges(h1, h2)

	return e
}

func attach(h Halfedge, u Node) error {
	h.SetFrom(u)
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

// RemoveEdge removes the edge between nodes identified by e.From and e.To and
// its adjacent faces from g.
func (g *Graph) RemoveEdge(e graph.Edge) {
	h := g.Halfedge(e.From(), e.To())
	if h == nil {
		// Nothing to do.
		return
	}

	// Remove any adjacent faces.
	if h.Face() != nil {
		g.RemoveFace(h.Face())
	}
	if h.Twin().Face() != nil {
		g.RemoveFace(h.Twin().Face())
	}

	// Detach both halfedges from their From nodes and update affected
	// halfedges.
	detach(h)
	detach(h.Twin())

	id := h.Edge().ID()
	delete(g.edges, id)
	if g.nextEdgeID != 0 && id == g.nextEdgeID {
		g.nextEdgeID--
	}
	g.freeEdges[id] = struct{}{}
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

	// Avoid memory leaks.
	// TODO(vladimir-ch): Consider having a pool of reusable Edges.
	h.SetFrom(nil)
	h.SetTwin(nil)
	h.SetNext(nil)
	h.SetPrev(nil)
	h.SetEdge(nil)
}

// HasFace returns whether a face with the given id exists in the graph.
func (g *Graph) HasFace(id int) bool {
	_, exists := g.faces[id]
	return exists
}

// AddFace adds a new face with given ID and with vertices given by nodes.
// Any missing node or edge between two consecutive nodes will be added to the
// graph first.
//
// If the nodes are not pair-wise distinct, if two consecutive nodes are
// already connected by a halfedge with an adjacent Face, or if the existing
// graph topology does not permit adding the face, an error will be returned.
//
// AddFace panics if a face with the given id already exists in the graph or if
// the length of nodes is less than 3.
func (g *Graph) AddFace(id int, nodes ...graph.Node) error {
	if g.HasFace(id) {
		panic(fmt.Sprintf("dcel: face ID collision: %d", id))
	}
	if len(nodes) < 3 {
		panic(fmt.Sprintf("dcel: cannot add face %d with only %d nodes", id, len(nodes)))
	}

	// Check that the nodes are pair-wise distinct.
	for i, x := range nodes {
		for j := i + 1; j < len(nodes); j++ {
			if x.ID() == nodes[j].ID() {
				return fmt.Errorf("dcel: cannot add face %d, duplicit node %d", id, x.ID())
			}
		}
	}

	// Collect (and add any missing) halfedges between consecutive nodes.
	// Absent nodes are added in addEdge().
	var hedges []Halfedge
	for i, x := range nodes {
		y := nodes[(i+1)%len(nodes)]
		h, err := g.addEdge(x, y)
		if err != nil {
			return err
		}
		if h.Face() != nil {
			return fmt.Errorf("dcel: cannot add face %d, halfedge from %d to %d is not free",
				id, x.ID(), y.ID())
		}
		hedges = append(hedges, h)
	}

	// Reconnect the halfedges so the Next and Prev point to consecutive
	// neighbors.
	for i, h1 := range hedges {
		h2 := hedges[(i+1)%len(hedges)]
		if err := reconnect(h1, h2); err != nil {
			return err
		}
	}

	// Allocate new face and set its halfedge.
	f := g.items.NewFace(id)
	f.SetHalfedge(hedges[0])
	// Set the face of adjacent halfedges.
	for _, h := range hedges {
		h.SetFace(f)
	}

	g.faces[id] = f
	delete(g.freeFaces, id)
	g.nextFaceID = max(g.nextFaceID, id)

	return nil
}

// reconnect adjusts the halfedges around the shared node between in and out so
// that in.Next() == out and out.Prev() == in.
// It panics if in and out do not share a common node.
func reconnect(in, out Halfedge) error {
	if in.Twin().From() != out.From() {
		panic("dcel.reconnect: halfedges are not connected")
	}

	if in.Next() == out || out.Prev() == in {
		if in.Next() != out || out.Prev() != in {
			// This would be our bug.
			panic(fmt.Sprintf("dcel.reconnect: halfedges around node %d are inconsistently connected",
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
		return fmt.Errorf("dcel: halfedge reconnection failed around node %d", out.From().ID())
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
	id := f.ID()
	if _, exists := g.faces[id]; !exists {
		// Nothing to do, a face with such id does not exist in the graph.
		return
	}

	// Disconnect the face from its adjacent halfedges.
	for _, h := range g.HalfedgesAround(f) {
		h.SetFace(nil)
	}
	f.SetHalfedge(nil)

	delete(g.faces, id)
	if g.nextFaceID != 0 && id == g.nextFaceID {
		g.nextFaceID--
	}
	g.freeFaces[id] = struct{}{}
}

func (g *Graph) newEdgeID() int {
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

// NewFaceID returns a new face id unique within the graph.
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

// HalfedgesFrom returns all halfedges whose From node is x.
func (g *Graph) HalfedgesFrom(x graph.Node) []Halfedge {
	u := g.Node(x.ID())
	if u == nil {
		// The node does not belong to the graph.
		return nil
	}
	if u.Halfedge() == nil {
		// The node is isolated.
		return nil
	}
	var (
		hedges []Halfedge
		start  = u.Halfedge() // An outgoing halfedge from u.
	)
	for iter := start; ; {
		hedges = append(hedges, iter)
		iter = iter.Twin().Next()
		if iter == start {
			break
		}
	}
	return hedges
}

// HalfedgesTo returns all halfedges whose Twin.From node is x.
func (g *Graph) HalfedgesTo(x graph.Node) []Halfedge {
	u := g.Node(x.ID())
	if u == nil {
		// The node does not belong to the graph.
		return nil
	}
	if u.Halfedge() == nil {
		// The node is isolated.
		return nil
	}
	var (
		hedges []Halfedge
		start  = u.Halfedge().Twin() // An incoming halfedge to u.
	)
	for iter := start; ; {
		hedges = append(hedges, iter)
		iter = iter.Next().Twin()
		if iter == start {
			break
		}
	}
	return hedges
}

// HalfedgesAround returns all halfedges adjacent to the given face.
func (g *Graph) HalfedgesAround(f Face) []Halfedge {
	if _, exists := g.faces[f.ID()]; !exists {
		return nil
	}
	if f.Halfedge() == nil {
		return nil
	}
	var (
		hedges []Halfedge
		start  = f.Halfedge()
	)
	for iter := start; ; {
		hedges = append(hedges, iter)
		iter = iter.Next()
		if iter == start {
			break
		}
	}
	return hedges
}

func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}

const maxInt int = int(^uint(0) >> 1)
