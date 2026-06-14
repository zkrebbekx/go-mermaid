package mindmap

import (
	"fmt"
	"strings"

	"github.com/zkrebbekx/go-mermaid/internal/svgutil"
	"github.com/zkrebbekx/go-mermaid/internal/theme"
)

// RenderOptions controls mindmap appearance.
type RenderOptions struct {
	Theme    string
	FontFace string
	FontSize float64
	Padding  float64
	Title    string
}

const (
	colW = 170.0
	rowH = 34.0
)

// Render parses and renders mindmap source to SVG.
func Render(src string, o RenderOptions) ([]byte, error) {
	d, err := Parse(src)
	if err != nil {
		return nil, err
	}
	return svg(d, o), nil
}

func svg(d *Diagram, o RenderOptions) []byte {
	pal := theme.For(o.Theme)
	pad := o.Padding
	titleH := svgutil.TitleHeight(o.Title, o.FontSize)

	var leaf float64
	maxDepth := 0
	var place func(n *Node)
	place = func(n *Node) {
		if n.Depth > maxDepth {
			maxDepth = n.Depth
		}
		n.X = pad + float64(n.Depth)*colW
		if len(n.Children) == 0 {
			n.Y = pad + titleH + leaf*rowH
			leaf++
			return
		}
		for _, c := range n.Children {
			place(c)
		}
		n.Y = (n.Children[0].Y + n.Children[len(n.Children)-1].Y) / 2
	}
	place(d.Root)

	w := pad*2 + float64(maxDepth+1)*colW
	h := pad*2 + titleH + leaf*rowH
	if leaf == 0 {
		h = pad*2 + titleH + rowH
	}

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

	writeEdges(&b, d.Root, o, pal)
	writeNodes(&b, d.Root, o, pal)

	b.WriteString("</svg>\n")
	return []byte(b.String())
}

func writeEdges(b *strings.Builder, n *Node, o RenderOptions, pal theme.Palette) {
	nw := nodeWidth(n, o.FontSize)
	for _, c := range n.Children {
		x1 := n.X + nw
		y1 := n.Y + rowH/2
		x2 := c.X
		y2 := c.Y + rowH/2
		mx := (x1 + x2) / 2
		fmt.Fprintf(b, `  <path d="M%s,%s C%s,%s %s,%s %s,%s" fill="none" stroke="%s"/>`,
			svgutil.Num(x1), svgutil.Num(y1), svgutil.Num(mx), svgutil.Num(y1),
			svgutil.Num(mx), svgutil.Num(y2), svgutil.Num(x2), svgutil.Num(y2), pal.Edge)
		b.WriteByte('\n')
		writeEdges(b, c, o, pal)
	}
}

func writeNodes(b *strings.Builder, n *Node, o RenderOptions, pal theme.Palette) {
	nw := nodeWidth(n, o.FontSize)
	fmt.Fprintf(b, `  <rect x="%s" y="%s" width="%s" height="%s" rx="%s" fill="%s" stroke="%s"/>`,
		svgutil.Num(n.X), svgutil.Num(n.Y), svgutil.Num(nw), svgutil.Num(rowH-8), svgutil.Num((rowH-8)/2), pal.NodeFill, pal.NodeStroke)
	b.WriteByte('\n')
	fmt.Fprintf(b, `  <text x="%s" y="%s" fill="%s" text-anchor="middle">%s</text>`,
		svgutil.Num(n.X+nw/2), svgutil.Num(n.Y+(rowH-8)/2+o.FontSize*0.35), pal.Text, svgutil.Esc(n.Text))
	b.WriteByte('\n')
	for _, c := range n.Children {
		writeNodes(b, c, o, pal)
	}
}

func nodeWidth(n *Node, fontSize float64) float64 {
	w := svgutil.TextWidth(n.Text, fontSize) + 20
	if w < 50 {
		w = 50
	}
	if w > colW-20 {
		w = colW - 20
	}
	return w
}
