package xychart

import (
	"fmt"
	"strings"

	"github.com/zkrebbekx/go-mermaid/internal/svgutil"
	"github.com/zkrebbekx/go-mermaid/internal/theme"
)

// RenderOptions controls xychart appearance.
type RenderOptions struct {
	Theme    string
	FontFace string
	FontSize float64
	Padding  float64
	Title    string
}

var seriesColors = []string{"#5B8FF9", "#F6BD16", "#61DDAA", "#7262FD", "#F6903D"}

const (
	plotW   = 480.0
	plotH   = 280.0
	axisPad = 44.0 // room for y labels / x labels
)

// Render parses and renders xychart source to SVG.
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

	left := pad + axisPad
	top := pad + titleH
	bottom := top + plotH
	right := left + plotW
	w := right + pad
	h := bottom + axisPad + pad

	lo, hi := d.Bounds()
	span := hi - lo
	if span == 0 {
		span = 1
	}
	yPix := func(v float64) float64 { return bottom - (v-lo)/span*plotH }

	n := len(d.XCats)
	if n == 0 {
		for _, s := range d.Series {
			if len(s.Values) > n {
				n = len(s.Values)
			}
		}
	}
	if n == 0 {
		n = 1
	}
	band := plotW / float64(n)

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

	// Axes.
	fmt.Fprintf(&b, `  <line x1="%s" y1="%s" x2="%s" y2="%s" stroke="%s"/><line x1="%s" y1="%s" x2="%s" y2="%s" stroke="%s"/>`,
		svgutil.Num(left), svgutil.Num(top), svgutil.Num(left), svgutil.Num(bottom), pal.NodeStroke,
		svgutil.Num(left), svgutil.Num(bottom), svgutil.Num(right), svgutil.Num(bottom), pal.NodeStroke)
	b.WriteByte('\n')
	// y min/max labels.
	fmt.Fprintf(&b, `  <text x="%s" y="%s" fill="%s" text-anchor="end">%s</text>`,
		svgutil.Num(left-6), svgutil.Num(bottom), pal.Text, svgutil.Esc(trimNum(lo)))
	b.WriteByte('\n')
	fmt.Fprintf(&b, `  <text x="%s" y="%s" fill="%s" text-anchor="end">%s</text>`,
		svgutil.Num(left-6), svgutil.Num(top+o.FontSize), pal.Text, svgutil.Esc(trimNum(hi)))
	b.WriteByte('\n')

	// x category labels.
	for i, c := range d.XCats {
		fmt.Fprintf(&b, `  <text x="%s" y="%s" fill="%s" text-anchor="middle">%s</text>`,
			svgutil.Num(left+band*(float64(i)+0.5)), svgutil.Num(bottom+o.FontSize+4), pal.Text, svgutil.Esc(c))
		b.WriteByte('\n')
	}

	// Bars: group bar-series within each band.
	barSeries := filter(d.Series, "bar")
	for bi, s := range barSeries {
		color := seriesColors[seriesIdx(d, s)%len(seriesColors)]
		bw := band * 0.7 / float64(len(barSeries))
		for i, v := range s.Values {
			x := left + band*float64(i) + band*0.15 + float64(bi)*bw
			y := yPix(v)
			fmt.Fprintf(&b, `  <rect x="%s" y="%s" width="%s" height="%s" fill="%s"/>`,
				svgutil.Num(x), svgutil.Num(y), svgutil.Num(bw), svgutil.Num(bottom-y), color)
			b.WriteByte('\n')
		}
	}

	// Lines.
	for _, s := range filter(d.Series, "line") {
		color := seriesColors[seriesIdx(d, s)%len(seriesColors)]
		var p strings.Builder
		for i, v := range s.Values {
			cmd := "L"
			if i == 0 {
				cmd = "M"
			}
			fmt.Fprintf(&p, "%s%s,%s ", cmd, svgutil.Num(left+band*(float64(i)+0.5)), svgutil.Num(yPix(v)))
		}
		fmt.Fprintf(&b, `  <path d="%s" fill="none" stroke="%s" stroke-width="2"/>`, strings.TrimSpace(p.String()), color)
		b.WriteByte('\n')
	}

	b.WriteString("</svg>\n")
	return []byte(b.String())
}

func filter(series []*Series, kind string) []*Series {
	var out []*Series
	for _, s := range series {
		if s.Kind == kind {
			out = append(out, s)
		}
	}
	return out
}

func seriesIdx(d *Diagram, s *Series) int {
	for i, x := range d.Series {
		if x == s {
			return i
		}
	}
	return 0
}

func trimNum(f float64) string {
	if f == float64(int64(f)) {
		return fmt.Sprintf("%d", int64(f))
	}
	return svgutil.Num(f)
}
