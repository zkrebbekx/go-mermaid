package parser

import (
	"strings"

	"github.com/Zac300/go-mermaid/internal/domain"
)

// classAssign records a pending "apply class name to these node ids".
type classAssign struct {
	ids  []string
	name string
}

// Preprocess pulls flowchart styling directives (classDef, class, style, and
// inline :::class) out of the source. It returns the source with those
// directives removed (so the lexer never sees their CSS-like payloads) and a
// map of node ID to resolved Style. Directive order does not matter.
func Preprocess(src string) (string, map[string]*domain.Style, map[string]string) {
	classDefs := map[string]*domain.Style{}
	styles := map[string]*domain.Style{}
	links := map[string]string{}

	var pending []classAssign

	var kept []string
	for _, line := range strings.Split(src, "\n") {
		t := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(t, "classDef "):
			name, st := parseClassDef(t)
			if name != "" {
				classDefs[name] = st
			}
		case strings.HasPrefix(t, "class "):
			rest := strings.Fields(strings.TrimSpace(t[len("class "):]))
			if len(rest) >= 2 {
				ids := strings.Split(rest[0], ",")
				pending = append(pending, classAssign{ids: ids, name: rest[1]})
			}
		case strings.HasPrefix(t, "style "):
			id, st := parseStyleStmt(t)
			if id != "" {
				mergeStyle(styles, id, st)
			}
		case strings.HasPrefix(t, "click "):
			if id, url := parseClick(t); id != "" && url != "" {
				links[id] = url
			}
		default:
			kept = append(kept, stripInline(line, &pending))
		}
	}

	for _, a := range pending {
		st, ok := classDefs[a.name]
		if !ok {
			continue
		}
		for _, id := range a.ids {
			mergeStyle(styles, strings.TrimSpace(id), st)
		}
	}
	return strings.Join(kept, "\n"), styles, links
}

// parseClick handles "click ID href "URL"" and "click ID "URL"".
func parseClick(line string) (id, url string) {
	rest := strings.Fields(strings.TrimSpace(line[len("click"):]))
	if len(rest) < 1 {
		return "", ""
	}
	id = rest[0]
	if i := strings.IndexByte(line, '"'); i >= 0 {
		if j := strings.IndexByte(line[i+1:], '"'); j >= 0 {
			url = line[i+1 : i+1+j]
		}
	}
	return id, url
}

// stripInline removes ":::class" occurrences, recording each as an assignment
// to the node that precedes it (looking past any shape brackets).
func stripInline(line string, pending *[]classAssign) string {
	for {
		idx := strings.Index(line, ":::")
		if idx < 0 {
			return line
		}
		j := idx + 3
		for j < len(line) && isWordByte(line[j]) {
			j++
		}
		class := line[idx+3 : j]
		if id := nodeIDBefore(line, idx); id != "" && class != "" {
			*pending = append(*pending, classAssign{ids: []string{id}, name: class})
		}
		line = line[:idx] + line[j:]
	}
}

// nodeIDBefore returns the node identifier ending at idx, skipping a trailing
// shape (balanced brackets) if present.
func nodeIDBefore(s string, idx int) string {
	i := idx
	if i > 0 {
		switch s[i-1] {
		case ']', ')', '}':
			depth := 0
			for i > 0 {
				switch s[i-1] {
				case ']', ')', '}':
					depth++
				case '[', '(', '{':
					depth--
				}
				i--
				if depth == 0 {
					goto readID
				}
			}
		}
	}
readID:
	end := i
	for i > 0 && isWordByte(s[i-1]) {
		i--
	}
	return s[i:end]
}

func isWordByte(c byte) bool {
	return c == '_' || (c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func parseClassDef(line string) (string, *domain.Style) {
	rest := strings.TrimSpace(line[len("classDef "):])
	name, props, _ := strings.Cut(rest, " ")
	return strings.TrimSpace(name), parseProps(props)
}

func parseStyleStmt(line string) (string, *domain.Style) {
	rest := strings.TrimSpace(line[len("style "):])
	id, props, _ := strings.Cut(rest, " ")
	return strings.TrimSpace(id), parseProps(props)
}

// parseProps parses "fill:#f9f,stroke:#333,color:#fff" into a Style.
func parseProps(s string) *domain.Style {
	st := &domain.Style{}
	for _, kv := range strings.Split(s, ",") {
		k, v, ok := strings.Cut(kv, ":")
		if !ok {
			continue
		}
		switch strings.TrimSpace(k) {
		case "fill":
			st.Fill = strings.TrimSpace(v)
		case "stroke":
			st.Stroke = strings.TrimSpace(v)
		case "color":
			st.Color = strings.TrimSpace(v)
		}
	}
	return st
}

// mergeStyle overlays non-empty fields of st onto the style for id.
func mergeStyle(m map[string]*domain.Style, id string, st *domain.Style) {
	cur := m[id]
	if cur == nil {
		cur = &domain.Style{}
		m[id] = cur
	}
	if st.Fill != "" {
		cur.Fill = st.Fill
	}
	if st.Stroke != "" {
		cur.Stroke = st.Stroke
	}
	if st.Color != "" {
		cur.Color = st.Color
	}
}
