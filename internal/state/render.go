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

	face := svgutil.FaceFor(o.FontFace)
	g := &domain.Graph{Direction: directionOf(d.Direction)}
	for _, s := range d.States {
		// A composite state is drawn as a box around its members, not as a
		// node of its own, so it is kept out of the layout graph.
		if d.composite(s.ID) != nil {
			continue
		}
		n := &domain.Node{ID: s.ID, Label: s.Label, Shape: domain.ShapeRound}
		switch {
		case s.Start || s.End:
			n.Size = domain.Size{W: pseudoSize, H: pseudoSize}
		case s.Kind == KindFork || s.Kind == KindJoin:
			n.Size = domain.Size{W: forkW, H: forkH}
		case s.Kind == KindChoice:
			n.Size = domain.Size{W: choiceSize, H: choiceSize}
			n.Shape = domain.ShapeDiamond
		default:
			n.Size = stateSize(s, face, o.FontSize)
		}
		g.Nodes = append(g.Nodes, n)
	}
	for _, t := range d.Transitions {
		from, to := d.anchor(t.From, false), d.anchor(t.To, true)
		if from == "" || to == "" || from == to {
			continue
		}
		g.Edges = append(g.Edges, &domain.Edge{From: from, To: to, Label: t.Label})
	}

	res, err := layout.Compute(g, layout.Options{NodeSep: 45, RankSep: 85, FontSize: o.FontSize, FontFace: o.FontFace})
	if err != nil {
		return nil, err
	}
	return svg(d, g, res, o), nil
}

const (
	forkW      = 60.0 // width of a <<fork>> / <<join>> bar
	forkH      = 8.0  // thickness of a <<fork>> / <<join>> bar
	choiceSize = 36.0 // width and height of a <<choice>> diamond
	clusterPad = 14.0 // gap between a composite box and its members
	noteGap    = 20.0 // gap between a state and its note
)

// directionOf maps a `direction` line onto a layout direction, defaulting to
// top-to-bottom when the source does not ask for one.
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

// anchor maps a transition endpoint onto a node that exists in the layout
// graph. A composite state has no node of its own, so an arrow into it lands
// on its entry pseudostate and an arrow out of it leaves from its exit.
func (d *Diagram) anchor(id string, asTarget bool) string {
	c := d.composite(id)
	if c == nil {
		return id
	}
	if p := d.state(pseudoID(id, !asTarget)); p != nil {
		return p.ID
	}
	for _, m := range c.Members {
		if d.composite(m) == nil {
			return m
		}
	}
	return ""
}

