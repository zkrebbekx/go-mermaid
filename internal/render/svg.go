// Package render turns a laid-out graph into SVG bytes. Output is
// deterministic so it can be compared against golden files.
package render

import (
	"fmt"
	"math"
	"strings"

	"github.com/Zac300/go-mermaid/internal/domain"
	"github.com/Zac300/go-mermaid/internal/layout"
	"github.com/Zac300/go-mermaid/internal/svgutil"
	"github.com/Zac300/go-mermaid/internal/theme"
)

// Options controls SVG appearance.
type Options struct {
	Theme    string
	FontFace string
	FontSize float64
	Padding  float64
	Title    string
	Curved   bool
}

// smoothPath builds a quadratic-smoothed path through the waypoints, rounding
// the interior corners of an orthogonal polyline.
func smoothPath(pts []domain.Point) string {
	var b strings.Builder
	fmt.Fprintf(&b, "M%s,%s ", num(pts[0].X), num(pts[0].Y))
	for i := 1; i < len(pts); i++ {
		mx, my := (pts[i-1].X+pts[i].X)/2, (pts[i-1].Y+pts[i].Y)/2
		fmt.Fprintf(&b, "Q%s,%s %s,%s ", num(pts[i-1].X), num(pts[i-1].Y), num(mx), num(my))
	}
	last := pts[len(pts)-1]
	fmt.Fprintf(&b, "L%s,%s", num(last.X), num(last.Y))
	return strings.TrimSpace(b.String())
}

// writeTitle draws a centered, bold diagram title at (x, y) if non-empty.
func writeTitle(b *strings.Builder, title string, x, y float64, pal theme.Palette) {
	if title == "" {
		return
	}
	fmt.Fprintf(b, `  <text x="%s" y="%s" fill="%s" text-anchor="middle" font-weight="bold">%s</text>`,
		num(x), num(y), pal.Text, esc(title))
	b.WriteByte('\n')
}

// SVG renders a laid-out graph to an SVG document.
func SVG(res *layout.Result, opts Options) ([]byte, error) {
	pal := theme.For(opts.Theme)
	pad := opts.Padding
	titleH := svgutil.TitleHeight(opts.Title, opts.FontSize)

	// Union node bounds with subgraph boxes (which can extend past the nodes
	// and into negative coordinates) so nothing clips.
	minX, minY, maxX, maxY := 0.0, 0.0, res.Width, res.Height
	for _, sg := range res.Graph.Subgraphs {
		if bx, by, bw, bh, ok := subgraphBox(sg, res.Graph, opts); ok {
			minX, minY = math.Min(minX, bx), math.Min(minY, by)
			maxX, maxY = math.Max(maxX, bx+bw), math.Max(maxY, by+bh)
		}
	}
	contentW := maxX - minX
	if tw := opts.FontSize * 0.6 * float64(len([]rune(opts.Title))); tw > contentW {
		contentW = tw
	}
	shiftX, shiftY := -minX, -minY
	w := contentW + pad*2
	h := (maxY - minY) + titleH + pad*2

	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%s" height="%s" viewBox="0 0 %s %s" font-family="%s" font-size="%s">`,
		num(w), num(h), num(w), num(h), esc(opts.FontFace), num(opts.FontSize))
	b.WriteByte('\n')

	// Arrowhead marker.
	fmt.Fprintf(&b, `  <defs><marker id="arrow" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="7" markerHeight="7" orient="auto-start-reverse"><path d="M0,0 L10,5 L0,10 z" fill="%s"/></marker></defs>`, pal.Edge)
	b.WriteByte('\n')

	fmt.Fprintf(&b, `  <rect width="100%%" height="100%%" fill="%s"/>`, pal.Background)
	b.WriteByte('\n')

	writeTitle(&b, opts.Title, w/2, pad+opts.FontSize, pal)

	fmt.Fprintf(&b, `  <g transform="translate(%s,%s)">`, num(pad+shiftX), num(pad+titleH+shiftY))
	b.WriteByte('\n')

	for _, sg := range res.Graph.Subgraphs {
		writeSubgraph(&b, sg, res.Graph, pal, opts)
	}
	for _, e := range res.Graph.Edges {
		writeEdge(&b, e, pal, opts.Curved)
	}
	for _, n := range res.Graph.Nodes {
		writeNode(&b, n, pal, opts)
	}

	b.WriteString("  </g>\n</svg>\n")
	return []byte(b.String()), nil
}

