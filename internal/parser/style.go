package parser

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/zkrebbekx/go-mermaid/internal/domain"
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
func Preprocess(src string) (string, map[string]*domain.Style, map[string]string, *LinkStyles) {
	classDefs := map[string]*domain.Style{}
	styles := map[string]*domain.Style{}
	links := map[string]string{}
	linkStyles := &LinkStyles{ByIndex: map[int]*domain.Style{}}

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
		case strings.HasPrefix(t, "linkStyle "):
			parseLinkStyle(t, linkStyles)
		case strings.HasPrefix(t, "click "):
			if id, url := parseClick(t); id != "" && url != "" {
				links[id] = url
			}
		default:
			kept = append(kept, stripInline(expandShapeMeta(line), &pending))
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
	return strings.Join(kept, "\n"), styles, links, linkStyles
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
	if !safeURL(url) {
		url = ""
	}
	return id, url
}

// safeURL reports whether a click target is safe to emit as an href. Only
// http(s), mailto, and same-document/relative links are allowed; schemes like
// javascript: are rejected so a diagram can't smuggle script into the SVG.
func safeURL(u string) bool {
	u = strings.TrimSpace(u)
	if u == "" {
		return false
	}
	lower := strings.ToLower(u)
	switch {
	case strings.HasPrefix(lower, "http://"), strings.HasPrefix(lower, "https://"),
		strings.HasPrefix(lower, "mailto:"):
		return true
	case strings.HasPrefix(u, "/"), strings.HasPrefix(u, "#"), strings.HasPrefix(u, "./"),
		strings.HasPrefix(u, "../"):
		return true
	}
	// A bare relative path with no scheme (no colon before the first / or #) is
	// fine; a colon earlier than any path separator implies a scheme we don't
	// trust.
	colon := strings.IndexByte(u, ':')
	if colon < 0 {
		return true
	}
	slash := strings.IndexAny(u, "/#")
	return slash >= 0 && slash < colon
}

// stripInline removes ":::class" occurrences, recording each as an assignment
// to the node that precedes it (looking past any shape brackets).
func stripInline(line string, pending *[]classAssign) string {
	for {
		idx := topLevelTripleColon(line)
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

// topLevelTripleColon returns the index of the first ":::" that sits outside any
// shape bracket or quoted label, or -1. This keeps a literal ":::" inside a node
// label from being mistaken for an inline class assignment.
func topLevelTripleColon(s string) int {
	depth, inQuote := 0, false
	for i := 0; i+2 < len(s); i++ {
		c := s[i]
		if inQuote {
			if c == '"' {
				inQuote = false
			}
			continue
		}
		switch c {
		case '"':
			inQuote = true
		case '[', '(', '{':
			depth++
		case ']', ')', '}':
			if depth > 0 {
				depth--
			}
		case ':':
			if depth == 0 && s[i+1] == ':' && s[i+2] == ':' {
				return i
			}
		}
	}
	return -1
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
		case "stroke-width":
			st.StrokeWidth = strings.TrimSuffix(strings.TrimSpace(v), "px")
		case "stroke-dasharray":
			st.StrokeDash = strings.TrimSpace(v)
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
	if st.StrokeWidth != "" {
		cur.StrokeWidth = st.StrokeWidth
	}
	if st.StrokeDash != "" {
		cur.StrokeDash = st.StrokeDash
	}
}

// LinkStyles holds the per-edge overrides from linkStyle directives. Default
// applies to every edge that has no index-specific entry.
type LinkStyles struct {
	ByIndex map[int]*domain.Style
	Default *domain.Style
}

// For returns the style for the edge at index i, or nil when none applies.
func (l *LinkStyles) For(i int) *domain.Style {
	if l == nil {
		return nil
	}
	if st, ok := l.ByIndex[i]; ok {
		return st
	}
	return l.Default
}

// parseLinkStyle reads `linkStyle 0,2 stroke:#f00,stroke-width:4px` and
// `linkStyle default ...`. The selector is a comma-separated list of edge
// indexes in source order, matching Mermaid.
func parseLinkStyle(line string, into *LinkStyles) {
	rest := strings.TrimSpace(line[len("linkStyle "):])
	sel, props, ok := strings.Cut(rest, " ")
	if !ok {
		return
	}
	st := parseProps(props)
	if strings.EqualFold(strings.TrimSpace(sel), "default") {
		into.Default = st
		return
	}
	for _, part := range strings.Split(sel, ",") {
		n, err := strconv.Atoi(strings.TrimSpace(part))
		if err != nil {
			continue
		}
		into.ByIndex[n] = st
	}
}

// shapeMetaRe matches the Mermaid 11 metadata form `A@{ shape: rect,
// label: "Hi" }` on one line.
var shapeMetaRe = regexp.MustCompile(`([A-Za-z0-9_.-]+)@\{([^}]*)\}`)

// shapeDelims maps a Mermaid 11 shape name onto the bracket pair the lexer
// already understands.
var shapeDelims = map[string][2]string{
	"rect": {"[", "]"}, "rectangle": {"[", "]"}, "process": {"[", "]"},
	"rounded": {"(", ")"}, "round": {"(", ")"},
	"stadium": {"([", "])"}, "pill": {"([", "])"},
	"circle":  {"((", "))"},
	"diamond": {"{", "}"}, "decision": {"{", "}"}, "rhombus": {"{", "}"},
	"hexagon": {"{{", "}}"}, "hex": {"{{", "}}"},
	"cylinder": {"[(", ")]"}, "database": {"[(", ")]"}, "db": {"[(", ")]"},
	"subroutine": {"[[", "]]"}, "subprocess": {"[[", "]]"},
}

// expandShapeMeta rewrites the Mermaid 11 metadata form into the bracket form
// the lexer understands, so `A@{ shape: circle, label: "Hi" }` becomes
// `A((Hi))`. An unknown shape falls back to a rectangle.
func expandShapeMeta(line string) string {
	return shapeMetaRe.ReplaceAllStringFunc(line, func(m string) string {
		sub := shapeMetaRe.FindStringSubmatch(m)
		id, body := sub[1], sub[2]
		shape, label := "", id
		for _, part := range splitMeta(body) {
			k, v, ok := strings.Cut(part, ":")
			if !ok {
				continue
			}
			v = strings.Trim(strings.TrimSpace(v), `"'`)
			switch strings.ToLower(strings.TrimSpace(k)) {
			case "shape":
				shape = strings.ToLower(v)
			case "label", "title":
				label = v
			}
		}
		d, ok := shapeDelims[shape]
		if !ok {
			d = [2]string{"[", "]"}
		}
		return id + d[0] + label + d[1]
	})
}

// splitMeta splits metadata on commas that sit outside quotes, so a label may
// contain a comma.
func splitMeta(s string) []string {
	var out []string
	var cur strings.Builder
	inQuote := byte(0)
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case inQuote != 0 && c == inQuote:
			inQuote = 0
			cur.WriteByte(c)
		case inQuote == 0 && (c == '"' || c == '\''):
			inQuote = c
			cur.WriteByte(c)
		case inQuote == 0 && c == ',':
			out = append(out, cur.String())
			cur.Reset()
		default:
			cur.WriteByte(c)
		}
	}
	if cur.Len() > 0 {
		out = append(out, cur.String())
	}
	return out
}