// clusterBox returns the box enclosing a composite state's members, with room
// for its title. ok is false when no member was placed.
func clusterBox(c *Composite, d *Diagram, g *domain.Graph, fontSize float64) (x, y, w, h float64, ok bool) {
	var bd svgutil.Bounds
	for _, id := range c.Members {
		if inner := d.composite(id); inner != nil {
			if ix, iy, iw, ih, iok := clusterBox(inner, d, g, fontSize); iok {
				bd.AddRect(ix, iy, iw, ih)
			}
			continue
		}
		n := g.NodeByID(id)
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

// noteBox returns the box of a note placed beside its target state.
func noteBox(n *Note, g *domain.Graph, face svgutil.Face, fontSize float64) (x, y, w, h float64, ok bool) {
	target := g.NodeByID(n.Target)
	if target == nil {
		return 0, 0, 0, 0, false
	}
	w = face.Width(n.Text, fontSize) + 20
	h = fontSize + 12
	y = target.Center().Y - h/2
	if n.Side == SideLeft {
		x = target.Pos.X - noteGap - w
	} else {
		x = target.Pos.X + target.Size.W + noteGap
	}
	return x, y, w, h, true
}

func svg(d *Diagram, g *domain.Graph, res *layout.Result, o RenderOptions) []byte {
	pal := theme.For(o.Theme)
	pad := o.Padding
	face := svgutil.FaceFor(o.FontFace)
	titleH := svgutil.TitleHeight(o.Title, o.FontSize)

	// Composite boxes and notes sit outside the node extents the layout
	// reported, and a note placed left of a state reaches past the origin.
	// Collect everything so the canvas covers it.
	var bd svgutil.Bounds
	bd.AddRect(0, 0, res.Width, res.Height)
	for _, c := range d.Composites {
		if cx, cy, cw, ch, ok := clusterBox(c, d, g, o.FontSize); ok {
			bd.AddRect(cx, cy, cw, ch)
		}
	}
	for _, n := range d.Notes {
		if nx, ny, nw, nh, ok := noteBox(n, g, face, o.FontSize); ok {
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
	fmt.Fprintf(&b, `  <defs><marker id="st-arrow" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="7" markerHeight="7" orient="auto-start-reverse"><path d="M0,0 L10,5 L0,10 z" fill="%s"/></marker></defs>`, pal.Edge)
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

	for _, c := range d.Composites {
		writeCluster(&b, c, d, g, pal, o)
	}
	for _, t := range d.Transitions {
		writeTransition(&b, t, d, g, pal)
	}
	for _, s := range d.States {
		writeState(&b, s, g.NodeByID(s.ID), pal, o)
	}
	for _, n := range d.Notes {
		writeNote(&b, n, g, pal, face, o)
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
	case s.Kind == KindFork || s.Kind == KindJoin:
		fmt.Fprintf(b, `    <rect x="%s" y="%s" width="%s" height="%s" rx="2" fill="%s"/>`,
			svgutil.Num(n.Pos.X), svgutil.Num(n.Pos.Y), svgutil.Num(n.Size.W), svgutil.Num(n.Size.H), pal.Edge)
		b.WriteByte('\n')
	case s.Kind == KindChoice:
		x, y, w, h := n.Pos.X, n.Pos.Y, n.Size.W, n.Size.H
		fmt.Fprintf(b, `    <polygon points="%s,%s %s,%s %s,%s %s,%s" fill="%s" stroke="%s"/>`,
			svgutil.Num(x+w/2), svgutil.Num(y), svgutil.Num(x+w), svgutil.Num(y+h/2),
			svgutil.Num(x+w/2), svgutil.Num(y+h), svgutil.Num(x), svgutil.Num(y+h/2),
			pal.NodeFill, pal.NodeStroke)
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

// writeCluster draws the dashed box and title of a composite state.
func writeCluster(b *strings.Builder, c *Composite, d *Diagram, g *domain.Graph, pal theme.Palette, o RenderOptions) {
	x, y, w, h, ok := clusterBox(c, d, g, o.FontSize)
	if !ok {
		return
	}
	fmt.Fprintf(b, `    <rect x="%s" y="%s" width="%s" height="%s" fill="none" stroke="%s" stroke-dasharray="4,3" rx="6"/>`,
		svgutil.Num(x), svgutil.Num(y), svgutil.Num(w), svgutil.Num(h), pal.NodeStroke)
	b.WriteByte('\n')
	fmt.Fprintf(b, `    <text x="%s" y="%s" fill="%s">%s</text>`,
		svgutil.Num(x+6), svgutil.Num(y+o.FontSize), pal.Text, svgutil.Esc(c.Label))
	b.WriteByte('\n')
}

// writeNote draws a note box beside the state it annotates.
func writeNote(b *strings.Builder, n *Note, g *domain.Graph, pal theme.Palette, face svgutil.Face, o RenderOptions) {
	x, y, w, h, ok := noteBox(n, g, face, o.FontSize)
	if !ok {
		return
	}
	fmt.Fprintf(b, `    <rect x="%s" y="%s" width="%s" height="%s" fill="#FFF5AD" stroke="%s"/>`,
		svgutil.Num(x), svgutil.Num(y), svgutil.Num(w), svgutil.Num(h), pal.NodeStroke)
	b.WriteByte('\n')
	fmt.Fprintf(b, `    <text x="%s" y="%s" fill="%s" text-anchor="middle">%s</text>`,
		svgutil.Num(x+w/2), svgutil.Num(y+h/2+o.FontSize*0.35), pal.Text, svgutil.Esc(n.Text))
	b.WriteByte('\n')
}

func writeTransition(b *strings.Builder, t *Transition, dia *Diagram, g *domain.Graph, pal theme.Palette) {
	from, to := dia.anchor(t.From, false), dia.anchor(t.To, true)
	var e *domain.Edge
	for _, ed := range g.Edges {
		if ed.From == from && ed.To == to && ed.Label == t.Label {
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
		mid := e.LabelPos
		fmt.Fprintf(b, `    <text x="%s" y="%s" fill="%s" text-anchor="middle" dy="-2">%s</text>`,
			svgutil.Num(mid.X), svgutil.Num(mid.Y), pal.Text, svgutil.Esc(t.Label))
		b.WriteByte('\n')
	}
}

func stateSize(s *State, face svgutil.Face, fontSize float64) domain.Size {
	w := face.Width(s.Label, fontSize) + 24
	if w < 50 {
		w = 50
	}
	return domain.Size{W: w, H: fontSize + 16}
}
