package state

import (
	"strings"

	"github.com/zkrebbekx/go-mermaid/internal/syntax"
)

// Parse builds a Diagram from state diagram source.
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
			w := strings.ToLower(firstWord(line))
			if w != "statediagram-v2" && w != "statediagram" {
				return nil, syntax.Errorf(lineNo, 1, "expected 'stateDiagram-v2' header")
			}
			headerSeen = true
			continue
		}

		switch {
		case strings.HasPrefix(line, "state ") && strings.HasSuffix(line, "{"):
			// Composite state: register it, skip nested body for now.
			name := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(line, "state "), "{"))
			d.ensureState(stateName(name))
			i = skipBlock(lines, i+1)
		case strings.Contains(line, "-->"):
			d.parseTransition(line)
		case strings.Contains(line, ":"):
			d.parseDescription(line)
		case strings.HasPrefix(line, "state "):
			d.ensureState(stateName(strings.TrimSpace(strings.TrimPrefix(line, "state "))))
		default:
			d.ensureState(line) // bare state declaration
		}
	}

	if !headerSeen {
		return nil, syntax.Errorf(1, 1, "expected 'stateDiagram-v2' header")
	}
	return d, nil
}

func (d *Diagram) parseTransition(line string) {
	idx := strings.Index(line, "-->")
	from := strings.TrimSpace(line[:idx])
	rest := line[idx+3:]
	label := ""
	if c := strings.IndexByte(rest, ':'); c >= 0 {
		label = strings.TrimSpace(rest[c+1:])
		rest = rest[:c]
	}
	to := strings.TrimSpace(rest)
	fromID := d.resolve(from, false)
	toID := d.resolve(to, true)
	d.Transitions = append(d.Transitions, &Transition{From: fromID, To: toID, Label: label})
}

func (d *Diagram) parseDescription(line string) {
	name, desc, _ := strings.Cut(line, ":")
	s := d.ensureState(stateName(strings.TrimSpace(name)))
	s.Label = strings.TrimSpace(desc)
}

// resolve maps a token to a state ID, turning [*] into the start or end
// pseudostate depending on whether it is a transition target.
func (d *Diagram) resolve(token string, asTarget bool) string {
	token = stateName(token)
	if token == "[*]" {
		return d.ensurePseudo(asTarget).ID
	}
	return d.ensureState(token).ID
}

// stateName strips an "as alias" suffix and trims whitespace.
func stateName(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.Index(s, " as "); i >= 0 {
		return strings.TrimSpace(s[:i])
	}
	return s
}

func skipBlock(lines []string, start int) int {
	for j := start; j < len(lines); j++ {
		if strings.TrimSpace(stripComment(lines[j])) == "}" {
			return j
		}
	}
	return len(lines) - 1
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
