// Package quadrant parses and renders Mermaid quadrant charts to SVG.
//
// Syntax:
//
//	quadrantChart
//	    title Reach and engagement
//	    x-axis Low Reach --> High Reach
//	    y-axis Low Engagement --> High Engagement
//	    quadrant-1 We should expand
//	    Campaign A: [0.3, 0.6]
package quadrant

import (
	"strconv"
	"strings"

	"github.com/zkrebbekx/go-mermaid/internal/syntax"
)

// Point is a plotted item with coordinates in the unit square.
type Point struct {
	Label string
	X, Y  float64
}

// Diagram is a parsed quadrant chart.
type Diagram struct {
	Title    string
	XLeft    string
	XRight   string
	YBottom  string
	YTop     string
	Quadrant [4]string // quadrant-1..4
	Points   []*Point
}

// Parse builds a Diagram from quadrant chart source.
func Parse(src string) (*Diagram, error) {
	d := &Diagram{}
	headerSeen := false
	for i, raw := range strings.Split(src, "\n") {
		lineNo := i + 1
		line := strings.TrimSpace(stripComment(raw))
		if line == "" {
			continue
		}
		if !headerSeen {
			if firstWord(line) != "quadrantChart" {
				return nil, syntax.Errorf(lineNo, 1, "expected 'quadrantChart' header")
			}
			headerSeen = true
			continue
		}
		switch {
		case strings.HasPrefix(line, "title "):
			d.Title = strings.TrimSpace(line[len("title "):])
		case strings.HasPrefix(line, "x-axis "):
			d.XLeft, d.XRight = splitAxis(line[len("x-axis "):])
		case strings.HasPrefix(line, "y-axis "):
			d.YBottom, d.YTop = splitAxis(line[len("y-axis "):])
		case strings.HasPrefix(line, "quadrant-"):
			parseQuadrant(d, line)
		default:
			p, err := parsePoint(line, lineNo)
			if err != nil {
				return nil, err
			}
			d.Points = append(d.Points, p)
		}
	}
	if !headerSeen {
		return nil, syntax.Errorf(1, 1, "expected 'quadrantChart' header")
	}
	return d, nil
}

func parseQuadrant(d *Diagram, line string) {
	num, label, _ := strings.Cut(line, " ")
	switch num {
	case "quadrant-1":
		d.Quadrant[0] = strings.TrimSpace(label)
	case "quadrant-2":
		d.Quadrant[1] = strings.TrimSpace(label)
	case "quadrant-3":
		d.Quadrant[2] = strings.TrimSpace(label)
	case "quadrant-4":
		d.Quadrant[3] = strings.TrimSpace(label)
	}
}

func parsePoint(line string, lineNo int) (*Point, error) {
	name, coords, ok := strings.Cut(line, ":")
	if !ok {
		return nil, syntax.Errorf(lineNo, 1, "expected 'label: [x, y]'")
	}
	coords = strings.TrimSpace(coords)
	coords = strings.TrimPrefix(coords, "[")
	coords = strings.TrimSuffix(coords, "]")
	xs, ys, ok := strings.Cut(coords, ",")
	if !ok {
		return nil, syntax.Errorf(lineNo, 1, "expected two coordinates")
	}
	x, err1 := strconv.ParseFloat(strings.TrimSpace(xs), 64)
	y, err2 := strconv.ParseFloat(strings.TrimSpace(ys), 64)
	if err1 != nil || err2 != nil {
		return nil, syntax.Errorf(lineNo, 1, "invalid coordinates")
	}
	return &Point{Label: strings.TrimSpace(name), X: x, Y: y}, nil
}

// splitAxis splits "Low --> High" into its two ends.
func splitAxis(s string) (lo, hi string) {
	if l, r, ok := strings.Cut(s, "-->"); ok {
		return strings.TrimSpace(l), strings.TrimSpace(r)
	}
	return strings.TrimSpace(s), ""
}

func firstWord(s string) string {
	if i := strings.IndexAny(s, " \t"); i >= 0 {
		return s[:i]
	}
	return s
}

func stripComment(s string) string {
	if i := strings.Index(s, "%%"); i >= 0 {
		return s[:i]
	}
	return s
}
