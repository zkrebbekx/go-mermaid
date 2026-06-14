// Package radar parses and renders Mermaid radar-beta charts to SVG as a polar
// plot with one polygon per data curve.
//
// Syntax (subset):
//
//	radar-beta
//	    title Skills
//	    axis a["Speed"], b["Power"], c["Range"]
//	    curve s1["Team A"]{80, 60, 90}
//	    curve s2["Team B"]{50, 90, 40}
package radar

import (
	"strconv"
	"strings"

	"github.com/zkrebbekx/go-mermaid/internal/syntax"
)

// Curve is a named series of values, one per axis.
type Curve struct {
	Name   string
	Values []float64
}

// Diagram is a parsed radar chart.
type Diagram struct {
	Title  string
	Axes   []string
	Curves []*Curve
}

// Max returns the largest value across curves (at least 1).
func (d *Diagram) Max() float64 {
	m := 1.0
	for _, c := range d.Curves {
		for _, v := range c.Values {
			if v > m {
				m = v
			}
		}
	}
	return m
}

// Parse builds a Diagram from radar-beta source.
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
			if w := strings.ToLower(firstWord(line)); w != "radar-beta" && w != "radar" {
				return nil, syntax.Errorf(lineNo, 1, "expected 'radar-beta' header")
			}
			headerSeen = true
			continue
		}
		switch {
		case strings.HasPrefix(line, "title "):
			d.Title = strings.TrimSpace(line[len("title "):])
		case strings.HasPrefix(line, "axis "):
			for _, tok := range strings.Split(line[len("axis "):], ",") {
				if name := labelOf(strings.TrimSpace(tok)); name != "" {
					d.Axes = append(d.Axes, name)
				}
			}
		case strings.HasPrefix(line, "curve "):
			d.Curves = append(d.Curves, parseCurve(line[len("curve "):]))
		}
	}
	if !headerSeen {
		return nil, syntax.Errorf(1, 1, "expected 'radar-beta' header")
	}
	if len(d.Axes) == 0 {
		return nil, syntax.Errorf(1, 1, "radar chart has no axes")
	}
	return d, nil
}

func parseCurve(s string) *Curve {
	c := &Curve{}
	if o := strings.IndexByte(s, '{'); o >= 0 {
		c.Name = labelOf(strings.TrimSpace(s[:o]))
		inner := s[o+1:]
		if cl := strings.IndexByte(inner, '}'); cl >= 0 {
			inner = inner[:cl]
		}
		for _, f := range strings.Split(inner, ",") {
			if v, err := strconv.ParseFloat(strings.TrimSpace(f), 64); err == nil {
				c.Values = append(c.Values, v)
			}
		}
	} else {
		c.Name = labelOf(strings.TrimSpace(s))
	}
	return c
}

// labelOf returns the bracketed label of "id[\"Label\"]", else the token.
func labelOf(tok string) string {
	if o := strings.IndexByte(tok, '['); o >= 0 && strings.HasSuffix(tok, "]") {
		return strings.Trim(strings.TrimSuffix(tok[o+1:], "]"), `"`)
	}
	return tok
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
