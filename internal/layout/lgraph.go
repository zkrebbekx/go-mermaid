package layout

import (
	"sort"

	"github.com/zkrebbekx/go-mermaid/internal/domain"
)

// lnode is a layout node: either a real graph node or a dummy inserted to
// route an edge that spans more than one rank. Dummies let long edges bend
// around intermediate ranks and make crossing counts meaningful.
type lnode struct {
	real  *domain.Node // nil for a dummy
	rank  int
	pos   int     // order within the rank (left to right in rank space)
	cross float64 // coordinate along the cross axis (set by positioning)
	pc    float64 // center along the primary/rank axis (set by positioning)
	w, h  float64 // domain box size (0 for dummies)
}

// lgraph is the layered working graph: nodes bucketed by rank, with up/down
// adjacency and a routing chain per edge.
type lgraph struct {
	layers [][]*lnode
	ups    map[*lnode][]*lnode // neighbors one rank up
	downs  map[*lnode][]*lnode // neighbors one rank down
	chains map[*domain.Edge][]*lnode
}

// buildLGraph buckets nodes by rank and inserts dummies along edges that span
// multiple ranks. Self-edges keep a two-element chain and are routed specially.
func buildLGraph(g *domain.Graph, ranks map[string]int) *lgraph {
	maxRank := 0
	for _, r := range ranks {
		if r > maxRank {
			maxRank = r
		}
	}
	lg := &lgraph{
		layers: make([][]*lnode, maxRank+1),
		ups:    map[*lnode][]*lnode{},
		downs:  map[*lnode][]*lnode{},
		chains: map[*domain.Edge][]*lnode{},
	}

	byID := make(map[string]*lnode, len(g.Nodes))
	for _, n := range g.Nodes {
		r := ranks[n.ID]
		ln := &lnode{real: n, rank: r, w: n.Size.W, h: n.Size.H}
		lg.layers[r] = append(lg.layers[r], ln)
		byID[n.ID] = ln
	}

	for _, e := range g.Edges {
		from, to := byID[e.From], byID[e.To]
		if from == nil || to == nil {
			continue
		}
		chain := []*lnode{from}
		if to.rank > from.rank+1 {
			prev := from
			for r := from.rank + 1; r < to.rank; r++ {
				d := &lnode{rank: r, w: 1}
				lg.layers[r] = append(lg.layers[r], d)
				lg.link(prev, d)
				chain = append(chain, d)
				prev = d
			}
			lg.link(prev, to)
		} else if to.rank > from.rank {
			lg.link(from, to)
		}
		chain = append(chain, to)
		lg.chains[e] = chain
	}

	lg.renumber()
	return lg
}

// link records an adjacency where a is one rank above b.
func (lg *lgraph) link(a, b *lnode) {
	lg.downs[a] = append(lg.downs[a], b)
	lg.ups[b] = append(lg.ups[b], a)
}

// renumber sets each node's pos to its index within its rank.
func (lg *lgraph) renumber() {
	for _, layer := range lg.layers {
		for i, ln := range layer {
			ln.pos = i
		}
	}
}

// reduceCrossings reorders nodes within each rank using the median heuristic,
// alternating downward and upward sweeps to lower edge crossings.
func reduceCrossings(lg *lgraph) {
	for iter := 0; iter < 6; iter++ {
		if iter%2 == 0 {
			for r := 1; r < len(lg.layers); r++ {
				sortByMedian(lg.layers[r], lg.ups)
			}
		} else {
			for r := len(lg.layers) - 2; r >= 0; r-- {
				sortByMedian(lg.layers[r], lg.downs)
			}
		}
		lg.renumber()
	}
}

// sortByMedian orders a layer by the median position of each node's neighbors
// in the adjacent (already-ordered) rank. Nodes with no neighbors keep their
// current relative position.
func sortByMedian(layer []*lnode, adj map[*lnode][]*lnode) {
	med := make(map[*lnode]float64, len(layer))
	for i, ln := range layer {
		ns := adj[ln]
		if len(ns) == 0 {
			med[ln] = float64(i) // fixed: keep place
			continue
		}
		ps := make([]float64, len(ns))
		for j, nb := range ns {
			ps[j] = float64(nb.pos)
		}
		sort.Float64s(ps)
		med[ln] = medianOf(ps)
	}
	sort.SliceStable(layer, func(i, j int) bool { return med[layer[i]] < med[layer[j]] })
	for i, ln := range layer {
		ln.pos = i
	}
}

func medianOf(sorted []float64) float64 {
	n := len(sorted)
	if n == 0 {
		return 0
	}
	if n%2 == 1 {
		return sorted[n/2]
	}
	return (sorted[n/2-1] + sorted[n/2]) / 2
}
