// Package layout assigns coordinates to a domain.Graph using a layered
// (Sugiyama-style) approach: make the graph acyclic, rank nodes into layers,
// insert dummy nodes so long edges can bend, order within layers to reduce
// crossings, then assign positions with a barycenter heuristic. A labelled
// edge's middle dummy is sized to its label, so every stage reserves room for
// the text the way it does for nodes. Ranking is
// longest-path; network-simplex can replace it behind the same interface.
package layout

import (
	"math"
	"slices"
	"sort"

	"github.com/zkrebbekx/go-mermaid/internal/domain"
	"github.com/zkrebbekx/go-mermaid/internal/svgutil"
)

// Options tunes spacing and text metrics used during layout.
type Options struct {
	NodeSep  float64 // gap between nodes within a layer
	RankSep  float64 // gap between layers
	FontSize float64 // used to estimate node sizes
	FontFace string  // CSS font-family the SVG will ask for; picks the metrics
}

// face returns the metric table that matches the font family the renderer
// will name in the SVG. Measuring with a different face than the viewer
// draws with makes every box the wrong size.
func (o Options) face() svgutil.Face { return svgutil.FaceFor(o.FontFace) }

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

	// Give every labelled edge a middle rank for its label dummy. Halving the
	// gap keeps the distance between real ranks unchanged when that middle
	// rank holds no label.
	popts := opts
	labelled := hasEdgeLabels(g)
	if labelled {
		for id := range ranks {
			ranks[id] *= 2
		}
		popts.RankSep /= 2
	}

	lg := buildLGraph(g, ranks, opts)
	reduceCrossings(lg)
	totalPrimary := positionLG(lg, g.Direction, popts)

	writeBackNodes(lg, g, totalPrimary)
	// Restore original edge directions before routing so arrowheads point the
	// right way on edges that were reversed to break cycles.
	restoreReversed(g, reversed)
	anchored := routeEdges(lg, g, totalPrimary)
	vertical := isVertical(g.Direction)
	// Edges that join the same node pair share one path unless each has its
	// own dummy, which is always so once ranks are doubled.
	var parallel [][]*domain.Edge
	if !labelled {
		parallel = parallelGroups(g)
	}
	separateParallel(g, parallel)
	spreadPorts(g, vertical, parallel, anchored, opts)
	dropRepeats(g)
	placeLabels(g, anchored, vertical, opts)
	normalizeOrigin(g, opts)

	w, h := bounds(g, opts)
	return &Result{Graph: g, Width: w, Height: h}, nil
}

func isVertical(dir domain.Direction) bool {
	return dir == domain.TopBottom || dir == domain.BottomTop
}

// hasEdgeLabels reports whether any edge between two distinct nodes carries
// a label. Self-loop labels sit beside their node and need no rank.
func hasEdgeLabels(g *domain.Graph) bool {
	for _, e := range g.Edges {
		if e.Label != "" && e.From != e.To {
			return true
		}
	}
	return false
}

// parallelSep is the perpendicular distance between edges that join the same
// node pair.
const parallelSep = 34.0

// parallelGroups returns the edges that share an unordered node pair, grouped
// in first-seen order. Self-loops and lone edges are left out.
func parallelGroups(g *domain.Graph) [][]*domain.Edge {
	key := func(a, b string) string {
		if a < b {
			return a + "\x00" + b
		}
		return b + "\x00" + a
	}
	idx := map[string]int{}
	var groups [][]*domain.Edge
	for _, e := range g.Edges {
		if e.From == e.To {
			continue
		}
		k := key(e.From, e.To)
		i, ok := idx[k]
		if !ok {
			i = len(groups)
			idx[k] = i
			groups = append(groups, nil)
		}
		groups[i] = append(groups[i], e)
	}
	var out [][]*domain.Edge
	for _, es := range groups {
		if len(es) >= 2 {
			out = append(out, es)
		}
	}
	return out
}

