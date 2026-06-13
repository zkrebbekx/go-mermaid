package sankey

import (
	"fmt"
	"strings"

	"github.com/Zac300/go-mermaid/internal/svgutil"
	"github.com/Zac300/go-mermaid/internal/theme"
)

// RenderOptions controls sankey appearance.
type RenderOptions struct {
	Theme    string
	FontFace string
	FontSize float64
	Padding  float64
	Title    string
}

var nodeColors = []string{"#5B8FF9", "#61DDAA", "#65789B", "#F6BD16", "#7262FD", "#78D3F8", "#9661BC", "#F6903D", "#008685", "#F08BB4"}

const (
	chartH = 360.0
	colGap = 180.0
	nodeW  = 14.0
	vGap   = 10.0 // vertical gap between stacked nodes
)

type placed struct {
	name       string
	col        int
	value      float64
	x, yTop, h float64
	outRun     float64
	inRun      float64
	colorIdx   int
}

// Render parses and renders sankey source to SVG.
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

	nodes := layoutNodes(d)
	maxCol := 0
	for _, n := range nodes {
		if n.col > maxCol {
			maxCol = n.col
		}
	}

	left := pad
	top := pad + titleH
	for _, n := range nodes {
		n.x = left + float64(n.col)*colGap
		n.yTop += top
	}

	w := left + float64(maxCol)*colGap + nodeW + 160 + pad // room for right labels
	h := top + chartH + pad

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

	byName := map[string]*placed{}
	for _, n := range nodes {
		byName[n.name] = n
	}
	scale := nodeScale(nodes)

	// Flow bands (drawn first, behind node bars).
	for _, f := range d.Flows {
		s, t := byName[f.Source], byName[f.Target]
		if s == nil || t == nil {
			continue
		}
		thick := f.Value * scale
		sy := s.yTop + s.outRun
		ty := t.yTop + t.inRun
		s.outRun += thick
		t.inRun += thick
		x0 := s.x + nodeW
		x1 := t.x
		mx := (x0 + x1) / 2
		color := nodeColors[s.colorIdx%len(nodeColors)]
		fmt.Fprintf(&b, `  <path d="M%s,%s C%s,%s %s,%s %s,%s L%s,%s C%s,%s %s,%s %s,%s Z" fill="%s" fill-opacity="0.4"/>`,
			svgutil.Num(x0), svgutil.Num(sy), svgutil.Num(mx), svgutil.Num(sy), svgutil.Num(mx), svgutil.Num(ty), svgutil.Num(x1), svgutil.Num(ty),
			svgutil.Num(x1), svgutil.Num(ty+thick), svgutil.Num(mx), svgutil.Num(ty+thick), svgutil.Num(mx), svgutil.Num(sy+thick), svgutil.Num(x0), svgutil.Num(sy+thick),
			color)
		b.WriteByte('\n')
	}

	// Node bars and labels.
	for _, n := range nodes {
		color := nodeColors[n.colorIdx%len(nodeColors)]
		fmt.Fprintf(&b, `  <rect x="%s" y="%s" width="%s" height="%s" fill="%s"/>`,
			svgutil.Num(n.x), svgutil.Num(n.yTop), svgutil.Num(nodeW), svgutil.Num(n.h), color)
		b.WriteByte('\n')
		fmt.Fprintf(&b, `  <text x="%s" y="%s" fill="%s">%s</text>`,
			svgutil.Num(n.x+nodeW+4), svgutil.Num(n.yTop+n.h/2+o.FontSize*0.35), pal.Text, svgutil.Esc(n.name))
		b.WriteByte('\n')
	}

	b.WriteString("</svg>\n")
	return []byte(b.String())
}

// layoutNodes assigns columns (longest path) and stacks nodes per column.
func layoutNodes(d *Diagram) []*placed {
	succ := map[string][]string{}
	indeg := map[string]int{}
	value := map[string]float64{}
	out := map[string]float64{}
	in := map[string]float64{}
	for _, name := range d.Nodes {
		indeg[name] = 0
	}
	for _, f := range d.Flows {
		succ[f.Source] = append(succ[f.Source], f.Target)
		indeg[f.Target]++
		out[f.Source] += f.Value
		in[f.Target] += f.Value
	}
	for _, name := range d.Nodes {
		value[name] = maxF(out[name], in[name])
	}

	// Longest-path columns via Kahn's algorithm.
	col := map[string]int{}
	deg := map[string]int{}
	for k, v := range indeg {
		deg[k] = v
	}
	var queue []string
	for _, name := range d.Nodes {
		if deg[name] == 0 {
			queue = append(queue, name)
		}
	}
	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]
		for _, s := range succ[n] {
			if col[n]+1 > col[s] {
				col[s] = col[n] + 1
			}
			deg[s]--
			if deg[s] == 0 {
				queue = append(queue, s)
			}
		}
	}

	placedByCol := map[int][]*placed{}
	idx := 0
	result := make([]*placed, 0, len(d.Nodes))
	for _, name := range d.Nodes {
		p := &placed{name: name, col: col[name], value: value[name], colorIdx: idx}
		idx++
		result = append(result, p)
		placedByCol[col[name]] = append(placedByCol[col[name]], p)
	}

	scale := nodeScale(result)
	for _, group := range placedByCol {
		y := 0.0
		for _, p := range group {
			p.h = p.value * scale
			p.yTop = y
			y += p.h + vGap
		}
	}
	return result
}

// nodeScale returns pixels-per-unit so the tallest column fits chartH.
func nodeScale(nodes []*placed) float64 {
	colTotal := map[int]float64{}
	colCount := map[int]int{}
	for _, n := range nodes {
		colTotal[n.col] += n.value
		colCount[n.col]++
	}
	maxTotal := 0.0
	maxCount := 1
	for c, tot := range colTotal {
		if tot > maxTotal {
			maxTotal = tot
		}
		if colCount[c] > maxCount {
			maxCount = colCount[c]
		}
	}
	if maxTotal <= 0 {
		return 1
	}
	usable := chartH - float64(maxCount-1)*vGap
	if usable < 20 {
		usable = 20
	}
	return usable / maxTotal
}

func maxF(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
