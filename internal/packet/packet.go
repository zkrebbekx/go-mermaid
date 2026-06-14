// Package packet parses and renders Mermaid packet-beta diagrams to SVG as a
// bit/byte field table.
//
// Syntax:
//
//	packet-beta
//	0-15: "Source Port"
//	16-31: "Destination Port"
//	32-63: "Sequence Number"
package packet

import (
	"strconv"
	"strings"

	"github.com/zkrebbekx/go-mermaid/internal/syntax"
)

// Field is a contiguous range of bits with a label.
type Field struct {
	Start int
	End   int
	Label string
}

// Diagram is a parsed packet diagram.
type Diagram struct {
	Fields []*Field
}

// Parse builds a Diagram from packet-beta source.
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
			if w := strings.ToLower(firstWord(line)); w != "packet-beta" && w != "packet" {
				return nil, syntax.Errorf(lineNo, 1, "expected 'packet-beta' header")
			}
			headerSeen = true
			continue
		}
		f, err := parseField(line, lineNo)
		if err != nil {
			return nil, err
		}
		d.Fields = append(d.Fields, f)
	}
	if !headerSeen {
		return nil, syntax.Errorf(1, 1, "expected 'packet-beta' header")
	}
	if len(d.Fields) == 0 {
		return nil, syntax.Errorf(1, 1, "packet has no fields")
	}
	return d, nil
}

func parseField(line string, lineNo int) (*Field, error) {
	rng, label, ok := strings.Cut(line, ":")
	if !ok {
		return nil, syntax.Errorf(lineNo, 1, "expected 'start-end: label'")
	}
	rng = strings.TrimSpace(rng)
	start, end := 0, 0
	if lo, hi, isRange := strings.Cut(rng, "-"); isRange {
		var e1, e2 error
		start, e1 = strconv.Atoi(strings.TrimSpace(lo))
		end, e2 = strconv.Atoi(strings.TrimSpace(hi))
		if e1 != nil || e2 != nil {
			return nil, syntax.Errorf(lineNo, 1, "invalid bit range")
		}
	} else {
		v, err := strconv.Atoi(strings.TrimSpace(rng))
		if err != nil {
			return nil, syntax.Errorf(lineNo, 1, "invalid bit index")
		}
		start, end = v, v
	}
	if end < start {
		return nil, syntax.Errorf(lineNo, 1, "bit range end before start")
	}
	return &Field{Start: start, End: end, Label: strings.Trim(strings.TrimSpace(label), `"`)}, nil
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
