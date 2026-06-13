// Package block parses and renders Mermaid block-beta diagrams to SVG as a
// column grid of labeled blocks.
//
// Syntax (subset):
//
//	block-beta
//	    columns 3
//	    a b c
//	    d["wide"]:2 e
package block

import (
	"strconv"
	"strings"

	"github.com/Zac300/go-mermaid/internal/syntax"
)

// Block is one cell in the grid.
type Block struct {
	ID    string
	Label string
	Span  int
}

// Diagram is a parsed block diagram.
type Diagram struct {
	Columns int
	Rows    [][]*Block
}

// Parse builds a Diagram from block-beta source.
func Parse(src string) (*Diagram, error) {
	d := &Diagram{Columns: 0}
	headerSeen := false
	for i, raw := range strings.Split(src, "\n") {
		lineNo := i + 1
		line := strings.TrimSpace(stripComment(raw))
		if line == "" {
			continue
		}
		if !headerSeen {
			if w := strings.TrimRight(strings.ToLower(firstWord(line)), "-beta"); w != "block" {
				return nil, syntax.Errorf(lineNo, 1, "expected 'block-beta' header")
			}
			headerSeen = true
			continue
		}
		if strings.HasPrefix(line, "columns ") {
			if n, err := strconv.Atoi(strings.TrimSpace(line[len("columns "):])); err == nil {
				d.Columns = n
			}
			continue
		}
		row := parseRow(line)
		if len(row) > 0 {
			d.Rows = append(d.Rows, row)
		}
	}
	if !headerSeen {
		return nil, syntax.Errorf(1, 1, "expected 'block-beta' header")
	}
	if d.Columns <= 0 {
		d.Columns = widestRow(d.Rows)
	}
	if d.Columns <= 0 {
		d.Columns = 1
	}
	return d, nil
}

// parseRow tokenizes a row, keeping bracketed labels intact.
func parseRow(line string) []*Block {
	var tokens []string
	var cur strings.Builder
	depth := 0
	for _, r := range line {
		switch {
		case r == '[':
			depth++
			cur.WriteRune(r)
		case r == ']':
			if depth > 0 {
				depth--
			}
			cur.WriteRune(r)
		case (r == ' ' || r == '\t') && depth == 0:
			if cur.Len() > 0 {
				tokens = append(tokens, cur.String())
				cur.Reset()
			}
		default:
			cur.WriteRune(r)
		}
	}
	if cur.Len() > 0 {
		tokens = append(tokens, cur.String())
	}

	var blocks []*Block
	for _, t := range tokens {
		blocks = append(blocks, parseBlock(t))
	}
	return blocks
}

// parseBlock parses id["label"]:span into a Block.
func parseBlock(t string) *Block {
	b := &Block{Span: 1}
	// trailing :span
	if i := strings.LastIndexByte(t, ':'); i >= 0 {
		if n, err := strconv.Atoi(t[i+1:]); err == nil {
			b.Span = n
			t = t[:i]
		}
	}
	if o := strings.IndexByte(t, '['); o >= 0 && strings.HasSuffix(t, "]") {
		b.ID = strings.TrimSpace(t[:o])
		b.Label = strings.Trim(strings.TrimSuffix(t[o+1:], "]"), `"`)
	} else {
		b.ID = t
		b.Label = t
	}
	if b.Span < 1 {
		b.Span = 1
	}
	return b
}

func widestRow(rows [][]*Block) int {
	widest := 0
	for _, r := range rows {
		span := 0
		for _, b := range r {
			span += b.Span
		}
		if span > widest {
			widest = span
		}
	}
	return widest
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
