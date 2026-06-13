// Package pie parses and renders Mermaid pie charts to SVG.
//
// Syntax:
//
//	pie showData title Pets adopted
//	    "Dogs" : 386
//	    "Cats" : 85
package pie

import (
	"strconv"
	"strings"

	"github.com/Zac300/go-mermaid/internal/syntax"
)

// Slice is one labeled wedge of the pie.
type Slice struct {
	Label string
	Value float64
}

// Diagram is a parsed pie chart.
type Diagram struct {
	Title  string
	Slices []Slice
}

// Total returns the sum of all slice values.
func (d *Diagram) Total() float64 {
	var t float64
	for _, s := range d.Slices {
		t += s.Value
	}
	return t
}

// Parse builds a Diagram from pie chart source.
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
			if firstWord(line) != "pie" {
				return nil, syntax.Errorf(lineNo, 1, "expected 'pie' header")
			}
			rest := strings.TrimSpace(strings.TrimPrefix(line, "pie"))
			rest = strings.TrimSpace(strings.TrimPrefix(rest, "showData"))
			if t, ok := strings.CutPrefix(rest, "title "); ok {
				d.Title = strings.TrimSpace(t)
			}
			headerSeen = true
			continue
		}
		if err := d.parseSlice(line, lineNo); err != nil {
			return nil, err
		}
	}
	if !headerSeen {
		return nil, syntax.Errorf(1, 1, "expected 'pie' header")
	}
	return d, nil
}

func (d *Diagram) parseSlice(line string, lineNo int) error {
	idx := strings.LastIndexByte(line, ':')
	if idx < 0 {
		return syntax.Errorf(lineNo, 1, "expected \"label\" : value")
	}
	label := strings.Trim(strings.TrimSpace(line[:idx]), `"`)
	v, err := strconv.ParseFloat(strings.TrimSpace(line[idx+1:]), 64)
	if err != nil {
		return syntax.Errorf(lineNo, 1, "invalid slice value")
	}
	if v < 0 {
		return syntax.Errorf(lineNo, 1, "slice value must not be negative")
	}
	d.Slices = append(d.Slices, Slice{Label: label, Value: v})
	return nil
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
