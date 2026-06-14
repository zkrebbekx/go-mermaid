package class

import (
	"fmt"
	"math"
	"strings"

	"github.com/zkrebbekx/go-mermaid/internal/domain"
	"github.com/zkrebbekx/go-mermaid/internal/layout"
	"github.com/zkrebbekx/go-mermaid/internal/svgutil"
	"github.com/zkrebbekx/go-mermaid/internal/theme"
)

// RenderOptions controls class diagram appearance.
type RenderOptions struct {
	Theme    string
	FontFace string
	FontSize float64
	Padding  float64
	Title    string
}

const (
	boxPadX = 10.0
	rowPad  = 6.0
)

// Render parses and renders class diagram source to SVG.
func Render(src string, o RenderOptions) ([]byte, error) {
	d, err := Parse(src)
	if err != nil {
		return nil, err
	}

	g := &domain.Graph{Direction: domain.TopBottom}
	for _, c := range d.Classes {
		n := &domain.Node{ID: c.Name, Label: c.Name, Shape: domain.ShapeRect}
		n.Size = classSize(c, o.FontSize)
		g.Nodes = append(g.Nodes, n)
	}
	for _, r := range d.Relations {
		g.Edges = append(g.Edges, &domain.Edge{From: r.From, To: r.To, Label: r.Label})
	}

	res, err := layout.Compute(g, layout.Options{NodeSep: 50, RankSep: 90, FontSize: o.FontSize})
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
	fmt.Fprintf(&b, `  <rect width="100%%" height="100%%" fill="%s"/>`, pal.Background)
	b.WriteByte('\n')
	if o.Title != "" {
		fmt.Fprintf(&b, `  <text x="%s" y="%s" fill="%s" text-anchor="middle" font-weight="bold">%s</text>`,
			svgutil.Num(w/2), svgutil.Num(pad+o.FontSize), pal.Text, svgutil.Esc(o.Title))
		b.WriteByte('\n')
	}
	fmt.Fprintf(&b, `  <g transform="translate(%s,%s)">`, svgutil.Num(pad), svgutil.Num(pad+titleH))
	b.WriteByte('\n')

	for i, r := range d.Relations {
		writeRelation(&b, r, g.Edges[i], pal)
	}
	for _, c := range d.Classes {
		writeClass(&b, c, g.NodeByID(c.Name), pal, o)
	}

	b.WriteString("  </g>\n</svg>\n")
	return []byte(b.String())
}

func writeClass(b *strings.Builder, c *Class, n *domain.Node, pal theme.Palette, o RenderOptions) {
	if n == nil {
		return
	}
	x, y, w := n.Pos.X, n.Pos.Y, n.Size.W
	row := o.FontSize + rowPad
	header := o.FontSize + 10

	fmt.Fprintf(b, `    <rect x="%s" y="%s" width="%s" height="%s" fill="%s" stroke="%s"/>`,
		svgutil.Num(x), svgutil.Num(y), svgutil.Num(w), svgutil.Num(n.Size.H), pal.NodeFill, pal.NodeStroke)
	b.WriteByte('\n')
	fmt.Fprintf(b, `    <text x="%s" y="%s" fill="%s" text-anchor="middle" font-weight="bold">%s</text>`,
		svgutil.Num(x+w/2), svgutil.Num(y+o.FontSize), pal.Text, svgutil.Esc(c.Name))
	b.WriteByte('\n')

	cy := y + header
	writeDivider := func() {
		fmt.Fprintf(b, `    <line x1="%s" y1="%s" x2="%s" y2="%s" stroke="%s"/>`,
			svgutil.Num(x), svgutil.Num(cy), svgutil.Num(x+w), svgutil.Num(cy), pal.NodeStroke)
		b.WriteByte('\n')
	}
	writeRows := func(rows []string) {
		for _, m := range rows {
			cy += row
			fmt.Fprintf(b, `    <text x="%s" y="%s" fill="%s">%s</text>`,
				svgutil.Num(x+boxPadX), svgutil.Num(cy-rowPad/2), pal.Text, svgutil.Esc(m))
			b.WriteByte('\n')
		}
	}

	if len(c.Attributes) > 0 || len(c.Methods) > 0 {
		writeDivider()
	}
	writeRows(c.Attributes)
	if len(c.Methods) > 0 {
		if len(c.Attributes) > 0 {
			writeDivider()
		}
		writeRows(c.Methods)
	}
}

