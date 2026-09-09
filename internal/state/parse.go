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
	// scopes is the stack of enclosing composite state IDs. The empty string
	// at the bottom is the diagram itself.
	scopes := []string{""}
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

		parent := scopes[len(scopes)-1]

		switch {
		case line == "}":
			if len(scopes) > 1 {
				scopes = scopes[:len(scopes)-1]
			}

		case isCompositeOpen(line):
			id, label := declaredState(strings.TrimSuffix(line, "{"))
			s := d.ensureState(id)
			s.Label = label
			s.Parent = parent
			d.addMember(parent, id)
			if d.composite(id) == nil {
				d.Composites = append(d.Composites, &Composite{ID: id, Label: label})
			}
			scopes = append(scopes, id)

		case isDirection(line):
			d.Direction = strings.ToUpper(strings.TrimSpace(line[len("direction"):]))

		case isNoteOpen(line):
			note, body, ok := parseNote(line)
			if !ok {
				break
			}
			if body != "" {
				note.Text = body
				d.Notes = append(d.Notes, note)
				break
			}
			// A note with no inline text runs until `end note`.
			var text []string
			for i+1 < len(lines) {
				i++
				n := strings.TrimSpace(stripComment(lines[i]))
				if strings.EqualFold(n, "end note") {
					break
				}
				text = append(text, n)
			}
			note.Text = strings.Join(text, " ")
			d.Notes = append(d.Notes, note)

		case line == "--":
			// A concurrency separator between regions of a composite state.
			// The regions are laid out together, so nothing is drawn for it.

		// A transition has "-->"; but a description like `S : text --> more`
		// also contains it, so only treat the line as a transition when the
		// arrow comes before any ':' (which would start a description).
		case isTransition(line):
			d.parseTransition(line, parent)

		case isStereotype(line):
			id, kind := parseStereotype(line)
			s := d.ensureState(id)
			s.Kind = kind
			s.Label = ""
			s.Parent = parent
			d.addMember(parent, id)

		case strings.Contains(line, ":"):
			d.parseDescription(line, parent)

		case strings.HasPrefix(line, "state "):
			id, label := declaredState(line)
			s := d.ensureState(id)
			s.Label = label
			s.Parent = parent
			d.addMember(parent, id)

		default:
			s := d.ensureState(line) // bare state declaration
			s.Parent = parent
			d.addMember(parent, line)
		}
	}

	if !headerSeen {
		return nil, syntax.Errorf(1, 1, "expected 'stateDiagram-v2' header")
	}
	return d, nil
}

// addMember records that id belongs to the composite parent. Top-level states
// have no parent and are not recorded.
func (d *Diagram) addMember(parent, id string) {
	if parent == "" {
		return
	}
	c := d.composite(parent)
	if c == nil {
		return
	}
	for _, m := range c.Members {
		if m == id {
			return
		}
	}
	c.Members = append(c.Members, id)
}

func (d *Diagram) parseTransition(line, parent string) {
	idx := strings.Index(line, "-->")
	from := strings.TrimSpace(line[:idx])
	rest := line[idx+3:]
	label := ""
	if c := strings.IndexByte(rest, ':'); c >= 0 {
		label = strings.TrimSpace(rest[c+1:])
		rest = rest[:c]
	}
	to := strings.TrimSpace(rest)
	fromID := d.resolve(from, parent, false)
	toID := d.resolve(to, parent, true)
	d.Transitions = append(d.Transitions, &Transition{From: fromID, To: toID, Label: label})
}

func (d *Diagram) parseDescription(line, parent string) {
	name, desc, _ := strings.Cut(line, ":")
	id, _ := declaredState(strings.TrimSpace(name))
	s := d.ensureState(id)
	s.Label = strings.TrimSpace(desc)
	if s.Parent == "" {
		s.Parent = parent
	}
	d.addMember(parent, id)
}

// resolve maps a token to a state ID, turning [*] into the start or end
// pseudostate of the enclosing scope depending on whether it is a target.
func (d *Diagram) resolve(token, parent string, asTarget bool) string {
	token = strings.TrimSpace(token)
	if token == "[*]" {
		p := d.ensurePseudo(parent, asTarget)
		// A composite's own entry and exit belong inside its box.
		d.addMember(parent, p.ID)
		return p.ID
	}
	id, _ := declaredState(token)
	s := d.ensureState(id)
	if s.Parent == "" {
		s.Parent = parent
	}
	d.addMember(parent, id)
	return s.ID
}

// declaredState splits a state declaration into its ID and its display label.
// It understands the bare form `S`, the `state S` prefix, and the aliased
// form `state "Some description" as S`, where the ID is the alias and the
// quoted text is the label.
func declaredState(s string) (id, label string) {
	s = strings.TrimSpace(s)
	s = strings.TrimSpace(strings.TrimPrefix(s, "state "))
	if i := strings.Index(s, " as "); i >= 0 {
		label = strings.Trim(strings.TrimSpace(s[:i]), `"`)
		id = strings.TrimSpace(s[i+4:])
		return id, label
	}
	s = strings.TrimSpace(s)
	return s, s
}

// isCompositeOpen reports whether the line opens a composite state body.
func isCompositeOpen(line string) bool {
	return strings.HasSuffix(line, "{") && strings.HasPrefix(line, "state ")
}

// isDirection reports whether the line sets the layout direction.
func isDirection(line string) bool {
	return strings.HasPrefix(line, "direction ") || line == "direction"
}

// isNoteOpen reports whether the line starts a note.
func isNoteOpen(line string) bool {
	return strings.HasPrefix(strings.ToLower(line), "note ")
}

// parseNote reads `note right of S : text` or `note left of S`. The returned
// body is empty when the note continues on following lines until `end note`.
func parseNote(line string) (n *Note, body string, ok bool) {
	rest := strings.TrimSpace(line[len("note"):])
	side := SideRight
	switch {
	case strings.HasPrefix(strings.ToLower(rest), "left of "):
		side = SideLeft
		rest = rest[len("left of "):]
	case strings.HasPrefix(strings.ToLower(rest), "right of "):
		rest = rest[len("right of "):]
	default:
		return nil, "", false
	}
	target, text, _ := strings.Cut(rest, ":")
	return &Note{Target: strings.TrimSpace(target), Side: side}, strings.TrimSpace(text), true
}

// isStereotype reports whether the line declares a pseudostate such as
// `state fork_state <<fork>>`.
func isStereotype(line string) bool {
	return strings.HasPrefix(line, "state ") &&
		strings.Contains(line, "<<") && strings.HasSuffix(line, ">>")
}

// parseStereotype reads the ID and kind out of `state ID <<kind>>`.
func parseStereotype(line string) (id string, kind Kind) {
	rest := strings.TrimSpace(strings.TrimPrefix(line, "state "))
	i := strings.Index(rest, "<<")
	id = strings.TrimSpace(rest[:i])
	switch strings.ToLower(strings.TrimSuffix(strings.TrimSpace(rest[i+2:]), ">>")) {
	case "fork":
		kind = KindFork
	case "join":
		kind = KindJoin
	case "choice":
		kind = KindChoice
	}
	return id, kind
}

// isTransition reports whether a line is a transition (`A --> B`) rather than a
// description (`A : text`). When both delimiters appear, the one that comes
// first wins, so `S : note about --> arrows` stays a description.
func isTransition(line string) bool {
	a := strings.Index(line, "-->")
	if a < 0 {
		return false
	}
	c := strings.Index(line, ":")
	return c < 0 || a < c
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
