// Package xychart parses and renders Mermaid xychart-beta diagrams (bar and
// line charts) to SVG.
//
// Syntax (subset):
//
//	xychart-beta
//	    title "Sales"
//	    x-axis [jan, feb, mar]
//	    y-axis "Revenue" 0 --> 10000
//	    bar [5000, 6000, 7500]
//	    line [4000, 6000, 9000]
package xychart

import (
	"strconv"
	"strings"

	"github.com/zkrebbekx/go-mermaid/internal/syntax"
)

// Series is one bar or line data set.
type Series struct {
	Kind   string // "bar" or "line"
	Values []float64
}

// Diagram is a parsed xychart.
type Diagram struct {
	Title     string
	XCats     []string
	YLabel    string
	YMin      float64
	YMax      float64
	HasYRange bool
	Series    []*Series
}

// Parse builds a Diagram from xychart-beta source.
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
			if w := strings.ToLower(firstWord(line)); w != "xychart-beta" && w != "xychart" {
				return nil, syntax.Errorf(lineNo, 1, "expected 'xychart-beta' header")
			}
			headerSeen = true
			continue
		}
		switch {
		case strings.HasPrefix(line, "title "):
			d.Title = unquote(strings.TrimSpace(line[len("title "):]))
		case strings.HasPrefix(line, "x-axis"):
			d.XCats = parseStrings(line)
		case strings.HasPrefix(line, "y-axis"):
			d.parseYAxis(line)
		case strings.HasPrefix(line, "bar"):
			d.Series = append(d.Series, &Series{Kind: "bar", Values: parseFloats(line)})
		case strings.HasPrefix(line, "line"):
			d.Series = append(d.Series, &Series{Kind: "line", Values: parseFloats(line)})
		}
	}
	if !headerSeen {
		return nil, syntax.Errorf(1, 1, "expected 'xychart-beta' header")
	}
	if len(d.Series) == 0 {
		return nil, syntax.Errorf(1, 1, "xychart has no data series")
	}
	return d, nil
}

func (d *Diagram) parseYAxis(line string) {
	rest := strings.TrimSpace(line[len("y-axis"):])
	if lo, hi, ok := strings.Cut(rest, "-->"); ok {
		// optional leading "label" before the range
		loFields := strings.Fields(lo)
		if len(loFields) > 0 {
			if v, err := strconv.ParseFloat(loFields[len(loFields)-1], 64); err == nil {
				d.YMin = v
				if len(loFields) > 1 {
					d.YLabel = unquote(strings.Join(loFields[:len(loFields)-1], " "))
				}
			} else {
				d.YLabel = unquote(strings.TrimSpace(lo))
			}
		}
		if v, err := strconv.ParseFloat(strings.TrimSpace(hi), 64); err == nil {
			d.YMax = v
			d.HasYRange = true
		}
		return
	}
	d.YLabel = unquote(rest)
}

// Bounds returns the y range, deriving 0..max from the data when unset.
func (d *Diagram) Bounds() (lo, hi float64) {
	if d.HasYRange {
		return d.YMin, d.YMax
	}
	hi = 1
	for _, s := range d.Series {
		for _, v := range s.Values {
			if v > hi {
				hi = v
			}
		}
	}
	return 0, hi
}

func parseStrings(line string) []string {
	inner := bracketInner(line)
	if inner == "" {
		return nil
	}
	var out []string
	for _, f := range strings.Split(inner, ",") {
		if f = unquote(strings.TrimSpace(f)); f != "" {
			out = append(out, f)
		}
	}
	return out
}

func parseFloats(line string) []float64 {
	inner := bracketInner(line)
	var out []float64
	for _, f := range strings.Split(inner, ",") {
		if v, err := strconv.ParseFloat(strings.TrimSpace(f), 64); err == nil {
			out = append(out, v)
		}
	}
	return out
}

func bracketInner(s string) string {
	o := strings.IndexByte(s, '[')
	c := strings.LastIndexByte(s, ']')
	if o < 0 || c <= o {
		return ""
	}
	return s[o+1 : c]
}

func unquote(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
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
