package quadrant

import (
	"fmt"
	"strings"

	"github.com/zkrebbekx/go-mermaid/internal/svgutil"
	"github.com/zkrebbekx/go-mermaid/internal/theme"
)

// RenderOptions controls quadrant chart appearance.
type RenderOptions struct {
	Theme    string
	FontFace string
	FontSize float64
	Padding  float64
	Title    string
}

const (
	plotSize = 320.0
	axisGap  = 26.0 // room for axis labels around the plot
)

// quadFills are light backgrounds for quadrants 1..4 (TR, TL, BL, BR).
var quadFills = [4]string{"#e8f5e9", "#fff8e1", "#fbe9e7", "#e3f2fd"}

// Render parses and renders quadrant chart source to SVG.
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

	left := pad + axisGap
	top := pad + titleH + axisGap
	w := left + plotSize + axisGap + pad
	h := top + plotSize + axisGap + pad
	midX := left + plotSize/2
	midY := top + plotSize/2

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

	// Quadrant fills: TR=Q1, TL=Q2, BL=Q3, BR=Q4.
	half := plotSize / 2
	rect := func(x, y float64, fill string) {
		fmt.Fprintf(&b, `  <rect x="%s" y="%s" width="%s" height="%s" fill="%s"/>`,
			svgutil.Num(x), svgutil.Num(y), svgutil.Num(half), svgutil.Num(half), fill)
		b.WriteByte('\n')
	}
	rect(midX, top, quadFills[0])  // Q1 top-right
	rect(left, top, quadFills[1])  // Q2 top-left
	rect(left, midY, quadFills[2]) // Q3 bottom-left
	rect(midX, midY, quadFills[3]) // Q4 bottom-right

	// Border and mid lines.
	fmt.Fprintf(&b, `  <rect x="%s" y="%s" width="%s" height="%s" fill="none" stroke="%s"/>`,
		svgutil.Num(left), svgutil.Num(top), svgutil.Num(plotSize), svgutil.Num(plotSize), pal.NodeStroke)
	b.WriteByte('\n')
	fmt.Fprintf(&b, `  <line x1="%s" y1="%s" x2="%s" y2="%s" stroke="%s"/><line x1="%s" y1="%s" x2="%s" y2="%s" stroke="%s"/>`,
		svgutil.Num(midX), svgutil.Num(top), svgutil.Num(midX), svgutil.Num(top+plotSize), pal.NodeStroke,
		svgutil.Num(left), svgutil.Num(midY), svgutil.Num(left+plotSize), svgutil.Num(midY), pal.NodeStroke)
	b.WriteByte('\n')

	// Quadrant labels at the top of each quadrant.
	qlabel := func(text string, cx, qy float64) {
		if text == "" {
			return
		}
		fmt.Fprintf(&b, `  <text x="%s" y="%s" fill="%s" text-anchor="middle">%s</text>`,
			svgutil.Num(cx), svgutil.Num(qy), pal.Text, svgutil.Esc(text))
		b.WriteByte('\n')
	}
	qlabel(d.Quadrant[0], midX+half/2, top+o.FontSize+2)
	qlabel(d.Quadrant[1], left+half/2, top+o.FontSize+2)
	qlabel(d.Quadrant[2], left+half/2, midY+o.FontSize+2)
	qlabel(d.Quadrant[3], midX+half/2, midY+o.FontSize+2)

	// Axis labels.
	axisText := func(text string, x, y, anchor string) {
		if text == "" {
			return
		}
		fmt.Fprintf(&b, `  <text x="%s" y="%s" fill="%s" text-anchor="%s">%s</text>`, x, y, pal.Text, anchor, svgutil.Esc(text))
		b.WriteByte('\n')
	}
	by := svgutil.Num(top + plotSize + o.FontSize + 4)
	axisText(d.XLeft, svgutil.Num(left), by, "start")
	axisText(d.XRight, svgutil.Num(left+plotSize), by, "end")
	// Y-axis labels run vertically along the left edge so long text fits in the
	// narrow gutter instead of spilling off the canvas.
	yAxisText := func(text string, cy float64) {
		if text == "" {
			return
		}
		lx := left - o.FontSize
		fmt.Fprintf(&b, `  <text x="%s" y="%s" fill="%s" text-anchor="middle" transform="rotate(-90 %s %s)">%s</text>`,
			svgutil.Num(lx), svgutil.Num(cy), pal.Text, svgutil.Num(lx), svgutil.Num(cy), svgutil.Esc(text))
		b.WriteByte('\n')
	}
	yAxisText(d.YTop, top+plotSize/4)
	yAxisText(d.YBottom, top+plotSize*3/4)

	// Points: X right, Y up.
	for _, p := range d.Points {
		cx := left + clampUnit(p.X)*plotSize
		cy := top + (1-clampUnit(p.Y))*plotSize
		fmt.Fprintf(&b, `  <circle cx="%s" cy="%s" r="6" fill="%s" stroke="%s"/>`,
			svgutil.Num(cx), svgutil.Num(cy), pal.NodeFill, pal.NodeStroke)
		b.WriteByte('\n')
		// Place the label to the right of the dot, but flip it left when it would
		// run past the right margin.
		lx, anchor := cx+9, "start"
		if cx+9+svgutil.TextWidth(p.Label, o.FontSize) > w-pad {
			lx, anchor = cx-9, "end"
		}
		fmt.Fprintf(&b, `  <text x="%s" y="%s" fill="%s" text-anchor="%s">%s</text>`,
			svgutil.Num(lx), svgutil.Num(cy+o.FontSize*0.35), pal.Text, anchor, svgutil.Esc(p.Label))
		b.WriteByte('\n')
	}

	b.WriteString("</svg>\n")
	return []byte(b.String())
}

func clampUnit(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
