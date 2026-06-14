package requirement

import (
	"fmt"
	"strings"

	"github.com/zkrebbekx/go-mermaid/internal/domain"
	"github.com/zkrebbekx/go-mermaid/internal/layout"
	"github.com/zkrebbekx/go-mermaid/internal/svgutil"
	"github.com/zkrebbekx/go-mermaid/internal/theme"
)

// RenderOptions controls requirement diagram appearance.
type RenderOptions struct {
	Theme    string
	FontFace string
	FontSize float64
	Padding  float64
	Title    string
}

// Render parses and renders requirement diagram source to SVG.
func Render(src string, o RenderOptions) ([]byte, error) {
	d, err := Parse(src)
	if err != nil {
		return nil, err
	}

	g := &domain.Graph{Direction: domain.TopBottom}
	for _, n := range d.Nodes {
		node := &domain.Node{ID: n.ID, Label: n.ID, Shape: domain.ShapeRect}
		node.Size = nodeSize(n, o.FontSize)
		g.Nodes = append(g.Nodes, node)
	}
	for _, r := range d.Rels {
		g.Edges = append(g.Edges, &domain.Edge{From: r.From, To: r.To, Label: r.Type})
	}
	res, err := layout.Compute(g, layout.Options{NodeSep: 55, RankSep: 100, FontSize: o.FontSize})
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
	fmt.Fprintf(&b, `  <defs><marker id="req-arrow" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="7" markerHeight="7" orient="auto-start-reverse"><path d="M0,0 L10,5 L0,10 z" fill="%s"/></marker></defs>`, pal.Edge)
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
	for _, n := range d.Nodes {
		writeNode(&b, n, g.NodeByID(n.ID), pal, o)
	}

	b.WriteString("  </g>\n</svg>\n")
	return []byte(b.String())
}

// rows returns the display rows (name tag plus selected fields) for a node.
func rows(n *Node) []string {
	out := []string{"«" + tagFor(n) + "»", n.ID}
	if n.IsElement {
		if v := n.Fields["type"]; v != "" {
			out = append(out, "type: "+v)
		}
	} else {
		if v := n.Fields["id"]; v != "" {
			out = append(out, "id: "+v)
		}
		if v := n.Fields["risk"]; v != "" {
			out = append(out, "risk: "+v)
		}
	}
	return out
}

func tagFor(n *Node) string {
	if n.IsElement {
		return "element"
	}
	if n.Kind != "" {
		return n.Kind
	}
	return "requirement"
}

func writeNode(b *strings.Builder, n *Node, dn *domain.Node, pal theme.Palette, o RenderOptions) {
	if dn == nil {
		return
	}
	x, y, w, h := dn.Pos.X, dn.Pos.Y, dn.Size.W, dn.Size.H
	fmt.Fprintf(b, `    <rect x="%s" y="%s" width="%s" height="%s" rx="3" fill="%s" stroke="%s"/>`,
		svgutil.Num(x), svgutil.Num(y), svgutil.Num(w), svgutil.Num(h), pal.NodeFill, pal.NodeStroke)
	b.WriteByte('\n')
	cy := y + o.FontSize + 2
	for i, r := range rows(n) {
		weight := ""
		if i == 1 {
			weight = ` font-weight="bold"`
		}
		fmt.Fprintf(b, `    <text x="%s" y="%s" fill="%s" text-anchor="middle"%s>%s</text>`,
			svgutil.Num(x+w/2), svgutil.Num(cy), pal.Text, weight, svgutil.Esc(r))
		b.WriteByte('\n')
		cy += o.FontSize + 2
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
	fmt.Fprintf(b, `    <path d="%s" fill="none" stroke="%s" stroke-dasharray="4,3" marker-end="url(#req-arrow)"/>`,
		strings.TrimSpace(d.String()), pal.Edge)
	b.WriteByte('\n')
	if r.Type != "" {
		lp0, lpn := e.Points[0], e.Points[len(e.Points)-1]
		mid := domain.Point{X: (lp0.X + lpn.X) / 2, Y: (lp0.Y + lpn.Y) / 2}
		fmt.Fprintf(b, `    <text x="%s" y="%s" fill="%s" text-anchor="middle" dy="-2">%s</text>`,
			svgutil.Num(mid.X), svgutil.Num(mid.Y), pal.Text, svgutil.Esc("«"+r.Type+"»"))
		b.WriteByte('\n')
	}
}

func nodeSize(n *Node, fontSize float64) domain.Size {
	maxW := 0.0
	rs := rows(n)
	for _, r := range rs {
		if wd := svgutil.TextWidth(r, fontSize); wd > maxW {
			maxW = wd
		}
	}
	w := maxW + 24
	if w < 110 {
		w = 110
	}
	h := float64(len(rs))*(fontSize+2) + 8
	return domain.Size{W: w, H: h}
}