// pairAxis returns the unit vector from the lower-ID node of the pair to the
// higher-ID node, measured between node centers. Every edge in a parallel
// group is offset against this one axis, so A->B and B->A move apart even
// though their own polylines run in opposite directions.
func pairAxis(g *domain.Graph, e *domain.Edge) (ax, ay float64, ok bool) {
	lo, hi := e.From, e.To
	if hi < lo {
		lo, hi = hi, lo
	}
	a, b := g.NodeByID(lo), g.NodeByID(hi)
	if a == nil || b == nil {
		return 0, 0, false
	}
	ca, cb := a.Center(), b.Center()
	d := math.Hypot(cb.X-ca.X, cb.Y-ca.Y)
	if d == 0 {
		return 0, 0, false
	}
	return (cb.X - ca.X) / d, (cb.Y - ca.Y) / d, true
}

// separateParallel offsets each group of edges that share a node pair
// perpendicular to the pair axis, so overlapping lines (e.g. A->B and B->A)
// don't sit on top of each other.
func separateParallel(g *domain.Graph, groups [][]*domain.Edge) {
	for _, es := range groups {
		ax, ay, ok := pairAxis(g, es[0])
		if !ok {
			continue
		}
		px, py := -ay, ax
		for i, e := range es {
			off := (float64(i) - float64(len(es)-1)/2) * parallelSep
			for j := range e.Points {
				e.Points[j].X += px * off
				e.Points[j].Y += py * off
			}
			e.LabelPos.X += px * off
			e.LabelPos.Y += py * off
		}
	}
}

// labelSize estimates the box a renderer draws around an edge label.
func labelSize(label string, face svgutil.Face, fontSize float64) domain.Size {
	lines := svgutil.SplitLines(label)
	w := 0.0
	for _, ln := range lines {
		if lw := face.Width(ln, fontSize); lw > w {
			w = lw
		}
	}
	return domain.Size{W: w + 6, H: fontSize*float64(len(lines)) + 4}
}

// labelRect returns the estimated box of e's label around LabelPos. The box
// matches the background rect the flowchart renderer draws: the text
// baseline sits near the bottom, so most of the box is above LabelPos.
func labelRect(e *domain.Edge, face svgutil.Face, fontSize float64) (domain.Rect, bool) {
	if e.Label == "" || len(e.Points) == 0 {
		return domain.Rect{}, false
	}
	sz := labelSize(e.Label, face, fontSize)
	// LabelPos is the baseline of the last line, so the box grows upward.
	return domain.Rect{
		Min:  domain.Point{X: e.LabelPos.X - sz.W/2, Y: e.LabelPos.Y - (sz.H - 4)},
		Size: sz,
	}, true
}

// labelGap is the clear space between a label box and the line it labels.
const labelGap = 4.0

// placeLabels sets LabelPos on every edge. LabelPos is the baseline of the
// label's last line (see labelRect). An edge routed through a label dummy
// already holds that dummy's centre, which positioning kept clear of other
// nodes and labels. Self-loop labels stack in a column just outside the
// outermost loop on their node. Any other edge anchors its label at the
// midpoint of its path.
func placeLabels(g *domain.Graph, anchored map[*domain.Edge]bool, vertical bool, opts Options) {
	baseline := func(c domain.Point, sz domain.Size) domain.Point {
		return domain.Point{X: c.X, Y: c.Y + sz.H/2 - 4}
	}
	type column struct {
		outer  float64 // far side of the outermost loop on the cross axis
		height float64 // total height of the stacked labels
		edges  []*domain.Edge
	}
	cols := map[string]*column{}
	var order []string
	for _, e := range g.Edges {
		switch {
		case e.Label == "" || len(e.Points) == 0:
			e.LabelPos = domain.PolylineMidpoint(e.Points)
		case anchored[e]:
			e.LabelPos = baseline(e.LabelPos, labelSize(e.Label, opts.face(), opts.FontSize))
		case e.From == e.To && len(e.Points) == 4:
			col, ok := cols[e.From]
			if !ok {
				col = &column{outer: math.Inf(-1)}
				cols[e.From] = col
				order = append(order, e.From)
			}
			col.edges = append(col.edges, e)
			col.height += labelSize(e.Label, opts.face(), opts.FontSize).H
		default:
			e.LabelPos = domain.PolylineMidpoint(e.Points)
		}
	}
	if len(cols) == 0 {
		return
	}
	// Unlabelled loops nest too, so the outermost may carry no label.
	for _, e := range g.Edges {
		if col := cols[e.From]; col != nil && e.From == e.To && len(e.Points) == 4 {
			if vertical {
				col.outer = math.Max(col.outer, e.Points[1].X)
			} else {
				col.outer = math.Max(col.outer, e.Points[1].Y)
			}
		}
	}
	for _, id := range order {
		col, n := cols[id], g.NodeByID(id)
		if n == nil {
			continue
		}
		nc := n.Center()
		// Labels run down from the top of the column, which is centred on
		// the node in TD and starts just below the loops in LR.
		at := nc.Y - col.height/2
		if !vertical {
			at = col.outer + labelGap
		}
		for _, e := range col.edges {
			sz := labelSize(e.Label, opts.face(), opts.FontSize)
			c := domain.Point{X: nc.X, Y: at + sz.H/2}
			if vertical {
				c.X = col.outer + labelGap + sz.W/2
			}
			e.LabelPos = baseline(c, sz)
			at += sz.H
		}
	}
}