func writeEdge(b *strings.Builder, e *domain.Edge, pal theme.Palette, curved bool) {
	if len(e.Points) < 2 {
		return
	}
	var d strings.Builder
	if curved && len(e.Points) > 2 {
		d.WriteString(smoothPath(e.Points))
	} else {
		for i, p := range e.Points {
			cmd := "L"
			if i == 0 {
				cmd = "M"
			}
			fmt.Fprintf(&d, "%s%s,%s ", cmd, num(p.X), num(p.Y))
		}
	}
	dash := ""
	if e.Arrow == domain.ArrowDotted {
		dash = ` stroke-dasharray="4,4"`
	}
	width := "1.5"
	if e.Arrow == domain.ArrowThick {
		width = "3"
	}
	marker := ` marker-end="url(#arrow)"`
	if e.Arrow == domain.ArrowOpen {
		marker = ""
	}
	fmt.Fprintf(b, `    <path d="%s" fill="none" stroke="%s" stroke-width="%s"%s%s/>`,
		strings.TrimSpace(d.String()), pal.Edge, width, dash, marker)
	b.WriteByte('\n')

	if e.Label != "" {
		// Midpoint of the polyline's endpoints, not Points[len/2] (which is
		// the target endpoint for a two-point line).
		first, last := e.Points[0], e.Points[len(e.Points)-1]
		midX, midY := (first.X+last.X)/2, (first.Y+last.Y)/2
		fmt.Fprintf(b, `    <text x="%s" y="%s" fill="%s" text-anchor="middle" dy="-2">%s</text>`,
			num(midX), num(midY), pal.Text, esc(e.Label))
		b.WriteByte('\n')
	}
}

// subgraphBox returns the cluster box (x, y, w, h) enclosing a subgraph's
// member nodes, with padding and title space. ok is false if it has no
// positioned members.
func subgraphBox(sg *domain.Subgraph, g *domain.Graph, opts Options) (x, y, w, h float64, ok bool) {
	const pad = 14.0
	first := true
	var minX, minY, maxX, maxY float64
	for _, id := range sg.NodeIDs {
		n := g.NodeByID(id)
		if n == nil {
			continue
		}
		if first {
			minX, minY = n.Pos.X, n.Pos.Y
			maxX, maxY = n.Pos.X+n.Size.W, n.Pos.Y+n.Size.H
			first = false
			continue
		}
		minX = math.Min(minX, n.Pos.X)
		minY = math.Min(minY, n.Pos.Y)
		maxX = math.Max(maxX, n.Pos.X+n.Size.W)
		maxY = math.Max(maxY, n.Pos.Y+n.Size.H)
	}
	if first {
		return 0, 0, 0, 0, false
	}
	titleH := 0.0
	if sg.Title != "" {
		titleH = opts.FontSize + 6
	}
	return minX - pad, minY - pad - titleH, maxX - minX + 2*pad, maxY - minY + 2*pad + titleH, true
}

// writeSubgraph draws a dashed cluster box around a subgraph's member nodes.
func writeSubgraph(b *strings.Builder, sg *domain.Subgraph, g *domain.Graph, pal theme.Palette, opts Options) {
	x, y, w, h, ok := subgraphBox(sg, g, opts)
	if !ok {
		return
	}
	fmt.Fprintf(b, `    <rect x="%s" y="%s" width="%s" height="%s" fill="none" stroke="%s" stroke-dasharray="4,3" rx="4"/>`,
		num(x), num(y), num(w), num(h), pal.NodeStroke)
	b.WriteByte('\n')
	if sg.Title != "" {
		fmt.Fprintf(b, `    <text x="%s" y="%s" fill="%s">%s</text>`,
			num(x+6), num(y+opts.FontSize), pal.Text, esc(sg.Title))
		b.WriteByte('\n')
	}
}

