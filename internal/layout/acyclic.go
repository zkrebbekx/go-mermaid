package layout

import "github.com/zkrebbekx/go-mermaid/internal/domain"

// makeAcyclic removes cycles by reversing back edges found during a DFS.
// It returns the edges that were reversed so they can be restored after
// positioning. Ranking and ordering then operate on a DAG.
func makeAcyclic(g *domain.Graph) []*domain.Edge {
	const (
		white = 0 // unvisited
		gray  = 1 // on the current DFS stack
		black = 2 // fully explored
	)
	state := map[string]int{}
	var reversed []*domain.Edge

	out := adjacency(g)

	var visit func(id string)
	visit = func(id string) {
		state[id] = gray
		for _, e := range out[id] {
			switch state[e.To] {
			case gray:
				// back edge: reverse it to break the cycle
				e.From, e.To = e.To, e.From
				reversed = append(reversed, e)
			case white:
				visit(e.To)
			}
		}
		state[id] = black
	}

	for _, n := range g.Nodes {
		if state[n.ID] == white {
			visit(n.ID)
		}
	}
	return reversed
}

// restoreReversed flips reversed edges back to their original direction so
// arrowheads render correctly.
func restoreReversed(_ *domain.Graph, reversed []*domain.Edge) {
	for _, e := range reversed {
		e.From, e.To = e.To, e.From
	}
}

// adjacency builds an out-edge map keyed by node ID.
func adjacency(g *domain.Graph) map[string][]*domain.Edge {
	out := make(map[string][]*domain.Edge, len(g.Nodes))
	for _, e := range g.Edges {
		out[e.From] = append(out[e.From], e)
	}
	return out
}
