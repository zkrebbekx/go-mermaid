package journey

import (
	"fmt"
	"strings"

	"github.com/zkrebbekx/go-mermaid/internal/svgutil"
	"github.com/zkrebbekx/go-mermaid/internal/theme"
)

// RenderOptions controls journey diagram appearance.
type RenderOptions struct {
	Theme    string
	FontFace string
	FontSize float64
	Padding  float64
	Title    string
}

// scoreColors maps satisfaction scores 1..5 to colors (red to green).
var scoreColors = map[int]string{
	1: "#e74c3c", 2: "#e67e22", 3: "#f1c40f", 4: "#2ecc71", 5: "#27ae60",
}

const (
	colW   = 110.0
	chartH = 150.0
	bandH  = 22.0
	labelH = 40.0
	pointR = 9.0
)

// Render parses and renders journey source to SVG.
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
	tasks := d.Tasks()
	n := len(tasks)

	w := pad*2 + float64(maxInt(n, 1))*colW
	chartTop := pad + titleH + bandH
	h := chartTop + chartH + labelH + pad

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

	// Score-to-y mapping: score 5 near the top, 1 near the bottom.
	yFor := func(score int) float64 {
		s := float64(clamp(score, 1, 5))
		return chartTop + (5-s)/4*chartH
	}

	// Section bands across their task columns.
	idx := 0
	for si, sec := range d.Sections {
		if len(sec.Tasks) == 0 {
			continue
		}
		x := pad + float64(idx)*colW
		bw := float64(len(sec.Tasks)) * colW
		band := scoreColors[(si%5)+1]
		fmt.Fprintf(&b, `  <rect x="%s" y="%s" width="%s" height="%s" fill="%s" fill-opacity="0.25"/>`,
			svgutil.Num(x), svgutil.Num(pad+titleH), svgutil.Num(bw), svgutil.Num(bandH), band)
		b.WriteByte('\n')
		fmt.Fprintf(&b, `  <text x="%s" y="%s" fill="%s" text-anchor="middle">%s</text>`,
			svgutil.Num(x+bw/2), svgutil.Num(pad+titleH+bandH*0.7), pal.Text, svgutil.Esc(sec.Name))
		b.WriteByte('\n')
		idx += len(sec.Tasks)
	}

	// Journey line connecting task points.
	var path strings.Builder
	for i, t := range tasks {
		x := pad + float64(i)*colW + colW/2
		cmd := "L"
		if i == 0 {
			cmd = "M"
		}
		fmt.Fprintf(&path, "%s%s,%s ", cmd, svgutil.Num(x), svgutil.Num(yFor(t.Score)))
	}
	if n > 1 {
		fmt.Fprintf(&b, `  <path d="%s" fill="none" stroke="%s" stroke-width="2"/>`, strings.TrimSpace(path.String()), pal.Edge)
		b.WriteByte('\n')
	}

	// Task points and labels.
	for i, t := range tasks {
		x := pad + float64(i)*colW + colW/2
		y := yFor(t.Score)
		fmt.Fprintf(&b, `  <circle cx="%s" cy="%s" r="%s" fill="%s" stroke="%s"/>`,
			svgutil.Num(x), svgutil.Num(y), svgutil.Num(pointR), scoreColors[clamp(t.Score, 1, 5)], pal.NodeStroke)
		b.WriteByte('\n')
		ly := chartTop + chartH + o.FontSize
		fmt.Fprintf(&b, `  <text x="%s" y="%s" fill="%s" text-anchor="middle">%s</text>`,
			svgutil.Num(x), svgutil.Num(ly), pal.Text, svgutil.Esc(t.Name))
		b.WriteByte('\n')
		if len(t.Actors) > 0 {
			fmt.Fprintf(&b, `  <text x="%s" y="%s" fill="%s" text-anchor="middle" font-size="%s">%s</text>`,
				svgutil.Num(x), svgutil.Num(ly+o.FontSize), pal.Text, svgutil.Num(o.FontSize*0.8),
				svgutil.Esc(strings.Join(t.Actors, ", ")))
			b.WriteByte('\n')
		}
	}

	b.WriteString("</svg>\n")
	return []byte(b.String())
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
