package state

import (
	"fmt"
	"strings"

	"github.com/zkrebbekx/go-mermaid/internal/domain"
	"github.com/zkrebbekx/go-mermaid/internal/layout"
	"github.com/zkrebbekx/go-mermaid/internal/svgutil"
	"github.com/zkrebbekx/go-mermaid/internal/theme"
)

// RenderOptions controls state diagram appearance.
type RenderOptions struct {
	Theme    string
	FontFace string
	FontSize float64
	Padding  float64
	Title    string
}

const pseudoSize = 18.0

// Render parses and renders state diagram source to SVG.
func Render(src string, o RenderOptions) ([]byte, error) {
	d, err := Parse(src)
	if err != nil {
		return nil, err
	}

	g := &domain.Graph{Direction: domain.TopBottom}
	for _, s := range d.States {
		n := &domain.Node{ID: s.ID, Label: s.Label, Shape: domain.ShapeRound}
		if s.Start || s.End {
			n.Size = domain.Size{W: pseudoSize, H: pseudoSize}
		} else {
			n.Size = stateSize(s, o.FontSize)
		}
		g.Nodes = append(g.Nodes, n)
	}
	for _, t := range d.Transitions {
		g.Edges = append(g.Edges, &domain.Edge{From: t.From, To: t.To, Label: t.Label})
	}

	res, err := layout.Compute(g, layout.Options{NodeSep: 45, RankSep: 85, FontSize: o.FontSize})
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
	fmt.Fprintf(&b, `  <defs><marker id="st-arrow" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="7" markerHeight="7" orient="auto-start-reverse"><path d="M0,0 L10,5 L0,10 z" fill="%s"/></marker></defs>`, pal.Edge)
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

	for _, t := range d.Transitions {
		writeTransition(&b, t, g, pal)
	}
	for _, s := range d.States {
		writeState(&b, s, g.NodeByID(s.ID), pal, o)
	}

	b.WriteString("  </g>\n</svg>\n")
	return []byte(b.String())
}

func writeState(b *strings.Builder, s *State, n *domain.Node, pal theme.Palette, o RenderOptions) {
	if n == nil {
		return
	}
	c := n.Center()
	switch {
	case s.Start:
		fmt.Fprintf(b, `    <circle cx="%s" cy="%s" r="7" fill="%s"/>`,
			svgutil.Num(c.X), svgutil.Num(c.Y), pal.Edge)
		b.WriteByte('\n')
	case s.End:
		fmt.Fprintf(b, `    <circle cx="%s" cy="%s" r="8" fill="%s" stroke="%s"/>`,
			svgutil.Num(c.X), svgutil.Num(c.Y), pal.Background, pal.Edge)
		fmt.Fprintf(b, `<circle cx="%s" cy="%s" r="4" fill="%s"/>`,
			svgutil.Num(c.X), svgutil.Num(c.Y), pal.Edge)
		b.WriteByte('\n')
	default:
		fmt.Fprintf(b, `    <rect x="%s" y="%s" width="%s" height="%s" rx="8" fill="%s" stroke="%s"/>`,
			svgutil.Num(n.Pos.X), svgutil.Num(n.Pos.Y), svgutil.Num(n.Size.W), svgutil.Num(n.Size.H), pal.NodeFill, pal.NodeStroke)
		b.WriteByte('\n')
		fmt.Fprintf(b, `    <text x="%s" y="%s" fill="%s" text-anchor="middle">%s</text>`,
			svgutil.Num(c.X), svgutil.Num(c.Y+o.FontSize*0.35), pal.Text, svgutil.Esc(s.Label))
		b.WriteByte('\n')
	}
}

func writeTransition(b *strings.Builder, t *Transition, g *domain.Graph, pal theme.Palette) {
	var e *domain.Edge
	for _, ed := range g.Edges {
		if ed.From == t.From && ed.To == t.To && ed.Label == t.Label {
			e = ed
			break
		}
	}
	if e == nil || len(e.Points) < 2 {
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
	fmt.Fprintf(b, `    <path d="%s" fill="none" stroke="%s" marker-end="url(#st-arrow)"/>`,
		strings.TrimSpace(d.String()), pal.Edge)
	b.WriteByte('\n')
	if t.Label != "" {
		lp0, lpn := e.Points[0], e.Points[len(e.Points)-1]
		mid := domain.Point{X: (lp0.X + lpn.X) / 2, Y: (lp0.Y + lpn.Y) / 2}
		fmt.Fprintf(b, `    <text x="%s" y="%s" fill="%s" text-anchor="middle" dy="-2">%s</text>`,
			svgutil.Num(mid.X), svgutil.Num(mid.Y), pal.Text, svgutil.Esc(t.Label))
		b.WriteByte('\n')
	}
}

func stateSize(s *State, fontSize float64) domain.Size {
	w := svgutil.TextWidth(s.Label, fontSize) + 24
	if w < 50 {
		w = 50
	}
	return domain.Size{W: w, H: fontSize + 16}
}
