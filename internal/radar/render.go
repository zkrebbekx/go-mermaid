package radar

import (
	"fmt"
	"math"
	"strings"

	"github.com/zkrebbekx/go-mermaid/internal/svgutil"
	"github.com/zkrebbekx/go-mermaid/internal/theme"
)

// RenderOptions controls radar appearance.
type RenderOptions struct {
	Theme    string
	FontFace string
	FontSize float64
	Padding  float64
	Title    string
}

var curveColors = []string{"#5B8FF9", "#F6903D", "#61DDAA", "#7262FD", "#F6BD16"}

const (
	radius = 150.0
	margin = 70.0 // room for axis labels around the plot
)

// Render parses and renders radar-beta source to SVG.
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
	titleH := svgutil.TitleHeight(o.Title, o.FontSize)

	cx := pad + margin + radius
	cy := pad + titleH + margin + radius
	w := cx + radius + margin + pad
	h := cy + radius + margin + pad
	n := len(d.Axes)
	maxV := d.Max()

	angle := func(i int) float64 { return -math.Pi/2 + 2*math.Pi*float64(i)/float64(n) }
	point := func(i int, frac float64) (float64, float64) {
		a := angle(i)
		return cx + radius*frac*math.Cos(a), cy + radius*frac*math.Sin(a)
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

	// Concentric grid rings.
	for _, frac := range []float64{0.25, 0.5, 0.75, 1.0} {
		var ring strings.Builder
		for i := 0; i < n; i++ {
			x, y := point(i, frac)
			cmd := "L"
			if i == 0 {
				cmd = "M"
			}
			fmt.Fprintf(&ring, "%s%s,%s ", cmd, svgutil.Num(x), svgutil.Num(y))
		}
		fmt.Fprintf(&b, `  <path d="%sZ" fill="none" stroke="%s" stroke-opacity="0.4"/>`, ring.String(), pal.NodeStroke)
		b.WriteByte('\n')
	}

	// Spokes and axis labels.
	for i, ax := range d.Axes {
		x, y := point(i, 1)
		fmt.Fprintf(&b, `  <line x1="%s" y1="%s" x2="%s" y2="%s" stroke="%s" stroke-opacity="0.4"/>`,
			svgutil.Num(cx), svgutil.Num(cy), svgutil.Num(x), svgutil.Num(y), pal.NodeStroke)
		b.WriteByte('\n')
		lx, ly := point(i, 1.12)
		fmt.Fprintf(&b, `  <text x="%s" y="%s" fill="%s" text-anchor="middle">%s</text>`,
			svgutil.Num(lx), svgutil.Num(ly), pal.Text, svgutil.Esc(ax))
		b.WriteByte('\n')
	}

	// Curves.
	for ci, c := range d.Curves {
		color := curveColors[ci%len(curveColors)]
		var poly strings.Builder
		for i := 0; i < n; i++ {
			v := 0.0
			if i < len(c.Values) {
				v = c.Values[i]
			}
			x, y := point(i, v/maxV)
			cmd := "L"
			if i == 0 {
				cmd = "M"
			}
			fmt.Fprintf(&poly, "%s%s,%s ", cmd, svgutil.Num(x), svgutil.Num(y))
		}
		fmt.Fprintf(&b, `  <path d="%sZ" fill="%s" fill-opacity="0.2" stroke="%s" stroke-width="2"/>`, poly.String(), color, color)
		b.WriteByte('\n')
	}

	// Legend.
	for ci, c := range d.Curves {
		ly := pad + titleH + float64(ci)*(o.FontSize+6)
		color := curveColors[ci%len(curveColors)]
		fmt.Fprintf(&b, `  <rect x="%s" y="%s" width="%s" height="%s" fill="%s"/>`,
			svgutil.Num(pad), svgutil.Num(ly), svgutil.Num(o.FontSize), svgutil.Num(o.FontSize), color)
		b.WriteByte('\n')
		fmt.Fprintf(&b, `  <text x="%s" y="%s" fill="%s">%s</text>`,
			svgutil.Num(pad+o.FontSize+5), svgutil.Num(ly+o.FontSize*0.85), pal.Text, svgutil.Esc(c.Name))
		b.WriteByte('\n')
	}

	b.WriteString("</svg>\n")
	return []byte(b.String())
}