func writeNode(b *strings.Builder, n *domain.Node, pal theme.Palette, opts Options) {
	if n.Link != "" {
		fmt.Fprintf(b, `    <a href="%s" target="_blank">`, esc(n.Link))
		b.WriteByte('\n')
	}
	x, y, w, h := n.Pos.X, n.Pos.Y, n.Size.W, n.Size.H
	fill, stroke, textColor := pal.NodeFill, pal.NodeStroke, pal.Text
	if n.Style != nil {
		if n.Style.Fill != "" {
			fill = n.Style.Fill
		}
		if n.Style.Stroke != "" {
			stroke = n.Style.Stroke
		}
		if n.Style.Color != "" {
			textColor = n.Style.Color
		}
	}
	switch n.Shape {
	case domain.ShapeRound, domain.ShapeStadium:
		rx := h / 2
		if n.Shape == domain.ShapeRound {
			rx = 6
		}
		fmt.Fprintf(b, `    <rect x="%s" y="%s" width="%s" height="%s" rx="%s" fill="%s" stroke="%s"/>`,
			num(x), num(y), num(w), num(h), num(rx), fill, stroke)
	case domain.ShapeCircle:
		fmt.Fprintf(b, `    <circle cx="%s" cy="%s" r="%s" fill="%s" stroke="%s"/>`,
			num(x+w/2), num(y+h/2), num(w/2), fill, stroke)
	case domain.ShapeDiamond:
		cx, cy := x+w/2, y+h/2
		pts := fmt.Sprintf("%s,%s %s,%s %s,%s %s,%s",
			num(cx), num(y), num(x+w), num(cy), num(cx), num(y+h), num(x), num(cy))
		fmt.Fprintf(b, `    <polygon points="%s" fill="%s" stroke="%s"/>`, pts, fill, stroke)
	case domain.ShapeHexagon:
		k := h / 2
		pts := fmt.Sprintf("%s,%s %s,%s %s,%s %s,%s %s,%s %s,%s",
			num(x), num(y+h/2), num(x+k), num(y), num(x+w-k), num(y),
			num(x+w), num(y+h/2), num(x+w-k), num(y+h), num(x+k), num(y+h))
		fmt.Fprintf(b, `    <polygon points="%s" fill="%s" stroke="%s"/>`, pts, fill, stroke)
	case domain.ShapeParallelogram:
		k := h / 2
		pts := fmt.Sprintf("%s,%s %s,%s %s,%s %s,%s",
			num(x+k), num(y), num(x+w), num(y), num(x+w-k), num(y+h), num(x), num(y+h))
		fmt.Fprintf(b, `    <polygon points="%s" fill="%s" stroke="%s"/>`, pts, fill, stroke)
	case domain.ShapeParallelogramAlt:
		k := h / 2
		pts := fmt.Sprintf("%s,%s %s,%s %s,%s %s,%s",
			num(x), num(y), num(x+w-k), num(y), num(x+w), num(y+h), num(x+k), num(y+h))
		fmt.Fprintf(b, `    <polygon points="%s" fill="%s" stroke="%s"/>`, pts, fill, stroke)
	case domain.ShapeTrapezoid:
		k := h / 2
		pts := fmt.Sprintf("%s,%s %s,%s %s,%s %s,%s",
			num(x+k), num(y), num(x+w-k), num(y), num(x+w), num(y+h), num(x), num(y+h))
		fmt.Fprintf(b, `    <polygon points="%s" fill="%s" stroke="%s"/>`, pts, fill, stroke)
	case domain.ShapeTrapezoidAlt:
		k := h / 2
		pts := fmt.Sprintf("%s,%s %s,%s %s,%s %s,%s",
			num(x), num(y), num(x+w), num(y), num(x+w-k), num(y+h), num(x+k), num(y+h))
		fmt.Fprintf(b, `    <polygon points="%s" fill="%s" stroke="%s"/>`, pts, fill, stroke)
	case domain.ShapeCylinder:
		ry := h * 0.12
		fmt.Fprintf(b, `    <path d="M%s,%s L%s,%s A%s,%s 0 0 0 %s,%s A%s,%s 0 0 0 %s,%s Z" fill="%s" stroke="%s"/>`,
			num(x), num(y+ry), num(x), num(y+h-ry),
			num(w/2), num(ry), num(x+w), num(y+h-ry),
			num(w/2), num(ry), num(x), num(y+ry),
			fill, stroke)
		fmt.Fprintf(b, `<path d="M%s,%s A%s,%s 0 0 0 %s,%s" fill="none" stroke="%s"/>`,
			num(x), num(y+ry), num(w/2), num(ry), num(x+w), num(y+ry), stroke)
	case domain.ShapeSubroutine:
		fmt.Fprintf(b, `    <rect x="%s" y="%s" width="%s" height="%s" fill="%s" stroke="%s"/>`,
			num(x), num(y), num(w), num(h), fill, stroke)
		fmt.Fprintf(b, `<line x1="%s" y1="%s" x2="%s" y2="%s" stroke="%s"/><line x1="%s" y1="%s" x2="%s" y2="%s" stroke="%s"/>`,
			num(x+6), num(y), num(x+6), num(y+h), stroke,
			num(x+w-6), num(y), num(x+w-6), num(y+h), stroke)
	default: // rect
		fmt.Fprintf(b, `    <rect x="%s" y="%s" width="%s" height="%s" fill="%s" stroke="%s"/>`,
			num(x), num(y), num(w), num(h), fill, stroke)
	}
	b.WriteByte('\n')

	label := n.Label
	if label == "" {
		label = n.ID
	}
	b.WriteString("    ")
	svgutil.MultilineText(b, svgutil.SplitLines(label), x+w/2, y+h/2+opts.FontSize*0.35, opts.FontSize+2, textColor, "")
	b.WriteByte('\n')
	if n.Link != "" {
		b.WriteString("    </a>\n")
	}
}
