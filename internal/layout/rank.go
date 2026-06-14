package layout

import "github.com/zkrebbekx/go-mermaid/internal/domain"

// assignRanks computes a layer index per node using longest-path ranking:
// each node's rank is one greater than the maximum rank of its predecessors.
// The input graph is assumed acyclic (see makeAcyclic).
func assignRanks(g *domain.Graph) map[string]int {
	preds := map[string][]string{}
	indeg := map[string]int{}
	out := adjacency(g)
	for _, n := range g.Nodes {
		indeg[n.ID] = 0
	}
	for _, e := range g.Edges {
		preds[e.To] = append(preds[e.To], e.From)
		indeg[e.To]++
	}

	rank := map[string]int{}
	// Kahn's algorithm in topological order guarantees predecessors are
	// ranked before their successors.
	var queue []string
	for _, n := range g.Nodes {
		if indeg[n.ID] == 0 {
			queue = append(queue, n.ID)
		}
	}
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		r := 0
		for _, p := range preds[id] {
			if rank[p]+1 > r {
				r = rank[p] + 1
			}
		}
		rank[id] = r
		for _, e := range out[id] {
			indeg[e.To]--
			if indeg[e.To] == 0 {
				queue = append(queue, e.To)
			}
		}
	}

	// Any node not reached (e.g. isolated) defaults to rank 0.
	for _, n := range g.Nodes {
		if _, ok := rank[n.ID]; !ok {
			rank[n.ID] = 0
		}
	}
	return rank
}
