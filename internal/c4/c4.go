// Package c4 parses and renders Mermaid C4 diagrams (C4Context, C4Container,
// …) to SVG, reusing the shared layered layout engine.
//
// Syntax (subset):
//
//	C4Context
//	    title System Context
//	    Person(custA, "Customer", "A bank customer")
//	    System(sysA, "Internet Banking", "Lets customers ...")
//	    Rel(custA, sysA, "Uses")
package c4

import (
	"strings"

	"github.com/zkrebbekx/go-mermaid/internal/syntax"
)

// Element is a C4 element (Person, System, Container, Component, …).
type Element struct {
	ID    string
	Kind  string
	Label string
	Descr string
}

// Rel is a relationship between two elements.
type Rel struct {
	From  string
	To    string
	Label string
}

// Diagram is a parsed C4 diagram.
type Diagram struct {
	Title    string
	Elements []*Element
	Rels     []*Rel
}

func (d *Diagram) element(id string) *Element {
	for _, e := range d.Elements {
		if e.ID == id {
			return e
		}
	}
	return nil
}

// Parse builds a Diagram from C4 source.
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
			if !strings.HasPrefix(strings.ToLower(line), "c4") {
				return nil, syntax.Errorf(lineNo, 1, "expected a C4 header")
			}
			headerSeen = true
			continue
		}
		switch {
		case strings.HasPrefix(line, "title "):
			d.Title = strings.TrimSpace(line[len("title "):])
		case strings.HasPrefix(line, "Rel") && strings.Contains(line, "("):
			args := parseArgs(line)
			if len(args) >= 2 {
				r := &Rel{From: args[0], To: args[1]}
				if len(args) >= 3 {
					r.Label = args[2]
				}
				d.Rels = append(d.Rels, r)
			}
		case strings.HasPrefix(line, "Boundary"), strings.HasPrefix(line, "Enterprise_Boundary"),
			strings.HasPrefix(line, "System_Boundary"), strings.HasPrefix(line, "}"), line == "{":
			// boundaries are not yet drawn; skip
		case strings.Contains(line, "("):
			kind := line[:strings.IndexByte(line, '(')]
			args := parseArgs(line)
			if len(args) >= 2 {
				e := &Element{ID: args[0], Kind: strings.TrimSpace(kind), Label: args[1]}
				if len(args) >= 3 {
					e.Descr = args[2]
				}
				d.Elements = append(d.Elements, e)
			}
		}
	}
	if !headerSeen {
		return nil, syntax.Errorf(1, 1, "expected a C4 header")
	}
	return d, nil
}

// parseArgs returns the comma-separated arguments inside the first (...) on the
// line, trimming quotes from each.
func parseArgs(line string) []string {
	open := strings.IndexByte(line, '(')
	closeIdx := strings.LastIndexByte(line, ')')
	if open < 0 || closeIdx <= open {
		return nil
	}
	inner := line[open+1 : closeIdx]

	var args []string
	var cur strings.Builder
	inQuote := false
	for _, r := range inner {
		switch {
		case r == '"':
			inQuote = !inQuote
		case r == ',' && !inQuote:
			args = append(args, strings.TrimSpace(cur.String()))
			cur.Reset()
		default:
			cur.WriteRune(r)
		}
	}
	if strings.TrimSpace(cur.String()) != "" {
		args = append(args, strings.TrimSpace(cur.String()))
	}
	return args
}

func stripComment(s string) string {
	if i := strings.Index(s, "%%"); i >= 0 {
		return s[:i]
	}
	return s
}
