// Package sankey parses and renders Mermaid sankey-beta diagrams to SVG as
// proportional flow bands between nodes arranged in columns.
//
// Syntax (CSV, one flow per line):
//
//	sankey-beta
//	A,B,10
//	B,C,5
package sankey

import (
	"strconv"
	"strings"

	"github.com/zkrebbekx/go-mermaid/internal/syntax"
)

// Flow is a weighted link from a source node to a target node.
type Flow struct {
	Source string
	Target string
	Value  float64
}

// Diagram is a parsed sankey diagram.
type Diagram struct {
	Flows []*Flow
	Nodes []string // first-seen order
	seen  map[string]bool
}

// Parse builds a Diagram from sankey-beta source.
func Parse(src string) (*Diagram, error) {
	d := &Diagram{seen: map[string]bool{}}
	headerSeen := false
	for i, raw := range strings.Split(src, "\n") {
		lineNo := i + 1
		line := strings.TrimSpace(stripComment(raw))
		if line == "" {
			continue
		}
		if !headerSeen {
			if w := strings.ToLower(firstWord(line)); w != "sankey-beta" && w != "sankey" {
				return nil, syntax.Errorf(lineNo, 1, "expected 'sankey-beta' header")
			}
			headerSeen = true
			continue
		}
		fields := splitCSV(line)
		if len(fields) < 3 {
			return nil, syntax.Errorf(lineNo, 1, "expected 'source,target,value'")
		}
		v, err := strconv.ParseFloat(strings.TrimSpace(fields[2]), 64)
		if err != nil {
			return nil, syntax.Errorf(lineNo, 1, "invalid flow value")
		}
		s, tgt := strings.TrimSpace(fields[0]), strings.TrimSpace(fields[1])
		d.addNode(s)
		d.addNode(tgt)
		d.Flows = append(d.Flows, &Flow{Source: s, Target: tgt, Value: v})
	}
	if !headerSeen {
		return nil, syntax.Errorf(1, 1, "expected 'sankey-beta' header")
	}
	return d, nil
}

func (d *Diagram) addNode(name string) {
	if !d.seen[name] {
		d.seen[name] = true
		d.Nodes = append(d.Nodes, name)
	}
}

// splitCSV splits a line on commas, honoring double-quoted fields.
func splitCSV(line string) []string {
	var fields []string
	var cur strings.Builder
	inQuote := false
	for _, r := range line {
		switch {
		case r == '"':
			inQuote = !inQuote
		case r == ',' && !inQuote:
			fields = append(fields, cur.String())
			cur.Reset()
		default:
			cur.WriteRune(r)
		}
	}
	fields = append(fields, cur.String())
	return fields
}

func firstWord(s string) string {
	if i := strings.IndexAny(s, " \t,"); i >= 0 {
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
