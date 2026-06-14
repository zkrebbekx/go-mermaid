// Package layout assigns coordinates to a domain.Graph using a layered
// (Sugiyama-style) approach: make the graph acyclic, rank nodes into layers,
// insert dummy nodes so long edges can bend, order within layers to reduce
// crossings, then assign positions with a barycenter heuristic. Ranking is
// longest-path; network-simplex can replace it behind the same interface.
package layout

import (
	"math"
	"sort"

	"github.com/Zac300/go-mermaid/internal/domain"
	"github.com/Zac300/go-mermaid/internal/svgutil"
)

// Options tunes spacing and text metrics used during layout.
type Options struct {
	NodeSep  float64 // gap between nodes within a layer
	RankSep  float64 // gap between layers
	FontSize float64 // used to estimate node sizes
}

// Result is a laid-out graph plus its overall bounds.
type Result struct {
	Graph  *domain.Graph
	Width  float64
	Height float64
}

// Compute lays out g in place and returns the result. The input graph's nodes
// and edges are mutated with positions and routed points.
func Compute(g *domain.Graph, opts Options) (*Result, error) {
	sizeNodes(g, opts)

	reversed := makeAcyclic(g)
	ranks := assignRanks(g)

	lg := buildLGraph(g, ranks)
	reduceCrossings(lg)
	totalPrimary := positionLG(lg, g.Direction, opts)

	writeBackNodes(lg, g, totalPrimary)
	// Restore original edge directions before routing so arrowheads point the
	// right way on edges that were reversed to break cycles.
	restoreReversed(g, reversed)
	routeEdges(lg, g, totalPrimary)
	separateParallel(g)
	spreadPorts(g, g.Direction == domain.TopBottom || g.Direction == domain.BottomTop)

	w, h := bounds(g)
	return &Result{Graph: g, Width: w, Height: h}, nil
}

// separateParallel offsets edges that share the same node pair perpendicular
// to their direction, so overlapping lines (e.g. A->B and B->A) and their
// labels don't sit on top of each other.
func separateParallel(g *domain.Graph) {
	key := func(a, b string) string {
		if a < b {
			return a + "\x00" + b
		}
		return b + "\x00" + a
	}
	groups := map[string][]*domain.Edge{}
	for _, e := range g.Edges {
		if e.From != e.To {
			groups[key(e.From, e.To)] = append(groups[key(e.From, e.To)], e)
		}
	}
	for _, es := range groups {
		if len(es) < 2 {
			continue
		}
		for i, e := range es {
			if len(e.Points) < 2 {
				continue
			}
			p0, pn := e.Points[0], e.Points[len(e.Points)-1]
			d := math.Hypot(pn.X-p0.X, pn.Y-p0.Y)
			if d == 0 {
				continue
			}
			px, py := -(pn.Y-p0.Y)/d, (pn.X-p0.X)/d
			off := (float64(i) - float64(len(es)-1)/2) * 34
			for j := range e.Points {
				e.Points[j].X += px * off
				e.Points[j].Y += py * off
			}
		}
	}
}

