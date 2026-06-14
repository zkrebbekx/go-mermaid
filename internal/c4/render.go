package c4

import (
	"fmt"
	"strings"

	"github.com/zkrebbekx/go-mermaid/internal/domain"
	"github.com/zkrebbekx/go-mermaid/internal/layout"
	"github.com/zkrebbekx/go-mermaid/internal/svgutil"
	"github.com/zkrebbekx/go-mermaid/internal/theme"
)

// RenderOptions controls C4 appearance.
type RenderOptions struct {
	Theme    string
	FontFace string
	FontSize float64
	Padding  float64
	Title    string
}

// kindFill maps an element kind to a fill color.
func kindFill(kind string) string {
	switch {
	case strings.HasPrefix(kind, "Person"):
		if strings.Contains(kind, "_Ext") {
			return "#686868"
		}
		return "#08427b"
	case strings.Contains(kind, "_Ext"):
		return "#999999"
	default:
		return "#1168bd"
	}
}

// Render parses and renders C4 source to SVG.
func Render(src string, o RenderOptions) ([]byte, error) {
	d, err := Parse(src)
	if err != nil {
		return nil, err
	}

	g := &domain.Graph{Direction: domain.TopBottom}
	for _, e := range d.Elements {
		n := &domain.Node{ID: e.ID, Label: e.Label, Shape: domain.ShapeRect}
		n.Size = elementSize(e, o.FontSize)
		g.Nodes = append(g.Nodes, n)
	}
	for _, r := range d.Rels {
		g.Edges = append(g.Edges, &domain.Edge{From: r.From, To: r.To, Label: r.Label})
	}
	res, err := layout.Compute(g, layout.Options{NodeSep: 60, RankSep: 100, FontSize: o.FontSize})
	if err != nil {
		return nil, err
	}
	return svg(d, g, res, o), nil
}

func svg(d *Diagram, g *domain.Graph, res *layout.Result, o RenderOptions) []byte {
	pal := theme.For(o.Theme)
	pad := o.Padding
	titleH := svgutil.TitleHeight(o.Title, o.FontSize)
	w := res.Width + pad*2
	h := res.Height + titleH + pad*2

	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%s" height="%s" viewBox="0 0 %s %s" font-family="%s" font-size="%s">`,
		svgutil.Num(w), svgutil.Num(h), svgutil.Num(w), svgutil.Num(h), svgutil.Esc(o.FontFace), svgutil.Num(o.FontSize))
	b.WriteByte('\n')
	fmt.Fprintf(&b, `  <defs><marker id="c4-arrow" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="7" markerHeight="7" orient="auto-start-reverse"><path d="M0,0 L10,5 L0,10 z" fill="%s"/></marker></defs>`, pal.Edge)
	b.WriteByte('\n')
	fmt.Fprintf(&b, `  <rect width="100%%" height="100%%" fill="%s"/>`, pal.Background)
	b.WriteByte('\n')
	if o.Title != "" {
		fmt.Fprintf(&b, `  <text x="%s" y="%s" fill="%s" text-anchor="middle" font-weight="bold">%s</text>`,
			svgutil.Num(w/2), svgutil.Num(pad+o.FontSize), pal.Text, svgutil.Esc(o.Title))
		b.WriteByte('\n')
	}
	fmt.Fprintf(&b, `  <g transform="translate(%s,%s)">`, svgutil.Num(pad), svgutil.Num(pad+titleH))
	b.WriteByte('\n')

	for i, r := range d.Rels {
		writeRel(&b, r, g.Edges[i], pal)
	}
	for _, e := range d.Elements {
		writeElement(&b, e, g.NodeByID(e.ID), o)
	}

	b.WriteString("  </g>\n</svg>\n")
	return []byte(b.String())
}

