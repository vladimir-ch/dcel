package dcel

import "github.com/gonum/graph"

var (
	_ Node     = ((*BaseNode)(nil))
	_ Halfedge = ((*BaseHalfedge)(nil))
	_ Edge     = ((*BaseEdge)(nil))
	_ Face     = ((*BaseFace)(nil))
)

type BaseNode struct {
	id int
	h  Halfedge
}

func NewBaseNode(id int) *BaseNode {
	return &BaseNode{id: id}
}

func (n *BaseNode) ID() int                { return n.id }
func (n *BaseNode) Halfedge() Halfedge     { return n.h }
func (n *BaseNode) SetHalfedge(h Halfedge) { n.h = h }

type BaseHalfedge struct {
	from       Node
	twin       Halfedge
	next, prev Halfedge
	edge       Edge
	face       Face
}

func NewBaseHalfedge() *BaseHalfedge {
	return &BaseHalfedge{}
}

func (h *BaseHalfedge) From() Node          { return h.from }
func (h *BaseHalfedge) SetFrom(u Node)      { h.from = u }
func (h *BaseHalfedge) Twin() Halfedge      { return h.twin }
func (h *BaseHalfedge) SetTwin(he Halfedge) { h.twin = he }
func (h *BaseHalfedge) Next() Halfedge      { return h.next }
func (h *BaseHalfedge) SetNext(he Halfedge) { h.next = he }
func (h *BaseHalfedge) Prev() Halfedge      { return h.prev }
func (h *BaseHalfedge) SetPrev(he Halfedge) { h.prev = he }
func (h *BaseHalfedge) Edge() Edge          { return h.edge }
func (h *BaseHalfedge) SetEdge(e Edge)      { h.edge = e }
func (h *BaseHalfedge) Face() Face          { return h.face }
func (h *BaseHalfedge) SetFace(f Face)      { h.face = f }

type BaseEdge struct {
	id     int
	h1, h2 Halfedge
}

func NewBaseEdge(id int) *BaseEdge {
	return &BaseEdge{id: id}
}

func (e *BaseEdge) ID() int                         { return e.id }
func (e *BaseEdge) From() graph.Node                { return e.h1.From() }
func (e *BaseEdge) To() graph.Node                  { return e.h2.From() }
func (e *BaseEdge) Weight() float64                 { return 1 }
func (e *BaseEdge) Halfedges() (Halfedge, Halfedge) { return e.h1, e.h2 }
func (e *BaseEdge) SetHalfedges(h1, h2 Halfedge)    { e.h1, e.h2 = h1, h2 }

type BaseFace struct {
	id int
	h  Halfedge
}

func NewBaseFace(id int) *BaseFace {
	return &BaseFace{id: id}
}

func (f *BaseFace) ID() int                { return f.id }
func (f *BaseFace) Halfedge() Halfedge     { return f.h }
func (f *BaseFace) SetHalfedge(h Halfedge) { f.h = h }

// Base implements Items interface for allocating base elements of DCEL data
// structure.
type Base struct{}

func (Base) NewNode(id int) Node   { return NewBaseNode(id) }
func (Base) NewHalfedge() Halfedge { return NewBaseHalfedge() }
func (Base) NewEdge(id int) Edge   { return NewBaseEdge(id) }
func (Base) NewFace(id int) Face   { return NewBaseFace(id) }
