package timeline

import (
	"fmt"
	"strings"

	"github.com/zkrebbekx/go-mermaid/internal/svgutil"
	"github.com/zkrebbekx/go-mermaid/internal/theme"
)

// RenderOptions controls timeline appearance.
type RenderOptions struct {
	Theme    string
	FontFace string
	FontSize float64
	Padding  float64
	Title    string
}

var sectionColors = []string{"#5B8FF9", "#61DDAA", "#F6BD16", "#7262FD", "#F6903D", "#008685"}

const (
	colW    = 150.0
	bandH   = 22.0
	periodH = 26.0
	eventH  = 30.0
)

// Render parses and renders timeline source to SVG.
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
	periods := d.Periods()

	maxEvents := 0
	for _, p := range periods {
		if len(p.Events) > maxEvents {
			maxEvents = len(p.Events)
		}
	}

	bandY := pad + titleH
	periodY := bandY + bandH + 8
	axisY := periodY + periodH
	w := pad*2 + float64(maxIntT(len(periods), 1))*colW
	h := axisY + 16 + float64(maxEvents)*eventH + pad

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

	// Axis line.
	fmt.Fprintf(&b, `  <line x1="%s" y1="%s" x2="%s" y2="%s" stroke="%s" stroke-width="2"/>`,
		svgutil.Num(pad), svgutil.Num(axisY), svgutil.Num(w-pad), svgutil.Num(axisY), pal.Edge)
	b.WriteByte('\n')

	// Section bands across their periods.
	idx := 0
	for si, sec := range d.Sections {
		if len(sec.Periods) == 0 {
			continue
		}
		x := pad + float64(idx)*colW
		bw := float64(len(sec.Periods)) * colW
		color := sectionColors[si%len(sectionColors)]
		fmt.Fprintf(&b, `  <rect x="%s" y="%s" width="%s" height="%s" fill="%s" fill-opacity="0.3"/>`,
			svgutil.Num(x), svgutil.Num(bandY), svgutil.Num(bw), svgutil.Num(bandH), color)
		b.WriteByte('\n')
		if sec.Name != "" {
			fmt.Fprintf(&b, `  <text x="%s" y="%s" fill="%s" text-anchor="middle">%s</text>`,
				svgutil.Num(x+bw/2), svgutil.Num(bandY+bandH*0.7), pal.Text, svgutil.Esc(sec.Name))
			b.WriteByte('\n')
		}
		idx += len(sec.Periods)
	}

	// Periods and their events.
	for i, p := range periods {
		cx := pad + float64(i)*colW + colW/2
		fmt.Fprintf(&b, `  <text x="%s" y="%s" fill="%s" text-anchor="middle" font-weight="bold">%s</text>`,
			svgutil.Num(cx), svgutil.Num(periodY+o.FontSize), pal.Text, svgutil.Esc(p.Time))
		b.WriteByte('\n')
		fmt.Fprintf(&b, `  <circle cx="%s" cy="%s" r="5" fill="%s" stroke="%s"/>`,
			svgutil.Num(cx), svgutil.Num(axisY), pal.NodeFill, pal.NodeStroke)
		b.WriteByte('\n')
		for j, ev := range p.Events {
			ey := axisY + 16 + float64(j)*eventH
			fmt.Fprintf(&b, `  <rect x="%s" y="%s" width="%s" height="%s" rx="4" fill="%s" stroke="%s"/>`,
				svgutil.Num(cx-colW/2+8), svgutil.Num(ey), svgutil.Num(colW-16), svgutil.Num(eventH-8), pal.NodeFill, pal.NodeStroke)
			b.WriteByte('\n')
			fmt.Fprintf(&b, `  <text x="%s" y="%s" fill="%s" text-anchor="middle">%s</text>`,
				svgutil.Num(cx), svgutil.Num(ey+(eventH-8)/2+o.FontSize*0.35), pal.Text, svgutil.Esc(ev))
			b.WriteByte('\n')
		}
	}

	b.WriteString("</svg>\n")
	return []byte(b.String())
}

func maxIntT(a, b int) int {
	if a > b {
		return a
	}
	return b
}
