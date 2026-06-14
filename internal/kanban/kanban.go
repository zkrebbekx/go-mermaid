// Package kanban parses and renders Mermaid kanban diagrams to SVG as columns
// of cards, derived from indentation.
//
// Syntax (subset):
//
//	kanban
//	  Todo
//	    [Task 1]
//	    [Task 2]
//	  Done
//	    [Task 3]
package kanban

import (
	"strings"

	"github.com/zkrebbekx/go-mermaid/internal/syntax"
)

// Card is a single item within a column.
type Card struct {
	Text string
}

// Column is a labeled list of cards.
type Column struct {
	Title string
	Cards []*Card
}

// Diagram is a parsed kanban board.
type Diagram struct {
	Columns []*Column
}

// Parse builds a Diagram from kanban source using indentation: the shallowest
// item lines are column titles, deeper lines are cards.
func Parse(src string) (*Diagram, error) {
	type item struct {
		indent int
		text   string
	}
	var items []item

	headerSeen := false
	for i, raw := range strings.Split(src, "\n") {
		lineNo := i + 1
		if strings.TrimSpace(stripComment(raw)) == "" {
			continue
		}
		if !headerSeen {
			if strings.TrimSpace(raw) != "kanban" {
				return nil, syntax.Errorf(lineNo, 1, "expected 'kanban' header")
			}
			headerSeen = true
			continue
		}
		items = append(items, item{indent: leadingSpaces(raw), text: strings.TrimSpace(stripComment(raw))})
	}
	if !headerSeen {
		return nil, syntax.Errorf(1, 1, "expected 'kanban' header")
	}

	minIndent := -1
	for _, it := range items {
		if minIndent < 0 || it.indent < minIndent {
			minIndent = it.indent
		}
	}

	d := &Diagram{}
	var cur *Column
	for _, it := range items {
		if it.indent == minIndent {
			cur = &Column{Title: it.text}
			d.Columns = append(d.Columns, cur)
			continue
		}
		if cur == nil {
			cur = &Column{}
			d.Columns = append(d.Columns, cur)
		}
		cur.Cards = append(cur.Cards, &Card{Text: cardText(it.text)})
	}
	return d, nil
}

// cardText strips an "id[text]" wrapper down to its text.
func cardText(s string) string {
	if o := strings.IndexByte(s, '['); o >= 0 && strings.HasSuffix(s, "]") {
		return strings.Trim(strings.TrimSuffix(s[o+1:], "]"), `"`)
	}
	return s
}

func leadingSpaces(s string) int {
	n := 0
	for _, r := range s {
		switch r {
		case ' ':
			n++
		case '\t':
			n += 2
		default:
			return n
		}
	}
	return n
}

func stripComment(s string) string {
	if i := strings.Index(s, "%%"); i >= 0 {
		return s[:i]
	}
	return s
}
