package packet

import (
	"fmt"
	"strings"

	"github.com/zkrebbekx/go-mermaid/internal/svgutil"
	"github.com/zkrebbekx/go-mermaid/internal/theme"
)

// RenderOptions controls packet diagram appearance.
type RenderOptions struct {
	Theme    string
	FontFace string
	FontSize float64
	Padding  float64
	Title    string
}

const (
	bitsPerRow = 32
	bitW       = 20.0
	rowH       = 40.0
	bitNumH    = 14.0
)

// Render parses and renders packet-beta source to SVG.
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

	maxBit := 0
	for _, f := range d.Fields {
		if f.End > maxBit {
			maxBit = f.End
		}
	}
	rowCount := maxBit/bitsPerRow + 1

	left := pad
	top := pad + titleH + bitNumH
	w := left + bitsPerRow*bitW + pad
	h := top + float64(rowCount)*rowH + pad

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

	// Bit-index ruler across the top row.
	for i := 0; i < bitsPerRow; i += 8 {
		fmt.Fprintf(&b, `  <text x="%s" y="%s" fill="%s" text-anchor="middle" font-size="%s">%d</text>`,
			svgutil.Num(left+float64(i)*bitW+bitW/2), svgutil.Num(top-3), pal.Text, svgutil.Num(o.FontSize*0.7), i)
		b.WriteByte('\n')
	}

	// Each field, split into per-row segments.
	for _, f := range d.Fields {
		for bit := f.Start; bit <= f.End; {
			row := bit / bitsPerRow
			rowEnd := (row+1)*bitsPerRow - 1
			segEnd := f.End
			if rowEnd < segEnd {
				segEnd = rowEnd
			}
			col := bit % bitsPerRow
			x := left + float64(col)*bitW
			y := top + float64(row)*rowH
			segW := float64(segEnd-bit+1) * bitW
			fmt.Fprintf(&b, `  <rect x="%s" y="%s" width="%s" height="%s" fill="%s" stroke="%s"/>`,
				svgutil.Num(x), svgutil.Num(y), svgutil.Num(segW), svgutil.Num(rowH-6), pal.NodeFill, pal.NodeStroke)
			b.WriteByte('\n')
			if bit == f.Start {
				fmt.Fprintf(&b, `  <text x="%s" y="%s" fill="%s" text-anchor="middle">%s</text>`,
					svgutil.Num(x+segW/2), svgutil.Num(y+(rowH-6)/2+o.FontSize*0.35), pal.Text, svgutil.Esc(f.Label))
				b.WriteByte('\n')
			}
			bit = segEnd + 1
		}
	}

	b.WriteString("</svg>\n")
	return []byte(b.String())
}
