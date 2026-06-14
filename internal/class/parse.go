package class

import (
	"strings"

	"github.com/zkrebbekx/go-mermaid/internal/syntax"
)

// relOps are class relationship operators, ordered so the longest match at a
// position wins. Every operator contains "--" or ".." as its line.
var relOps = []string{
	"<|--", "--|>", "<|..", "..|>",
	"*--", "--*", "o--", "--o",
	"-->", "<--", "..>", "<..",
	"--", "..",
}

// Parse builds a Diagram from class diagram source.
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
			if firstWord(line) != "classDiagram" {
				return nil, syntax.Errorf(lineNo, 1, "expected 'classDiagram' header")
			}
			headerSeen = true
			continue
		}

		switch {
		case strings.HasPrefix(line, "class ") && strings.HasSuffix(line, "{"):
			name := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(line, "class "), "{"))
			c := d.ensureClass(className(name))
			i = d.consumeBlock(c, lines, i+1) // advance past the block
		case strings.HasPrefix(line, "class "):
			d.ensureClass(className(strings.TrimSpace(strings.TrimPrefix(line, "class "))))
		case relIndex(line) >= 0:
			if err := d.parseRelation(line, lineNo); err != nil {
				return nil, err
			}
		case strings.Contains(line, ":"):
			d.parseShorthandMember(line)
		default:
			return nil, syntax.Errorf(lineNo, 1, "unrecognized statement %q", line)
		}
	}

	if !headerSeen {
		return nil, syntax.Errorf(1, 1, "expected 'classDiagram' header")
	}
	return d, nil
}

// consumeBlock reads member lines until a closing "}" and returns the index of
// that closing line so the caller's loop continues after it.
func (d *Diagram) consumeBlock(c *Class, lines []string, start int) int {
	for j := start; j < len(lines); j++ {
		line := strings.TrimSpace(stripComment(lines[j]))
		if line == "" {
			continue
		}
		if line == "}" {
			return j
		}
		c.addMember(line)
	}
	return len(lines) - 1
}

func (d *Diagram) parseShorthandMember(line string) {
	name, member, _ := strings.Cut(line, ":")
	c := d.ensureClass(className(strings.TrimSpace(name)))
	c.addMember(strings.TrimSpace(member))
}

func (d *Diagram) parseRelation(line string, lineNo int) error {
	core := line
	label := ""
	if c := strings.Index(line, ":"); c >= 0 {
		core = line[:c]
		label = strings.TrimSpace(line[c+1:])
	}
	idx, op := findRelOp(core)
	if idx < 0 {
		return syntax.Errorf(lineNo, 1, "invalid relationship")
	}
	from := className(strings.TrimSpace(core[:idx]))
	to := className(strings.TrimSpace(core[idx+len(op):]))
	if from == "" || to == "" {
		return syntax.Errorf(lineNo, 1, "relationship needs two classes")
	}
	d.ensureClass(from)
	d.ensureClass(to)
	d.Relations = append(d.Relations, &Relation{
		From: from, To: to, Label: label,
		Dashed: strings.Contains(op, ".."),
		Left:   leftHead(op),
		Right:  rightHead(op),
	})
	return nil
}

// addMember classifies a member as a method (contains "(") or attribute.
func (c *Class) addMember(m string) {
	if m == "" {
		return
	}
	if strings.Contains(m, "(") {
		c.Methods = append(c.Methods, m)
	} else {
		c.Attributes = append(c.Attributes, m)
	}
}

// className strips a generic suffix like "List~T~" down to "List".
func className(s string) string {
	if i := strings.IndexByte(s, '~'); i >= 0 {
		return strings.TrimSpace(s[:i])
	}
	return strings.TrimSpace(s)
}

// relIndex reports the operator index in a line (ignoring any label), or -1.
func relIndex(line string) int {
	core := line
	if c := strings.Index(line, ":"); c >= 0 {
		core = line[:c]
	}
	idx, _ := findRelOp(core)
	return idx
}

// findRelOp returns the earliest (longest at that position) operator in s.
func findRelOp(s string) (int, string) {
	for i := 0; i < len(s); i++ {
		for _, op := range relOps {
			if strings.HasPrefix(s[i:], op) {
				return i, op
			}
		}
	}
	return -1, ""
}

func leftHead(op string) headKind {
	switch {
	case strings.HasPrefix(op, "<|"):
		return headTriangle
	case strings.HasPrefix(op, "*"):
		return headDiamondFilled
	case strings.HasPrefix(op, "o"):
		return headDiamondHollow
	case strings.HasPrefix(op, "<"):
		return headArrow
	}
	return headNone
}

func rightHead(op string) headKind {
	switch {
	case strings.HasSuffix(op, "|>"):
		return headTriangle
	case strings.HasSuffix(op, "*"):
		return headDiamondFilled
	case strings.HasSuffix(op, "o"):
		return headDiamondHollow
	case strings.HasSuffix(op, ">"):
		return headArrow
	}
	return headNone
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