func writeElement(b *strings.Builder, e *Element, n *domain.Node, o RenderOptions) {
	if n == nil {
		return
	}
	x, y, w, h := n.Pos.X, n.Pos.Y, n.Size.W, n.Size.H
	fill := kindFill(e.Kind)
	rx := 4.0
	if strings.HasPrefix(e.Kind, "Person") {
		rx = 16
	}
	fmt.Fprintf(b, `    <rect x="%s" y="%s" width="%s" height="%s" rx="%s" fill="%s" stroke="%s"/>`,
		svgutil.Num(x), svgutil.Num(y), svgutil.Num(w), svgutil.Num(h), svgutil.Num(rx), fill, fill)
	b.WriteByte('\n')
	fmt.Fprintf(b, `    <text x="%s" y="%s" fill="#ffffff" text-anchor="middle" font-weight="bold">%s</text>`,
		svgutil.Num(x+w/2), svgutil.Num(y+o.FontSize+4), svgutil.Esc(e.Label))
	b.WriteByte('\n')
	fmt.Fprintf(b, `    <text x="%s" y="%s" fill="#e6e6e6" text-anchor="middle" font-size="%s">[%s]</text>`,
		svgutil.Num(x+w/2), svgutil.Num(y+o.FontSize*2), svgutil.Num(o.FontSize*0.8), svgutil.Esc(kindShort(e.Kind)))
	b.WriteByte('\n')
	if e.Descr != "" {
		for j, ln := range svgutil.SplitLines(wrapText(e.Descr, 26)) {
			fmt.Fprintf(b, `    <text x="%s" y="%s" fill="#f0f0f0" text-anchor="middle" font-size="%s">%s</text>`,
				svgutil.Num(x+w/2), svgutil.Num(y+o.FontSize*3+float64(j)*(o.FontSize)), svgutil.Num(o.FontSize*0.8), svgutil.Esc(ln))
			b.WriteByte('\n')
		}
	}
}

func writeRel(b *strings.Builder, r *Rel, e *domain.Edge, pal theme.Palette) {
	if len(e.Points) < 2 {
		return
	}
	var d strings.Builder
	for i, p := range e.Points {
		cmd := "L"
		if i == 0 {
			cmd = "M"
		}
		fmt.Fprintf(&d, "%s%s,%s ", cmd, svgutil.Num(p.X), svgutil.Num(p.Y))
	}
	fmt.Fprintf(b, `    <path d="%s" fill="none" stroke="%s" stroke-dasharray="4,3" marker-end="url(#c4-arrow)"/>`,
		strings.TrimSpace(d.String()), pal.Edge)
	b.WriteByte('\n')
	if r.Label != "" {
		lp0, lpn := e.Points[0], e.Points[len(e.Points)-1]
		mid := domain.Point{X: (lp0.X + lpn.X) / 2, Y: (lp0.Y + lpn.Y) / 2}
		fmt.Fprintf(b, `    <text x="%s" y="%s" fill="%s" text-anchor="middle" dy="-2">%s</text>`,
			svgutil.Num(mid.X), svgutil.Num(mid.Y), pal.Text, svgutil.Esc(r.Label))
		b.WriteByte('\n')
	}
}

func elementSize(e *Element, fontSize float64) domain.Size {
	maxW := svgutil.TextWidth(e.Label, fontSize)
	if l := len([]rune(e.Descr)); l < 28 {
		if wd := svgutil.TextWidth(e.Descr, fontSize); wd > maxW {
			maxW = wd
		}
	}
	w := maxW + 24
	if w < 120 {
		w = 120
	}
	if w > 200 {
		w = 200
	}
	h := fontSize*3 + 8
	if e.Descr != "" {
		lines := len(svgutil.SplitLines(wrapText(e.Descr, 26)))
		h += float64(lines) * fontSize
	}
	return domain.Size{W: w, H: h}
}

// kindShort returns a short label for the element kind tag.
func kindShort(kind string) string {
	switch {
	case strings.HasPrefix(kind, "Person"):
		return "Person"
	case strings.HasPrefix(kind, "System"):
		return "System"
	case strings.HasPrefix(kind, "Container"):
		return "Container"
	case strings.HasPrefix(kind, "Component"):
		return "Component"
	}
	return kind
}

// wrapText inserts <br> roughly every width runes at word boundaries.
func wrapText(s string, width int) string {
	words := strings.Fields(s)
	var lines []string
	var cur strings.Builder
	for _, w := range words {
		if cur.Len() > 0 && cur.Len()+1+len(w) > width {
			lines = append(lines, cur.String())
			cur.Reset()
		}
		if cur.Len() > 0 {
			cur.WriteByte(' ')
		}
		cur.WriteString(w)
	}
	if cur.Len() > 0 {
		lines = append(lines, cur.String())
	}
	return strings.Join(lines, "<br>")
}
