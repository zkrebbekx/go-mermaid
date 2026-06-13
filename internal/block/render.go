package block

import (
	"fmt"
	"strings"

	"github.com/Zac300/go-mermaid/internal/svgutil"
	"github.com/Zac300/go-mermaid/internal/theme"
)

// RenderOptions controls block diagram appearance.
type RenderOptions struct {
	Theme    string
	FontFace string
	FontSize float64
	Padding  float64
	Title    string
}

const (
	cellW = 110.0
	rowH  = 48.0
	gap   = 8.0
)

// Render parses and renders block-beta source to SVG.
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

	gridW := float64(d.Columns) * cellW
	w := gridW + pad*2
	top := pad + titleH
	h := top + float64(len(d.Rows))*rowH + pad

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

	for ri, row := range d.Rows {
		y := top + float64(ri)*rowH
		col := 0
		for _, blk := range row {
			span := blk.Span
			if col+span > d.Columns {
				span = d.Columns - col
				if span < 1 {
					span = 1
				}
			}
			x := pad + float64(col)*cellW
			bw := float64(span)*cellW - gap
			fmt.Fprintf(&b, `  <rect x="%s" y="%s" width="%s" height="%s" rx="4" fill="%s" stroke="%s"/>`,
				svgutil.Num(x), svgutil.Num(y), svgutil.Num(bw), svgutil.Num(rowH-gap), pal.NodeFill, pal.NodeStroke)
			b.WriteByte('\n')
			fmt.Fprintf(&b, `  <text x="%s" y="%s" fill="%s" text-anchor="middle">%s</text>`,
				svgutil.Num(x+bw/2), svgutil.Num(y+(rowH-gap)/2+o.FontSize*0.35), pal.Text, svgutil.Esc(blk.Label))
			b.WriteByte('\n')
			col += blk.Span
		}
	}

	b.WriteString("</svg>\n")
	return []byte(b.String())
}