// normalizeOrigin shifts the whole drawing so no node, edge point, or label
// box has a negative coordinate. Parallel-edge offsets and wide labels can
// push content past the left or top edge; without this the renderer would
// clip them.
func normalizeOrigin(g *domain.Graph, opts Options) {
	minX, minY := 0.0, 0.0
	for _, n := range g.Nodes {
		minX, minY = math.Min(minX, n.Pos.X), math.Min(minY, n.Pos.Y)
	}
	for _, e := range g.Edges {
		for _, p := range e.Points {
			minX, minY = math.Min(minX, p.X), math.Min(minY, p.Y)
		}
		if r, ok := labelRect(e, opts.face(), opts.FontSize); ok {
			minX, minY = math.Min(minX, r.Min.X), math.Min(minY, r.Min.Y)
		}
	}
	if minX == 0 && minY == 0 {
		return
	}
	dx, dy := -minX, -minY
	for _, n := range g.Nodes {
		n.Pos.X += dx
		n.Pos.Y += dy
	}
	for _, e := range g.Edges {
		for j := range e.Points {
			e.Points[j].X += dx
			e.Points[j].Y += dy
		}
		e.LabelPos.X += dx
		e.LabelPos.Y += dy
	}
}

// spreadPorts fans out the attach points of edges that meet a node on the same
// face, so multiple edges (and their arrowheads) don't pile up at the box
// center. Endpoints are distributed evenly along the face and ordered by their
// far end to avoid introducing crossings; see shiftPort for how each path
// follows its endpoint. Edges in parallel groups are left alone, since
// separateParallel already fanned them out wider. anchored holds the edges
// whose LabelPos is still the centre of their label dummy.
func spreadPorts(g *domain.Graph, vertical bool, parallel [][]*domain.Edge, anchored map[*domain.Edge]bool, opts Options) {
	type touch struct {
		e    *domain.Edge
		last bool // the edge's last point touches this node, not its first
	}
	separated := map[*domain.Edge]bool{}
	for _, es := range parallel {
		for _, e := range es {
			separated[e] = true
		}
	}
	byNode := map[string][]touch{}
	for _, e := range g.Edges {
		// A two-point edge has no interior point to bend at: moving one end
		// would drag the other node's attach point with it.
		if e.From == e.To || len(e.Points) < 3 || separated[e] {
			continue
		}
		byNode[e.From] = append(byNode[e.From], touch{e, false})
		byNode[e.To] = append(byNode[e.To], touch{e, true})
	}
	// Ends are looked up on use: a jog inserted at one end shifts the other.
	end := func(t touch) domain.Point {
		if t.last {
			return t.e.Points[len(t.e.Points)-1]
		}
		return t.e.Points[0]
	}
	far := func(t touch) domain.Point {
		if t.last {
			return t.e.Points[0]
		}
		return t.e.Points[len(t.e.Points)-1]
	}
	// Iterate nodes in a stable order so output never depends on map iteration.
	ids := make([]string, 0, len(byNode))
	for id := range byNode {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		ts := byNode[id]
		n := g.NodeByID(id)
		if n == nil {
			continue
		}
		c := n.Center()
		faces := map[bool][]touch{}
		for _, t := range ts {
			p := end(t)
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
			sort.SliceStable(grp, func(i, j int) bool {
				if vertical {
					return far(grp[i]).X < far(grp[j]).X
				}
				return far(grp[i]).Y < far(grp[j]).Y
			})
			for k, t := range grp {
				frac := float64(k+1) / float64(len(grp)+1)
				port := n.Pos.Y + n.Size.H*frac
				if vertical {
					port = n.Pos.X + n.Size.W*frac
				}
				var inset float64
				if anchored[t.e] {
					sz := labelSize(t.e.Label, opts.face(), opts.FontSize)
					inset = sz.H / 2
					if !vertical {
						inset = sz.W / 2
					}
				}
				shiftPort(t.e, t.last, port, vertical, inset)
			}
		}
	}
}

