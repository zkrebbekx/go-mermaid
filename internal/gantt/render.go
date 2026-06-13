package gantt

import (
	"fmt"
	"strings"
	"time"

	"github.com/Zac300/go-mermaid/internal/svgutil"
	"github.com/Zac300/go-mermaid/internal/theme"
)

// RenderOptions controls gantt appearance.
type RenderOptions struct {
	Theme    string
	FontFace string
	FontSize float64
	Padding  float64
	Title    string
}

const (
	labelW = 170.0
	chartW = 560.0
	rowH   = 26.0
	axisH  = 22.0
)

var sectionColors = []string{"#5B8FF9", "#61DDAA", "#F6BD16", "#7262FD", "#F6903D", "#008685"}

// Render parses and renders gantt source to SVG.
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

	minT, maxT := d.Bounds()
	totalDays := maxT.Sub(minT).Hours() / 24
	if totalDays <= 0 {
		totalDays = 1
	}

	chartLeft := pad + labelW
	chartTop := pad + titleH + axisH
	w := chartLeft + chartW + pad
	h := chartTop + float64(len(d.Tasks))*rowH + pad

	dayX := func(t time.Time) float64 {
		days := t.Sub(minT).Hours() / 24
		return chartLeft + days/totalDays*chartW
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

	// Axis: start and end date labels with bounding gridlines.
	if !minT.IsZero() {
		fmt.Fprintf(&b, `  <text x="%s" y="%s" fill="%s">%s</text>`,
			svgutil.Num(chartLeft), svgutil.Num(chartTop-6), pal.Text, svgutil.Esc(minT.Format("2006-01-02")))
		b.WriteByte('\n')
		fmt.Fprintf(&b, `  <text x="%s" y="%s" fill="%s" text-anchor="end">%s</text>`,
			svgutil.Num(chartLeft+chartW), svgutil.Num(chartTop-6), pal.Text, svgutil.Esc(maxT.Format("2006-01-02")))
		b.WriteByte('\n')
	}
	fmt.Fprintf(&b, `  <line x1="%s" y1="%s" x2="%s" y2="%s" stroke="%s"/>`,
		svgutil.Num(chartLeft), svgutil.Num(chartTop), svgutil.Num(chartLeft+chartW), svgutil.Num(chartTop), pal.NodeStroke)
	b.WriteByte('\n')

	sectionIndex := map[string]int{}
	for i, s := range d.Sections {
		sectionIndex[s] = i
	}

	for i, t := range d.Tasks {
		y := chartTop + float64(i)*rowH + 3
		bh := rowH - 6
		color := sectionColors[sectionIndex[t.Section]%len(sectionColors)]

		// Task name in the left gutter.
		fmt.Fprintf(&b, `  <text x="%s" y="%s" fill="%s">%s</text>`,
			svgutil.Num(pad), svgutil.Num(y+bh/2+o.FontSize*0.35), pal.Text, svgutil.Esc(t.Name))
		b.WriteByte('\n')

		if t.Start.IsZero() {
			continue
		}
		x1, x2 := dayX(t.Start), dayX(t.End())
		bw := x2 - x1
		if bw < 2 {
			bw = 2
		}
		fmt.Fprintf(&b, `  <rect x="%s" y="%s" width="%s" height="%s" rx="3" fill="%s" stroke="%s"/>`,
			svgutil.Num(x1), svgutil.Num(y), svgutil.Num(bw), svgutil.Num(bh), color, pal.NodeStroke)
		b.WriteByte('\n')
	}

	b.WriteString("</svg>\n")
	return []byte(b.String())
}
