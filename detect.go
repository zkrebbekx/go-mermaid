package mermaid

import "strings"

// kind identifies a Mermaid diagram type.
type kind int

const (
	kindUnknown kind = iota
	kindFlowchart
	kindSequence
	kindPie
)

// detectKind inspects the first non-empty, non-comment line and maps its
// leading keyword to a diagram kind. Detection is case-insensitive on the
// keyword so "sequenceDiagram" and "flowchart" match regardless of casing.
func detectKind(src string) kind {
	for _, raw := range strings.Split(src, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "%%") {
			continue
		}
		word := strings.ToLower(firstField(line))
		switch word {
		case "graph", "flowchart":
			return kindFlowchart
		case "sequencediagram":
			return kindSequence
		case "pie":
			return kindPie
		default:
			return kindUnknown
		}
	}
	return kindUnknown
}

func firstField(s string) string {
	if i := strings.IndexAny(s, " \t"); i >= 0 {
		return s[:i]
	}
	return s
}
