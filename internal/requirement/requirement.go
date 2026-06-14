// Package requirement parses and renders Mermaid requirement diagrams to SVG,
// reusing the shared layered layout engine.
//
// Syntax (subset):
//
//	requirementDiagram
//	    requirement test_req {
//	        id: 1
//	        text: the test text
//	        risk: high
//	    }
//	    element test_entity {
//	        type: simulation
//	    }
//	    test_entity - satisfies -> test_req
package requirement

import (
	"strings"

	"github.com/zkrebbekx/go-mermaid/internal/syntax"
)

// Node is a requirement or an element.
type Node struct {
	ID        string
	Kind      string // requirement type or "element"
	Fields    map[string]string
	IsElement bool
}

// Rel is a typed relationship (satisfies, traces, derives, …).
type Rel struct {
	From string
	To   string
	Type string
}

// Diagram is a parsed requirement diagram.
type Diagram struct {
	Nodes []*Node
	Rels  []*Rel
}

func (d *Diagram) node(id string) *Node {
	for _, n := range d.Nodes {
		if n.ID == id {
			return n
		}
	}
	return nil
}

// Parse builds a Diagram from requirement diagram source.
func Parse(src string) (*Diagram, error) {
	d := &Diagram{}
	lines := strings.Split(src, "\n")

	headerSeen := false
	for i := 0; i < len(lines); i++ {
		lineNo := i + 1
		line := strings.TrimSpace(stripComment(lines[i]))
		if line == "" {
			continue
		}
		if !headerSeen {
			if firstWord(line) != "requirementDiagram" {
				return nil, syntax.Errorf(lineNo, 1, "expected 'requirementDiagram' header")
			}
			headerSeen = true
			continue
		}
		switch {
		case strings.HasSuffix(line, "{"):
			kind := firstWord(line)
			name := strings.TrimSpace(strings.TrimSuffix(line[len(kind):], "{"))
			n := &Node{ID: name, Kind: kind, Fields: map[string]string{}, IsElement: kind == "element"}
			i = d.consumeBlock(n, lines, i+1)
			d.Nodes = append(d.Nodes, n)
		case strings.Contains(line, "->"):
			d.parseRel(line)
		}
	}
	if !headerSeen {
		return nil, syntax.Errorf(1, 1, "expected 'requirementDiagram' header")
	}
	return d, nil
}

func (d *Diagram) consumeBlock(n *Node, lines []string, start int) int {
	for j := start; j < len(lines); j++ {
		line := strings.TrimSpace(stripComment(lines[j]))
		if line == "" {
			continue
		}
		if line == "}" {
			return j
		}
		if k, v, ok := strings.Cut(line, ":"); ok {
			n.Fields[strings.TrimSpace(k)] = strings.TrimSpace(v)
		}
	}
	return len(lines) - 1
}

func (d *Diagram) parseRel(line string) {
	idx := strings.Index(line, "->")
	to := strings.TrimSpace(line[idx+2:])
	lhs := strings.TrimSpace(line[:idx])
	from, typ := lhs, ""
	if di := strings.LastIndex(lhs, "-"); di >= 0 {
		from = strings.TrimSpace(lhs[:di])
		typ = strings.TrimSpace(lhs[di+1:])
	}
	if from == "" || to == "" {
		return
	}
	d.ensure(from)
	d.ensure(to)
	d.Rels = append(d.Rels, &Rel{From: from, To: to, Type: typ})
}

func (d *Diagram) ensure(id string) {
	if d.node(id) == nil {
		d.Nodes = append(d.Nodes, &Node{ID: id, Fields: map[string]string{}})
	}
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
