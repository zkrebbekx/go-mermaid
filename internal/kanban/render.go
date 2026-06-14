package kanban

import (
	"fmt"
	"strings"

	"github.com/zkrebbekx/go-mermaid/internal/svgutil"
	"github.com/zkrebbekx/go-mermaid/internal/theme"
)

// RenderOptions controls kanban appearance.
type RenderOptions struct {
	Theme    string
	FontFace string
	FontSize float64
	Padding  float64
	Title    string
}

var columnColors = []string{"#5B8FF9", "#61DDAA", "#F6BD16", "#7262FD", "#F6903D", "#008685"}

const (
	colW    = 160.0
	colGap  = 14.0
	headerH = 30.0
	cardH   = 34.0
	cardGap = 8.0
)

// Render parses and renders kanban source to SVG.
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

	maxCards := 0
	for _, c := range d.Columns {
		if len(c.Cards) > maxCards {
			maxCards = len(c.Cards)
		}
	}

	top := pad + titleH
	w := pad*2 + float64(len(d.Columns))*(colW+colGap)
	h := top + headerH + 8 + float64(maxCards)*(cardH+cardGap) + pad

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

	for ci, col := range d.Columns {
		x := pad + float64(ci)*(colW+colGap)
		color := columnColors[ci%len(columnColors)]
		fmt.Fprintf(&b, `  <rect x="%s" y="%s" width="%s" height="%s" rx="4" fill="%s"/>`,
			svgutil.Num(x), svgutil.Num(top), svgutil.Num(colW), svgutil.Num(headerH), color)
		b.WriteByte('\n')
		fmt.Fprintf(&b, `  <text x="%s" y="%s" fill="#ffffff" text-anchor="middle" font-weight="bold">%s</text>`,
			svgutil.Num(x+colW/2), svgutil.Num(top+headerH*0.66), svgutil.Esc(col.Title))
		b.WriteByte('\n')
		cy := top + headerH + 8
		for _, card := range col.Cards {
			fmt.Fprintf(&b, `  <rect x="%s" y="%s" width="%s" height="%s" rx="4" fill="%s" stroke="%s"/>`,
				svgutil.Num(x+6), svgutil.Num(cy), svgutil.Num(colW-12), svgutil.Num(cardH), pal.NodeFill, pal.NodeStroke)
			b.WriteByte('\n')
			fmt.Fprintf(&b, `  <text x="%s" y="%s" fill="%s" text-anchor="middle">%s</text>`,
				svgutil.Num(x+colW/2), svgutil.Num(cy+cardH/2+o.FontSize*0.35), pal.Text, svgutil.Esc(card.Text))
			b.WriteByte('\n')
			cy += cardH + cardGap
		}
	}

	b.WriteString("</svg>\n")
	return []byte(b.String())
}
