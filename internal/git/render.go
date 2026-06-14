package git

import (
	"fmt"
	"strings"

	"github.com/zkrebbekx/go-mermaid/internal/svgutil"
	"github.com/zkrebbekx/go-mermaid/internal/theme"
)

// RenderOptions controls gitGraph appearance.
type RenderOptions struct {
	Theme    string
	FontFace string
	FontSize float64
	Padding  float64
	Title    string
}

// branchColors cycles per lane.
var branchColors = []string{"#5B8FF9", "#61DDAA", "#F6BD16", "#7262FD", "#F6903D", "#008685"}

const (
	commitGap = 52.0
	laneH     = 50.0
	dotR      = 7.0
	labelGap  = 70.0 // left room for branch names
)

// Render parses and renders gitGraph source to SVG.
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

	originX := pad + labelGap
	originY := pad + titleH + 20
	maxOrder := 0
	for _, c := range d.Commits {
		if c.Order > maxOrder {
			maxOrder = c.Order
		}
	}
	w := originX + float64(maxOrder+1)*commitGap + pad
	h := originY + float64(len(d.Branches))*laneH + pad

	cx := func(order int) float64 { return originX + float64(order)*commitGap }
	cy := func(lane int) float64 { return originY + float64(lane)*laneH }
	colorOf := func(lane int) string { return branchColors[lane%len(branchColors)] }

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

	// Branch lane lines and labels.
	tip := lastOrderPerBranch(d)
	for _, br := range d.Branches {
		y := cy(br.Lane)
		end := cx(tip[br.Name])
		if tip[br.Name] < 0 {
			end = originX
		}
		fmt.Fprintf(&b, `  <line x1="%s" y1="%s" x2="%s" y2="%s" stroke="%s" stroke-width="2"/>`,
			svgutil.Num(originX), svgutil.Num(y), svgutil.Num(end), svgutil.Num(y), colorOf(br.Lane))
		b.WriteByte('\n')
		fmt.Fprintf(&b, `  <text x="%s" y="%s" fill="%s" text-anchor="end">%s</text>`,
			svgutil.Num(originX-12), svgutil.Num(y+o.FontSize*0.35), pal.Text, svgutil.Esc(br.Name))
		b.WriteByte('\n')
	}

	// Merge connectors.
	laneOf := commitLane(d)
	for _, m := range d.Merges {
		fmt.Fprintf(&b, `  <line x1="%s" y1="%s" x2="%s" y2="%s" stroke="%s" stroke-dasharray="3,3"/>`,
			svgutil.Num(cx(m.FromOrder)), svgutil.Num(cy(laneOf[m.FromOrder])),
			svgutil.Num(cx(m.ToOrder)), svgutil.Num(cy(laneOf[m.ToOrder])), pal.Edge)
		b.WriteByte('\n')
	}

	// Commit dots with id/tag labels.
	for _, c := range d.Commits {
		x, y := cx(c.Order), cy(d.lane(c.Branch))
		fmt.Fprintf(&b, `  <circle cx="%s" cy="%s" r="%s" fill="%s" stroke="%s"/>`,
			svgutil.Num(x), svgutil.Num(y), svgutil.Num(dotR), colorOf(d.lane(c.Branch)), pal.NodeStroke)
		b.WriteByte('\n')
		if lbl := commitLabel(c); lbl != "" {
			fmt.Fprintf(&b, `  <text x="%s" y="%s" fill="%s" text-anchor="middle">%s</text>`,
				svgutil.Num(x), svgutil.Num(y-dotR-4), pal.Text, svgutil.Esc(lbl))
			b.WriteByte('\n')
		}
	}

	b.WriteString("</svg>\n")
	return []byte(b.String())
}

func commitLabel(c *Commit) string {
	switch {
	case c.Tag != "":
		return c.Tag
	case c.ID != "":
		return c.ID
	}
	return ""
}

func lastOrderPerBranch(d *Diagram) map[string]int {
	m := map[string]int{}
	for _, br := range d.Branches {
		m[br.Name] = -1
	}
	for _, c := range d.Commits {
		m[c.Branch] = c.Order
	}
	return m
}

func commitLane(d *Diagram) map[int]int {
	m := map[int]int{}
	for _, c := range d.Commits {
		m[c.Order] = d.lane(c.Branch)
	}
	return m
}