// spreadPorts fans out the attach points of edges that meet a node on the same
// face, so multiple edges (and their arrowheads) don't pile up at the box
// center. Endpoints are distributed evenly along the face and ordered by their
// far end to avoid introducing crossings; the adjacent elbow shifts with each
// endpoint so the connecting segment stays orthogonal.
func spreadPorts(g *domain.Graph, vertical bool) {
	type touch struct {
		e   *domain.Edge
		idx int // index of the endpoint touching this node
	}
	// Edges that share a node pair are already fanned out by separateParallel;
	// leaving them out here keeps that wider spacing intact.
	pairKey := func(a, b string) string {
		if a < b {
			return a + "\x00" + b
		}
		return b + "\x00" + a
	}
	pairCount := map[string]int{}
	for _, e := range g.Edges {
		if e.From != e.To {
			pairCount[pairKey(e.From, e.To)]++
		}
	}
	byNode := map[string][]touch{}
	for _, e := range g.Edges {
		if e.From == e.To || len(e.Points) < 2 || pairCount[pairKey(e.From, e.To)] > 1 {
			continue
		}
		byNode[e.From] = append(byNode[e.From], touch{e, 0})
		byNode[e.To] = append(byNode[e.To], touch{e, len(e.Points) - 1})
	}
	for id, ts := range byNode {
		n := g.NodeByID(id)
		if n == nil {
			continue
		}
		c := n.Center()
		faces := map[bool][]touch{}
		for _, t := range ts {
			p := t.e.Points[t.idx]
			pos := p.Y >= c.Y
			if !vertical {
				pos = p.X >= c.X
			}
			faces[pos] = append(faces[pos], t)
		}
		for _, grp := range faces {
			if len(grp) < 2 {
				continue
			}
			far := func(t touch) domain.Point {
				if t.idx == 0 {
					return t.e.Points[len(t.e.Points)-1]
				}
				return t.e.Points[0]
			}
			sort.SliceStable(grp, func(i, j int) bool {
				if vertical {
					return far(grp[i]).X < far(grp[j]).X
				}
				return far(grp[i]).Y < far(grp[j]).Y
			})
			for k, t := range grp {
				frac := float64(k+1) / float64(len(grp)+1)
				var port, cur float64
				if vertical {
					port = n.Pos.X + n.Size.W*frac
					cur = t.e.Points[t.idx].X
				} else {
					port = n.Pos.Y + n.Size.H*frac
					cur = t.e.Points[t.idx].Y
				}
				delta := port - cur
				adj := t.idx + 1
				if t.idx != 0 {
					adj = t.idx - 1
				}
				if vertical {
					t.e.Points[t.idx].X += delta
					t.e.Points[adj].X += delta
				} else {
					t.e.Points[t.idx].Y += delta
					t.e.Points[adj].Y += delta
				}
			}
		}
	}
}

// sizeNodes estimates a box size for each node from its label and font size.
// Nodes that already have a non-zero size (e.g. class boxes sized by their
// own renderer) are left untouched.
func sizeNodes(g *domain.Graph, opts Options) {
	const padX, padY = 20.0, 14.0
	for _, n := range g.Nodes {
		if n.Size.W != 0 || n.Size.H != 0 {
			continue
		}
		label := n.Label
		if label == "" {
			label = n.ID
		}
		lines := svgutil.SplitLines(label)
		maxW := 0.0
		for _, ln := range lines {
			if wd := svgutil.TextWidth(ln, opts.FontSize); wd > maxW {
				maxW = wd
			}
		}
		w := maxW + padX*2
		h := opts.FontSize*float64(len(lines)) + padY*2
		switch n.Shape {
		case domain.ShapeCircle:
			if w < h {
				w = h
			}
			h = w
		case domain.ShapeHexagon, domain.ShapeParallelogram, domain.ShapeParallelogramAlt,
			domain.ShapeTrapezoid, domain.ShapeTrapezoidAlt:
			w += h // slant on each side eats horizontal room
		}
		n.Size = domain.Size{W: w, H: h}
	}
}

// center maps an lnode's rank-space coordinates to a screen point, honoring
// the diagram direction (BT/RL flip the primary axis).
func (ln *lnode) center(dir domain.Direction, totalPrimary float64) domain.Point {
	switch dir {
	case domain.BottomTop:
		return domain.Point{X: ln.cross, Y: totalPrimary - ln.pc}
	case domain.LeftRight:
		return domain.Point{X: ln.pc, Y: ln.cross}
	case domain.RightLeft:
		return domain.Point{X: totalPrimary - ln.pc, Y: ln.cross}
	default: // TopBottom
		return domain.Point{X: ln.cross, Y: ln.pc}
	}
}

// writeBackNodes sets each real node's top-left position from its center.
func writeBackNodes(lg *lgraph, g *domain.Graph, totalPrimary float64) {
	for _, layer := range lg.layers {
		for _, ln := range layer {
			if ln.real == nil {
				continue
			}
			c := ln.center(g.Direction, totalPrimary)
			ln.real.Pos = domain.Point{X: c.X - ln.w/2, Y: c.Y - ln.h/2}
		}
	}
}

