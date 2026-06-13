package mermaid

import "strings"

// kind identifies a Mermaid diagram type.
type kind int

const (
	kindUnknown kind = iota
	kindFlowchart
	kindSequence
	kindPie
	kindClass
	kindState
	kindER
	kindJourney
	kindQuadrant
	kindGit
	kindTimeline
	kindMindmap
	kindGantt
	kindC4
	kindRequirement
	kindSankey
	kindXYChart
	kindBlock
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
		word := strings.TrimRight(strings.ToLower(firstField(line)), ":")
		if strings.HasPrefix(word, "c4") {
			return kindC4
		}
		switch word {
		case "graph", "flowchart":
			return kindFlowchart
		case "sequencediagram":
			return kindSequence
		case "pie":
			return kindPie
		case "classdiagram":
			return kindClass
		case "statediagram", "statediagram-v2":
			return kindState
		case "erdiagram":
			return kindER
		case "journey":
			return kindJourney
		case "quadrantchart":
			return kindQuadrant
		case "gitgraph":
			return kindGit
		case "timeline":
			return kindTimeline
		case "mindmap":
			return kindMindmap
		case "gantt":
			return kindGantt
		case "requirementdiagram":
			return kindRequirement
		case "sankey-beta", "sankey":
			return kindSankey
		case "xychart-beta", "xychart":
			return kindXYChart
		case "block-beta", "block":
			return kindBlock
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
