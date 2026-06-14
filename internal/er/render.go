package er

import (
	"fmt"
	"math"
	"strings"

	"github.com/zkrebbekx/go-mermaid/internal/domain"
	"github.com/zkrebbekx/go-mermaid/internal/layout"
	"github.com/zkrebbekx/go-mermaid/internal/svgutil"
	"github.com/zkrebbekx/go-mermaid/internal/theme"
)

// RenderOptions controls ER diagram appearance.
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

// Render parses and renders ER diagram source to SVG.
func Render(src string, o RenderOptions) ([]byte, error) {
	d, err := Parse(src)
	if err != nil {
		return nil, err
	}

	g := &domain.Graph{Direction: domain.TopBottom}
	for _, e := range d.Entities {
		n := &domain.Node{ID: e.Name, Label: e.Name, Shape: domain.ShapeRect}
		n.Size = entitySize(e, o.FontSize)
		g.Nodes = append(g.Nodes, n)
	}
	for _, r := range d.Relationships {
		g.Edges = append(g.Edges, &domain.Edge{From: r.From, To: r.To, Label: r.Label})
	}

	res, err := layout.Compute(g, layout.Options{NodeSep: 55, RankSep: 95, FontSize: o.FontSize})
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

	for i, r := range d.Relationships {
		writeRelationship(&b, r, g.Edges[i], pal)
	}
	for _, e := range d.Entities {
		writeEntity(&b, e, g.NodeByID(e.Name), pal, o)
	}

	b.WriteString("  </g>\n</svg>\n")
	return []byte(b.String())
}

func writeEntity(b *strings.Builder, e *Entity, n *domain.Node, pal theme.Palette, o RenderOptions) {
	if n == nil {
		return
	}
	x, y, w := n.Pos.X, n.Pos.Y, n.Size.W
	header := o.FontSize + 10
	row := o.FontSize + rowPad

	fmt.Fprintf(b, `    <rect x="%s" y="%s" width="%s" height="%s" fill="%s" stroke="%s"/>`,
		svgutil.Num(x), svgutil.Num(y), svgutil.Num(w), svgutil.Num(n.Size.H), pal.NodeFill, pal.NodeStroke)
	b.WriteByte('\n')
	fmt.Fprintf(b, `    <text x="%s" y="%s" fill="%s" text-anchor="middle" font-weight="bold">%s</text>`,
		svgutil.Num(x+w/2), svgutil.Num(y+o.FontSize), pal.Text, svgutil.Esc(e.Name))
	b.WriteByte('\n')
	if len(e.Attributes) > 0 {
		fmt.Fprintf(b, `    <line x1="%s" y1="%s" x2="%s" y2="%s" stroke="%s"/>`,
			svgutil.Num(x), svgutil.Num(y+header), svgutil.Num(x+w), svgutil.Num(y+header), pal.NodeStroke)
		b.WriteByte('\n')
	}
	cy := y + header
	for _, a := range e.Attributes {
		cy += row
		fmt.Fprintf(b, `    <text x="%s" y="%s" fill="%s">%s</text>`,
			svgutil.Num(x+boxPadX), svgutil.Num(cy-rowPad/2), pal.Text, svgutil.Esc(a))
		b.WriteByte('\n')
	}
}

func writeRelationship(b *strings.Builder, r *Relationship, e *domain.Edge, pal theme.Palette) {
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

	last := len(e.Points) - 1
	writeCrow(b, r.LeftKind, e.Points[0], e.Points[1], pal)
	writeCrow(b, r.RightKind, e.Points[last], e.Points[last-1], pal)

	if r.Label != "" {
		lp0, lpn := e.Points[0], e.Points[len(e.Points)-1]
		mid := domain.Point{X: (lp0.X + lpn.X) / 2, Y: (lp0.Y + lpn.Y) / 2}
		fmt.Fprintf(b, `    <text x="%s" y="%s" fill="%s" text-anchor="middle" dy="-2">%s</text>`,
			svgutil.Num(mid.X), svgutil.Num(mid.Y), pal.Text, svgutil.Esc(r.Label))
		b.WriteByte('\n')
	}
}

// writeCrow draws a crow's-foot cardinality marker at tip, where the line
// continues toward next. The marker faces into the entity at tip.
func writeCrow(b *strings.Builder, kind Card, tip, next domain.Point, pal theme.Palette) {
	dx, dy := tip.X-next.X, tip.Y-next.Y
	d := math.Hypot(dx, dy)
	if d == 0 {
		return
	}
	dx, dy = dx/d, dy/d // unit vector pointing into the entity
	px, py := -dy, dx   // perpendicular

	line := func(x1, y1, x2, y2 float64) {
		fmt.Fprintf(b, `    <line x1="%s" y1="%s" x2="%s" y2="%s" stroke="%s"/>`,
			svgutil.Num(x1), svgutil.Num(y1), svgutil.Num(x2), svgutil.Num(y2), pal.Edge)
		b.WriteByte('\n')
	}
	bar := func(dist, half float64) {
		bx, by := tip.X-dx*dist, tip.Y-dy*dist
		line(bx+px*half, by+py*half, bx-px*half, by-py*half)
	}
	circle := func(dist, r float64) {
		fmt.Fprintf(b, `    <circle cx="%s" cy="%s" r="%s" fill="%s" stroke="%s"/>`,
			svgutil.Num(tip.X-dx*dist), svgutil.Num(tip.Y-dy*dist), svgutil.Num(r), pal.Background, pal.Edge)
		b.WriteByte('\n')
	}
	foot := func() {
		const depth, spread = 12.0, 6.0
		bx, by := tip.X-dx*depth, tip.Y-dy*depth
		line(bx, by, tip.X+px*spread, tip.Y+py*spread)
		line(bx, by, tip.X-px*spread, tip.Y-py*spread)
		line(bx, by, tip.X, tip.Y)
	}

	switch kind {
	case CardOne:
		bar(7, 5)
	case CardZeroOne:
		bar(8, 5)
		circle(15, 4)
	case CardMany:
		foot()
	case CardOneMany:
		foot()
		bar(14, 5)
	case CardZeroMany:
		foot()
		circle(18, 4)
	}
}

func entitySize(e *Entity, fontSize float64) domain.Size {
	maxW := svgutil.TextWidth(e.Name, fontSize)
	for _, a := range e.Attributes {
		if wd := svgutil.TextWidth(a, fontSize); wd > maxW {
			maxW = wd
		}
	}
	w := maxW + boxPadX*2
	if w < 90 {
		w = 90
	}
	h := fontSize + 10
	if len(e.Attributes) > 0 {
		h += float64(len(e.Attributes)) * (fontSize + rowPad)
	}
	return domain.Size{W: w, H: h}
}