// routeEdges builds each edge's polyline through its dummy chain, clipping the
// first and last segments to the node boundaries so arrowheads land on the
// box edge. Self-edges are routed as a small loop beside the node.
func routeEdges(lg *lgraph, g *domain.Graph, totalPrimary float64) {
	for _, e := range g.Edges {
		chain := lg.chains[e]
		if len(chain) < 2 {
			continue
		}
		if e.From == e.To {
			e.Points = selfLoop(chain[0], g.Direction, totalPrimary)
			continue
		}
		pts := make([]domain.Point, len(chain))
		for i, ln := range chain {
			pts[i] = ln.center(g.Direction, totalPrimary)
		}
		vertical := g.Direction == domain.TopBottom || g.Direction == domain.BottomTop
		pts = orthogonalize(pts, vertical)
		from, to := chain[0], chain[len(chain)-1]
		pts[0] = clipToBox(pts[0], domain.Size{W: from.w, H: from.h}, pts[1])
		last := len(pts) - 1
		pts[last] = clipToBox(pts[last], domain.Size{W: to.w, H: to.h}, pts[last-1])
		e.Points = pts
	}
}

// orthogonalize converts a polyline of waypoint centers into a right-angle
// (Manhattan) path. Between consecutive points it inserts an elbow at the
// midpoint of the primary axis, so segments are either horizontal or vertical.
// Aligned points produce no elbow, keeping straight edges straight.
func orthogonalize(pts []domain.Point, vertical bool) []domain.Point {
	if len(pts) < 2 {
		return pts
	}
	out := []domain.Point{pts[0]}
	for i := 1; i < len(pts); i++ {
		a, b := out[len(out)-1], pts[i]
		if vertical {
			if a.X != b.X {
				mid := (a.Y + b.Y) / 2
				out = append(out, domain.Point{X: a.X, Y: mid}, domain.Point{X: b.X, Y: mid})
			}
		} else {
			if a.Y != b.Y {
				mid := (a.X + b.X) / 2
				out = append(out, domain.Point{X: mid, Y: a.Y}, domain.Point{X: mid, Y: b.Y})
			}
		}
		out = append(out, b)
	}
	return out
}

// selfLoop returns a small rectangular loop on the trailing side of a node.
func selfLoop(ln *lnode, dir domain.Direction, totalPrimary float64) []domain.Point {
	c := ln.center(dir, totalPrimary)
	const out = 24.0
	rx := c.X + ln.w/2
	qy := ln.h / 4
	return []domain.Point{
		{X: rx, Y: c.Y - qy},
		{X: rx + out, Y: c.Y - qy},
		{X: rx + out, Y: c.Y + qy},
		{X: rx, Y: c.Y + qy},
	}
}

// clipToBox returns the point where the segment from center toward target
// crosses the boundary of an axis-aligned box of the given size centered at
// center. If target coincides with center, center is returned.
func clipToBox(center domain.Point, size domain.Size, target domain.Point) domain.Point {
	dx, dy := target.X-center.X, target.Y-center.Y
	if dx == 0 && dy == 0 {
		return center
	}
	hw, hh := size.W/2, size.H/2
	t := math.Inf(1)
	if dx != 0 {
		t = math.Min(t, hw/math.Abs(dx))
	}
	if dy != 0 {
		t = math.Min(t, hh/math.Abs(dy))
	}
	return domain.Point{X: center.X + dx*t, Y: center.Y + dy*t}
}

// bounds computes the diagram extent over node boxes and routed edge points.
func bounds(g *domain.Graph) (width, height float64) {
	for _, n := range g.Nodes {
		if r := n.Pos.X + n.Size.W; r > width {
			width = r
		}
		if b := n.Pos.Y + n.Size.H; b > height {
			height = b
		}
	}
	for _, e := range g.Edges {
		for _, p := range e.Points {
			if p.X > width {
				width = p.X
			}
			if p.Y > height {
				height = p.Y
			}
		}
	}
	return width, height
}