// shiftPort moves one end of e to port along the node face. When the point
// next to the end is an elbow, it moves too and the path stays orthogonal.
// When the path runs straight on through it (a dummy, such as a label
// anchor), a jog is inserted halfway instead so the rest of the path, and the
// label on it, stays put. If that dummy is e's label, inset is half the
// label's extent along the path, and the jog goes halfway to the label's edge
// rather than its centre so the label never hides it.
func shiftPort(e *domain.Edge, last bool, port float64, vertical bool, inset float64) {
	pts := e.Points
	if last {
		pts = slices.Clone(pts)
		slices.Reverse(pts)
	}
	primary := func(p domain.Point) float64 {
		if vertical {
			return p.Y
		}
		return p.X
	}
	cur := pts[0].Y
	if vertical {
		cur = pts[0].X
	}
	if cur == port {
		return
	}
	set := func(p *domain.Point, v float64) {
		if vertical {
			p.X = v
		} else {
			p.Y = v
		}
	}
	// An elbow turns onto the cross axis, so the segment after it keeps the
	// same primary coordinate.
	if primary(pts[1]) == primary(pts[2]) {
		set(&pts[0], port)
		set(&pts[1], port)
	} else {
		mid := pts[1]
		if !(inset > 0 && pts[1] == e.LabelPos) {
			inset = 0
		}
		a, b := primary(pts[0]), primary(pts[1])
		if b < a {
			inset = -inset
		}
		jog := (a + b - inset) / 2
		if vertical {
			mid.Y = jog
		} else {
			mid.X = jog
		}
		near := mid
		set(&near, port)
		set(&pts[0], port)
		pts = slices.Insert(pts, 1, near, mid)
	}
	if last {
		slices.Reverse(pts)
	}
	e.Points = pts
}

// dropRepeats removes consecutive duplicate points from every edge. Moving a
// port onto the next bend can leave a zero-length segment behind.
func dropRepeats(g *domain.Graph) {
	for _, e := range g.Edges {
		e.Points = slices.Compact(e.Points)
	}
}

