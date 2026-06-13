// Package layout assigns coordinates to a domain.Graph using a layered
// (Sugiyama-style) approach: make the graph acyclic, rank nodes into
// layers, order within layers to reduce crossings, then assign pixel
// positions. v0 uses longest-path ranking; network-simplex can replace
// it behind the same interface.
package layout

import (
	"math"

	"github.com/Zac300/go-mermaid/internal/domain"
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

// Compute lays out g in place and returns the result. The input graph's
// nodes and edges are mutated with positions and routed points.
func Compute(g *domain.Graph, opts Options) (*Result, error) {
	sizeNodes(g, opts)

	reversed := makeAcyclic(g)
	ranks := assignRanks(g)
	layers := orderLayers(g, ranks)

	w, h := position(g, layers, opts)
	// Restore original edge directions before routing so arrowheads point
	// the right way on edges that were reversed to break cycles.
	restoreReversed(g, reversed)
	routeEdges(g)

	return &Result{Graph: g, Width: w, Height: h}, nil
}

// sizeNodes estimates a box size for each node from its label and font size.
func sizeNodes(g *domain.Graph, opts Options) {
	const padX, padY = 20.0, 14.0
	charW := opts.FontSize * 0.6
	for _, n := range g.Nodes {
		label := n.Label
		if label == "" {
			label = n.ID
		}
		w := float64(len([]rune(label)))*charW + padX*2
		h := opts.FontSize + padY*2
		if n.Shape == domain.ShapeCircle {
			if w < h {
				w = h
			}
			h = w
		}
		n.Size = domain.Size{W: w, H: h}
	}
}

// routeEdges sets a straight two-point polyline running from the boundary of
// the source box to the boundary of the target box, so arrowheads land on the
// node edge rather than hidden under its center. Orthogonal/spline routing is
// a later refinement.
func routeEdges(g *domain.Graph) {
	for _, e := range g.Edges {
		from := g.NodeByID(e.From)
		to := g.NodeByID(e.To)
		if from == nil || to == nil {
			continue
		}
		fc, tc := from.Center(), to.Center()
		start := clipToBox(fc, from.Size, tc)
		end := clipToBox(tc, to.Size, fc)
		e.Points = []domain.Point{start, end}
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
