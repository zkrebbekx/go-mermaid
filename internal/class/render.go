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

	g := &domain.Graph{Direction: directionOf(d.Direction)}
	for _, c := range d.Classes {
		n := &domain.Node{ID: c.Name, Label: c.Name, Shape: domain.ShapeRect}
		n.Size = classSize(c, svgutil.FaceFor(o.FontFace), o.FontSize)
		g.Nodes = append(g.Nodes, n)
	}
	for _, r := range d.Relations {
		g.Edges = append(g.Edges, &domain.Edge{From: r.From, To: r.To, Label: r.Label})
	}

	res, err := layout.Compute(g, layout.Options{NodeSep: 50, RankSep: 90, FontSize: o.FontSize, FontFace: o.FontFace})
	if err != nil {
		return nil, err
	}
	return svg(d, g, res, o), nil
}

const (
	clusterPad = 14.0 // gap between a namespace box and its classes
	cardGap    = 6.0  // gap between a relationship end and its multiplicity
)

// directionOf maps a `direction` line onto a layout direction.
func directionOf(dir string) domain.Direction {
	switch dir {
	case "LR":
		return domain.LeftRight
	case "RL":
		return domain.RightLeft
	case "BT":
		return domain.BottomTop
	default:
		return domain.TopBottom
	}
}

// namespaceBox returns the box enclosing a namespace's classes, with room for
// its title. ok is false when no member was placed.
func namespaceBox(ns *Namespace, g *domain.Graph, fontSize float64) (x, y, w, h float64, ok bool) {
	var bd svgutil.Bounds
	for _, name := range ns.Members {
		n := g.NodeByID(name)
		if n == nil {
			continue
		}
		bd.AddRect(n.Pos.X, n.Pos.Y, n.Size.W, n.Size.H)
	}
	if bd.Empty() {
		return 0, 0, 0, 0, false
	}
	titleH := fontSize + 6
	return bd.MinX - clusterPad, bd.MinY - clusterPad - titleH,
		bd.MaxX - bd.MinX + clusterPad*2, bd.MaxY - bd.MinY + clusterPad*2 + titleH, true
}

func svg(d *Diagram, g *domain.Graph, res *layout.Result, o RenderOptions) []byte {
	pal := theme.For(o.Theme)
	pad := o.Padding
	titleH := svgutil.TitleHeight(o.Title, o.FontSize)

	// Namespace boxes reach outside the class extents the layout reported.
	var bd svgutil.Bounds
	bd.AddRect(0, 0, res.Width, res.Height)
	for _, ns := range d.Namespaces {
		if nx, ny, nw, nh, ok := namespaceBox(ns, g, o.FontSize); ok {
			bd.AddRect(nx, ny, nw, nh)
		}
	}
	shiftX, shiftY := bd.Offset()
	contentW, contentH := bd.Size()
	w := contentW + pad*2
	h := contentH + titleH + pad*2

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
	fmt.Fprintf(&b, `  <g transform="translate(%s,%s)">`, svgutil.Num(pad+shiftX), svgutil.Num(pad+titleH+shiftY))
	b.WriteByte('\n')

	for _, ns := range d.Namespaces {
		writeNamespace(&b, ns, g, pal, o)
	}
	for i, r := range d.Relations {
		writeRelation(&b, r, g.Edges[i], pal, o)
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
	nameY := y + o.FontSize
	if c.Annotation != "" {
		fmt.Fprintf(b, `    <text x="%s" y="%s" fill="%s" text-anchor="middle">%s</text>`,
			svgutil.Num(x+w/2), svgutil.Num(nameY), pal.Text, svgutil.Esc("«"+c.Annotation+"»"))
		b.WriteByte('\n')
		nameY += o.FontSize + 2
		header += o.FontSize + 2
	}
	fmt.Fprintf(b, `    <text x="%s" y="%s" fill="%s" text-anchor="middle" font-weight="bold">%s</text>`,
		svgutil.Num(x+w/2), svgutil.Num(nameY), pal.Text, svgutil.Esc(c.Label()))
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

func writeRelation(b *strings.Builder, r *Relation, e *domain.Edge, pal theme.Palette, o RenderOptions) {
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

	writeCardinality(b, r.LeftCard, e.Points[0], e.Points[1], pal, o)
	last := len(e.Points) - 1
	writeCardinality(b, r.RightCard, e.Points[last], e.Points[last-1], pal, o)

	if r.Label != "" {
		mid := e.LabelPos
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
func classSize(c *Class, face svgutil.Face, fontSize float64) domain.Size {
	maxW := face.Width(c.Label(), fontSize)
	if c.Annotation != "" {
		if wd := face.Width("«"+c.Annotation+"»", fontSize); wd > maxW {
			maxW = wd
		}
	}
	for _, m := range append(append([]string{}, c.Attributes...), c.Methods...) {
		if wd := face.Width(m, fontSize); wd > maxW {
			maxW = wd
		}
	}
	w := maxW + boxPadX*2
	if w < 80 {
		w = 80
	}
	h := fontSize + 10 // header
	if c.Annotation != "" {
		h += fontSize + 2
	}
	rows := len(c.Attributes) + len(c.Methods)
	if rows > 0 {
		h += float64(rows) * (fontSize + rowPad)
	}
	return domain.Size{W: w, H: h}
}

// writeCardinality draws a multiplicity label just inside the end of a
// relationship line. tip is the end point and next is the neighbouring
// waypoint, so the label sits along the line rather than on top of the class.
func writeCardinality(b *strings.Builder, card string, tip, next domain.Point, pal theme.Palette, o RenderOptions) {
	if card == "" {
		return
	}
	dx, dy := unit(tip, next)
	// Step along the line past the end decoration, then offset perpendicular
	// so the line does not strike through the text.
	x := tip.X + dx*(cardGap+10) - dy*9
	y := tip.Y + dy*(cardGap+10) + dx*9 + o.FontSize*0.35
	fmt.Fprintf(b, `    <text x="%s" y="%s" fill="%s" text-anchor="middle" font-size="%s">%s</text>`,
		svgutil.Num(x), svgutil.Num(y), pal.Text, svgutil.Num(o.FontSize-2), svgutil.Esc(card))
	b.WriteByte('\n')
}

// writeNamespace draws the dashed box and title of a namespace block.
func writeNamespace(b *strings.Builder, ns *Namespace, g *domain.Graph, pal theme.Palette, o RenderOptions) {
	x, y, w, h, ok := namespaceBox(ns, g, o.FontSize)
	if !ok {
		return
	}
	fmt.Fprintf(b, `    <rect x="%s" y="%s" width="%s" height="%s" fill="none" stroke="%s" stroke-dasharray="4,3" rx="6"/>`,
		svgutil.Num(x), svgutil.Num(y), svgutil.Num(w), svgutil.Num(h), pal.NodeStroke)
	b.WriteByte('\n')
	fmt.Fprintf(b, `    <text x="%s" y="%s" fill="%s">%s</text>`,
		svgutil.Num(x+6), svgutil.Num(y+o.FontSize), pal.Text, svgutil.Esc(ns.Name))
	b.WriteByte('\n')
}
