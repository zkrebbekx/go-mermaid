package er

import (
	"strings"

	"github.com/Zac300/go-mermaid/internal/syntax"
)

// Parse builds a Diagram from ER diagram source.
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
			if firstWord(line) != "erDiagram" {
				return nil, syntax.Errorf(lineNo, 1, "expected 'erDiagram' header")
			}
			headerSeen = true
			continue
		}

		switch {
		case strings.HasSuffix(line, "{"):
			name := strings.TrimSpace(strings.TrimSuffix(line, "{"))
			e := d.ensureEntity(name)
			i = d.consumeBlock(e, lines, i+1)
		case hasRelation(line):
			if err := d.parseRelation(line, lineNo); err != nil {
				return nil, err
			}
		default:
			d.ensureEntity(line) // bare entity declaration
		}
	}

	if !headerSeen {
		return nil, syntax.Errorf(1, 1, "expected 'erDiagram' header")
	}
	return d, nil
}

func (d *Diagram) consumeBlock(e *Entity, lines []string, start int) int {
	for j := start; j < len(lines); j++ {
		line := strings.TrimSpace(stripComment(lines[j]))
		if line == "" {
			continue
		}
		if line == "}" {
			return j
		}
		e.Attributes = append(e.Attributes, line)
	}
	return len(lines) - 1
}

func (d *Diagram) parseRelation(line string, lineNo int) error {
	core := line
	label := ""
	if c := strings.IndexByte(line, ':'); c >= 0 {
		core = line[:c]
		label = strings.TrimSpace(line[c+1:])
	}
	fields := strings.Fields(core)
	opIdx := -1
	for i, f := range fields {
		if strings.Contains(f, "--") || strings.Contains(f, "..") {
			opIdx = i
			break
		}
	}
	if opIdx < 1 || opIdx >= len(fields)-1 {
		return syntax.Errorf(lineNo, 1, "invalid relationship")
	}
	from := fields[opIdx-1]
	to := fields[opIdx+1]
	op := fields[opIdx]

	lineSep := "--"
	if strings.Contains(op, "..") {
		lineSep = ".."
	}
	li := strings.Index(op, lineSep)
	if li < 0 {
		return syntax.Errorf(lineNo, 1, "invalid relationship operator")
	}
	leftCard := op[:li]
	rightCard := op[li+2:]

	d.ensureEntity(from)
	d.ensureEntity(to)
	d.Relationships = append(d.Relationships, &Relationship{
		From: from, To: to, Label: label,
		LeftCard:  cardLabel(leftCard),
		RightCard: cardLabel(rightCard),
		LeftKind:  cardKind(leftCard),
		RightKind: cardKind(rightCard),
		Dashed:    lineSep == "..",
	})
	return nil
}

// cardKind decodes a crow's-foot token into a cardinality kind.
func cardKind(tok string) Card {
	many := strings.ContainsAny(tok, "{}")
	zero := strings.Contains(tok, "o")
	switch {
	case many && zero:
		return CardZeroMany
	case many:
		return CardOneMany
	case zero:
		return CardZeroOne
	default:
		return CardOne
	}
}

// cardLabel maps a crow's-foot cardinality token to a readable label.
func cardLabel(tok string) string {
	many := strings.ContainsAny(tok, "{}")
	zero := strings.Contains(tok, "o")
	switch {
	case many && zero:
		return "0..N"
	case many:
		return "1..N"
	case zero:
		return "0..1"
	default:
		return "1"
	}
}

func hasRelation(line string) bool {
	core := line
	if c := strings.IndexByte(line, ':'); c >= 0 {
		core = line[:c]
	}
	for _, f := range strings.Fields(core) {
		if strings.Contains(f, "--") || strings.Contains(f, "..") {
			return true
		}
	}
	return false
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