func writeRelation(b *strings.Builder, r *Relation, e *domain.Edge, pal theme.Palette) {
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
	dash := ""
	if r.Dashed {
		dash = ` stroke-dasharray="5,4"`
	}
	fmt.Fprintf(b, `    <path d="%s" fill="none" stroke="%s"%s/>`, strings.TrimSpace(d.String()), pal.Edge, dash)
	b.WriteByte('\n')

	p0, p1 := e.Points[0], e.Points[1]
	ldx, ldy := unit(p0, p1)
	writeHead(b, r.Left, p0, ldx, ldy, pal)
	pn, pm := e.Points[len(e.Points)-1], e.Points[len(e.Points)-2]
	rdx, rdy := unit(pn, pm)
	writeHead(b, r.Right, pn, rdx, rdy, pal)

	if r.Label != "" {
		lp0, lpn := e.Points[0], e.Points[len(e.Points)-1]
		mid := domain.Point{X: (lp0.X + lpn.X) / 2, Y: (lp0.Y + lpn.Y) / 2}
		fmt.Fprintf(b, `    <text x="%s" y="%s" fill="%s" text-anchor="middle" dy="-2">%s</text>`,
			svgutil.Num(mid.X), svgutil.Num(mid.Y), pal.Text, svgutil.Esc(r.Label))
		b.WriteByte('\n')
	}
}

// writeHead draws a relationship decoration at tip pointing in direction (dx,dy).
func writeHead(b *strings.Builder, kind headKind, tip domain.Point, dx, dy float64, pal theme.Palette) {
	if kind == headNone {
		return
	}
	const l, hw = 12.0, 6.0
	bx, by := tip.X+dx*l, tip.Y+dy*l // base, back along the line
	px, py := -dy, dx                // perpendicular
	switch kind {
	case headArrow:
		fmt.Fprintf(b, `    <path d="M%s,%s L%s,%s L%s,%s Z" fill="%s"/>`,
			svgutil.Num(tip.X), svgutil.Num(tip.Y),
			svgutil.Num(bx+px*hw), svgutil.Num(by+py*hw),
			svgutil.Num(bx-px*hw), svgutil.Num(by-py*hw), pal.Edge)
	case headTriangle:
		fmt.Fprintf(b, `    <path d="M%s,%s L%s,%s L%s,%s Z" fill="%s" stroke="%s"/>`,
			svgutil.Num(tip.X), svgutil.Num(tip.Y),
			svgutil.Num(bx+px*hw), svgutil.Num(by+py*hw),
			svgutil.Num(bx-px*hw), svgutil.Num(by-py*hw), pal.Background, pal.Edge)
	case headDiamondFilled, headDiamondHollow:
		mx, my := tip.X+dx*l/2, tip.Y+dy*l/2
		fill := pal.Edge
		if kind == headDiamondHollow {
			fill = pal.Background
		}
		fmt.Fprintf(b, `    <path d="M%s,%s L%s,%s L%s,%s L%s,%s Z" fill="%s" stroke="%s"/>`,
			svgutil.Num(tip.X), svgutil.Num(tip.Y),
			svgutil.Num(mx+px*hw), svgutil.Num(my+py*hw),
			svgutil.Num(bx), svgutil.Num(by),
			svgutil.Num(mx-px*hw), svgutil.Num(my-py*hw), fill, pal.Edge)
	}
	b.WriteByte('\n')
}

// unit returns the unit vector from a toward b (zero if coincident).
func unit(a, b domain.Point) (float64, float64) {
	dx, dy := b.X-a.X, b.Y-a.Y
	d := math.Hypot(dx, dy)
	if d == 0 {
		return 0, 0
	}
	return dx / d, dy / d
}

// classSize computes a box size that fits the name and all members.
func classSize(c *Class, fontSize float64) domain.Size {
	maxW := svgutil.TextWidth(c.Name, fontSize)
	for _, m := range append(append([]string{}, c.Attributes...), c.Methods...) {
		if wd := svgutil.TextWidth(m, fontSize); wd > maxW {
			maxW = wd
		}
	}
	w := maxW + boxPadX*2
	if w < 80 {
		w = 80
	}
	h := fontSize + 10 // header
	rows := len(c.Attributes) + len(c.Methods)
	if rows > 0 {
		h += float64(rows) * (fontSize + rowPad)
	}
	return domain.Size{W: w, H: h}
}
