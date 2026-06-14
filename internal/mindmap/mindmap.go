// Package mindmap parses and renders Mermaid mindmaps to SVG using an
// indentation-based hierarchy drawn as a left-to-right tree.
//
// Syntax:
//
//	mindmap
//	  root((Root))
//	    Origins
//	      Long history
//	    Tools
package mindmap

import (
	"strings"

	"github.com/zkrebbekx/go-mermaid/internal/syntax"
)

// Node is a mindmap node with children.
type Node struct {
	Text     string
	Depth    int
	Children []*Node

	X float64
	Y float64
}

// Diagram is a parsed mindmap.
type Diagram struct {
	Root *Node
}

// Parse builds a Diagram from mindmap source using leading-space indentation.
func Parse(src string) (*Diagram, error) {
	d := &Diagram{}
	type frame struct {
		node   *Node
		indent int
	}
	var stack []frame

	headerSeen := false
	for i, raw := range strings.Split(src, "\n") {
		lineNo := i + 1
		if strings.TrimSpace(stripComment(raw)) == "" {
			continue
		}
		if !headerSeen {
			if strings.TrimSpace(raw) != "mindmap" {
				return nil, syntax.Errorf(lineNo, 1, "expected 'mindmap' header")
			}
			headerSeen = true
			continue
		}
		indent := leadingSpaces(raw)
		text := cleanText(strings.TrimSpace(stripComment(raw)))
		n := &Node{Text: text}

		for len(stack) > 0 && stack[len(stack)-1].indent >= indent {
			stack = stack[:len(stack)-1]
		}
		if len(stack) == 0 {
			if d.Root == nil {
				d.Root = n
			} else {
				// Additional roots attach under the first root.
				d.Root.Children = append(d.Root.Children, n)
			}
		} else {
			parent := stack[len(stack)-1].node
			n.Depth = parent.Depth + 1
			parent.Children = append(parent.Children, n)
		}
		stack = append(stack, frame{node: n, indent: indent})
	}
	if !headerSeen {
		return nil, syntax.Errorf(1, 1, "expected 'mindmap' header")
	}
	if d.Root == nil {
		return nil, syntax.Errorf(1, 1, "mindmap has no root node")
	}
	return d, nil
}

// cleanText extracts a node's display text from an optional shape, e.g.
// "root((Ideas))" -> "Ideas", "id[Text]" -> "Text", "Plain" -> "Plain".
func cleanText(s string) string {
	for _, pair := range [][2]string{{"((", "))"}, {"{{", "}}"}, {"([", "])"}, {"[", "]"}, {"(", ")"}, {"{", "}"}} {
		open, closer := pair[0], pair[1]
		if !strings.HasSuffix(s, closer) {
			continue
		}
		i := strings.Index(s, open)
		if i >= 0 && i+len(open) <= len(s)-len(closer) {
			return strings.TrimSpace(s[i+len(open) : len(s)-len(closer)])
		}
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