// sizeNodes estimates a box size for each node from its label and font size.
// Nodes that already have a non-zero size (e.g. class boxes sized by their
// own renderer) are left untouched.
func sizeNodes(g *domain.Graph, opts Options) {
	const padX, padY = 20.0, 14.0
	face := opts.face()
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
			if wd := face.Width(ln, opts.FontSize); wd > maxW {
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
//
// An edge routed through a label dummy gets that dummy's centre in LabelPos,
// and is reported in the returned set; placeLabels finishes the anchor.
func routeEdges(lg *lgraph, g *domain.Graph, totalPrimary float64) map[*domain.Edge]bool {
	anchored := map[*domain.Edge]bool{}
	vertical := isVertical(g.Direction)
	for _, e := range g.Edges {
		chain := lg.chains[e]
		if len(chain) < 2 {
			continue
		}
		if e.From == e.To {
			e.Points = selfLoop(chain[0], lg.loopIdx[e], vertical, g.Direction, totalPrimary)
			continue
		}
		pts := make([]domain.Point, len(chain))
		half := make([]float64, len(chain))
		for i, ln := range chain {
			pts[i] = ln.center(g.Direction, totalPrimary)
			half[i] = primarySize(ln, vertical) / 2
			if ln.label == e {
				e.LabelPos = pts[i]
				anchored[e] = true
			}
		}
		pts = orthogonalize(pts, half, vertical)
		from, to := chain[0], chain[len(chain)-1]
		pts[0] = clipToBox(pts[0], domain.Size{W: from.w, H: from.h}, pts[1])
		last := len(pts) - 1
		pts[last] = clipToBox(pts[last], domain.Size{W: to.w, H: to.h}, pts[last-1])
		// The chain was built while the edge was reversed to break a cycle.
		// The edge direction is restored by now, so flip the points too or
		// the arrowhead lands on the source node.
		if from.real != nil && from.real.ID != e.From {
			for i, j := 0, len(pts)-1; i < j; i, j = i+1, j-1 {
				pts[i], pts[j] = pts[j], pts[i]
			}
		}
		e.Points = pts
	}
	return anchored
}

// orthogonalize converts a polyline of waypoint centers into a right-angle
// (Manhattan) path. Between consecutive points it inserts an elbow in the gap
// between the two boxes, so segments are either horizontal or vertical.
// Aligned points produce no elbow, keeping straight edges straight.
//
// half holds each waypoint's half-extent along the primary axis, and is zero
// for a dummy. The elbow must sit in the gap, not at the midpoint of the two
// centers: a node much taller than the rank gap puts that midpoint inside its
// own box, and the edge then doubles back through the node it just left.
func orthogonalize(pts []domain.Point, half []float64, vertical bool) []domain.Point {
	if len(pts) < 2 {
		return pts
	}
	out := []domain.Point{pts[0]}
	for i := 1; i < len(pts); i++ {
		a, b := pts[i-1], pts[i]
		ha, hb := half[i-1], half[i]
		if vertical {
			if a.X != b.X {
				mid := gapMid(a.Y, ha, b.Y, hb)
				out = append(out, domain.Point{X: a.X, Y: mid}, domain.Point{X: b.X, Y: mid})
			}
		} else {
			if a.Y != b.Y {
				mid := gapMid(a.X, ha, b.X, hb)
				out = append(out, domain.Point{X: mid, Y: a.Y}, domain.Point{X: mid, Y: b.Y})
			}
		}
		out = append(out, b)
	}
	return out
}

// gapMid returns the midpoint of the clear space between two boxes centered
// at ca and cb with half-extents ha and hb along the same axis. When the
// boxes overlap there is no gap, so it falls back to the midpoint of the
// centers.
func gapMid(ca, ha, cb, hb float64) float64 {
	lo, hi := ca+ha, cb-hb
	if cb < ca {
		lo, hi = cb+hb, ca-ha
	}
	if hi < lo {
		return (ca + cb) / 2
	}
	return (lo + hi) / 2
}

// loopOut is how far a self-loop reaches out from its node; each further
// loop on the same node nests loopStep beyond the previous one.
const (
	loopOut  = 24.0
	loopStep = 10.0
)

// loopReach is how far the outermost of n self-loops reaches from its node.
func loopReach(n int) float64 {
	return loopOut + float64(max(n-1, 0))*loopStep
}

// selfLoop returns a small rectangular loop on the trailing cross side of a
// node: the right in TD/BT, the bottom in LR/RL. That side never faces
// another rank, so the loop stays clear of the node's other edges. Loop k
// nests outside loops 0..k-1, reaching further and attaching wider apart.
func selfLoop(ln *lnode, k int, vertical bool, dir domain.Direction, totalPrimary float64) []domain.Point {
	c := ln.center(dir, totalPrimary)
	out := loopReach(k + 1)
	spread := func(size float64) float64 {
		return math.Min(size/4+float64(k)*loopStep/2, size/2)
	}
	if !vertical {
		by := c.Y + ln.h/2
		qx := spread(ln.w)
		return []domain.Point{
			{X: c.X - qx, Y: by},
			{X: c.X - qx, Y: by + out},
			{X: c.X + qx, Y: by + out},
			{X: c.X + qx, Y: by},
		}
	}
	rx := c.X + ln.w/2
	qy := spread(ln.h)
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

// bounds computes the diagram extent over node boxes, routed edge points,
// and edge label boxes.
func bounds(g *domain.Graph, opts Options) (width, height float64) {
	for _, n := range g.Nodes {
		width = math.Max(width, n.Pos.X+n.Size.W)
		height = math.Max(height, n.Pos.Y+n.Size.H)
	}
	for _, e := range g.Edges {
		for _, p := range e.Points {
			width = math.Max(width, p.X)
			height = math.Max(height, p.Y)
		}
		if r, ok := labelRect(e, opts.face(), opts.FontSize); ok {
			width = math.Max(width, r.Min.X+r.Size.W)
			height = math.Max(height, r.Min.Y+r.Size.H)
		}
	}
	return width, height
}
