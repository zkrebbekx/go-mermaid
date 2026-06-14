package pie

import (
	"fmt"
	"math"
	"strings"

	"github.com/zkrebbekx/go-mermaid/internal/svgutil"
	"github.com/zkrebbekx/go-mermaid/internal/theme"
)

// RenderOptions controls pie chart appearance.
type RenderOptions struct {
	Theme    string
	FontFace string
	FontSize float64
	Padding  float64
	Title    string
}

// sliceColors is the rotating wedge palette.
var sliceColors = []string{
	"#5B8FF9", "#61DDAA", "#65789B", "#F6BD16", "#7262FD",
	"#78D3F8", "#9661BC", "#F6903D", "#008685", "#F08BB4",
}

// Render parses and renders pie chart source to SVG.
func Render(src string, o RenderOptions) ([]byte, error) {
	d, err := Parse(src)
	if err != nil {
		return nil, err
	}
	if o.Title == "" {
		o.Title = d.Title
	}
	return svg(d, o), nil
}

func svg(d *Diagram, o RenderOptions) []byte {
	pal := theme.For(o.Theme)
	pad := o.Padding
	const r = 120.0
	titleH := svgutil.TitleHeight(o.Title, o.FontSize)

	legendW := legendWidth(d, o.FontSize)
	cx := pad + r
	cy := pad + titleH + r
	w := pad + 2*r + 24 + legendW + pad
	h := pad + titleH + 2*r + pad
	if lh := pad + titleH + float64(len(d.Slices))*(o.FontSize+8) + pad; lh > h {
		h = lh
	}

	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%s" height="%s" viewBox="0 0 %s %s" font-family="%s" font-size="%s">`,
		svgutil.Num(w), svgutil.Num(h), svgutil.Num(w), svgutil.Num(h), svgutil.Esc(o.FontFace), svgutil.Num(o.FontSize))
	b.WriteByte('\n')
	fmt.Fprintf(&b, `  <rect width="100%%" height="100%%" fill="%s"/>`, pal.Background)
	b.WriteByte('\n')
	if o.Title != "" {
		fmt.Fprintf(&b, `  <text x="%s" y="%s" fill="%s" text-anchor="middle" font-weight="bold">%s</text>`,
			svgutil.Num(pad+r), svgutil.Num(pad+o.FontSize), pal.Text, svgutil.Esc(o.Title))
		b.WriteByte('\n')
	}

	total := d.Total()
	writeSlices(&b, d, total, cx, cy, r, pal)
	writeLegend(&b, d, total, pad+2*r+24, pad+titleH, o, pal)

	b.WriteString("</svg>\n")
	return []byte(b.String())
}

func writeSlices(b *strings.Builder, d *Diagram, total, cx, cy, r float64, pal theme.Palette) {
	if total <= 0 {
		return
	}
	angle := -math.Pi / 2 // start at the top
	for i, s := range d.Slices {
		sweep := s.Value / total * 2 * math.Pi
		color := sliceColors[i%len(sliceColors)]
		if len(d.Slices) == 1 || sweep >= 2*math.Pi {
			fmt.Fprintf(b, `  <circle cx="%s" cy="%s" r="%s" fill="%s" stroke="%s"/>`,
				svgutil.Num(cx), svgutil.Num(cy), svgutil.Num(r), color, pal.Background)
			b.WriteByte('\n')
			return
		}
		x1, y1 := cx+r*math.Cos(angle), cy+r*math.Sin(angle)
		angle += sweep
		x2, y2 := cx+r*math.Cos(angle), cy+r*math.Sin(angle)
		largeArc := 0
		if sweep > math.Pi {
			largeArc = 1
		}
		fmt.Fprintf(b, `  <path d="M%s,%s L%s,%s A%s,%s 0 %d 1 %s,%s Z" fill="%s" stroke="%s"/>`,
			svgutil.Num(cx), svgutil.Num(cy), svgutil.Num(x1), svgutil.Num(y1),
			svgutil.Num(r), svgutil.Num(r), largeArc, svgutil.Num(x2), svgutil.Num(y2),
			color, pal.Background)
		b.WriteByte('\n')
	}
}

func writeLegend(b *strings.Builder, d *Diagram, total, x, y float64, o RenderOptions, pal theme.Palette) {
	row := o.FontSize + 8
	for i, s := range d.Slices {
		cy := y + float64(i)*row
		color := sliceColors[i%len(sliceColors)]
		fmt.Fprintf(b, `  <rect x="%s" y="%s" width="%s" height="%s" fill="%s"/>`,
			svgutil.Num(x), svgutil.Num(cy), svgutil.Num(o.FontSize), svgutil.Num(o.FontSize), color)
		b.WriteByte('\n')
		pct := 0.0
		if total > 0 {
			pct = s.Value / total * 100
		}
		label := fmt.Sprintf("%s: %s (%.1f%%)", s.Label, trimNum(s.Value), pct)
		fmt.Fprintf(b, `  <text x="%s" y="%s" fill="%s">%s</text>`,
			svgutil.Num(x+o.FontSize+6), svgutil.Num(cy+o.FontSize*0.85), pal.Text, svgutil.Esc(label))
		b.WriteByte('\n')
	}
}

func legendWidth(d *Diagram, fontSize float64) float64 {
	maxLen := 0
	for _, s := range d.Slices {
		l := len([]rune(s.Label)) + 14 // label + ": value (pct%)"
		if l > maxLen {
			maxLen = l
		}
	}
	return fontSize*0.6*float64(maxLen) + fontSize + 6
}

func trimNum(f float64) string {
	if f == math.Trunc(f) {
		return fmt.Sprintf("%d", int64(f))
	}
	return svgutil.Num(f)
}
