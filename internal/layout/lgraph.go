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

	// cluster maps a node to the subgraph it belongs to, empty when it is
	// outside every subgraph. Ordering keeps a cluster's members together.
	cluster map[*lnode]string
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
		layers:  make([][]*lnode, maxRank+1),
		ups:     map[*lnode][]*lnode{},
		downs:   map[*lnode][]*lnode{},
		chains:  map[*domain.Edge][]*lnode{},
		cluster: map[*lnode]string{},
	}

	// Which subgraph each node belongs to, by ID.
	nodeCluster := map[string]string{}
	for _, sg := range g.Subgraphs {
		key := sg.ID
		if key == "" {
			key = sg.Title
		}
		for _, id := range sg.NodeIDs {
			nodeCluster[id] = key
		}
	}

	byID := make(map[string]*lnode, len(g.Nodes))
	for _, n := range g.Nodes {
		r := ranks[n.ID]
		ln := &lnode{real: n, rank: r, w: n.Size.W, h: n.Size.H}
		lg.layers[r] = append(lg.layers[r], ln)
		byID[n.ID] = ln
		if c := nodeCluster[n.ID]; c != "" {
			lg.cluster[ln] = c
		}
	}

	for _, e := range g.Edges {
		from, to := byID[e.From], byID[e.To]
		if from == nil || to == nil {
			continue
		}
		chain := []*lnode{from}
		if to.rank > from.rank+1 {
			prev := from
			// A dummy joins the cluster only when both ends share one, so a
			// long edge crossing a subgraph does not drag the box open.
			dummyCluster := ""
			if c := nodeCluster[e.From]; c != "" && c == nodeCluster[e.To] {
				dummyCluster = c
			}
			for r := from.rank + 1; r < to.rank; r++ {
				d := &lnode{rank: r, w: 1}
				if dummyCluster != "" {
					lg.cluster[d] = dummyCluster
				}
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
				groupByCluster(lg.layers[r], lg.cluster)
			}
		} else {
			for r := len(lg.layers) - 2; r >= 0; r-- {
				sortByMedian(lg.layers[r], lg.downs)
				groupByCluster(lg.layers[r], lg.cluster)
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

// groupByCluster reorders a layer so the members of one subgraph sit next to
// each other, keeping their relative order. Each cluster is placed at the
// average position of its members, so the pass does not undo the crossing
// reduction it follows.
//
// Without it the ordering step is blind to subgraphs: members interleave with
// outside nodes, and the boxes drawn around them overlap.
func groupByCluster(layer []*lnode, cluster map[*lnode]string) {
	if len(layer) < 2 {
		return
	}
	type group struct {
		members []*lnode
		sum     float64
	}
	var groups []*group
	byKey := map[string]*group{}
	clustered := false
	for i, ln := range layer {
		key := cluster[ln]
		if key == "" {
			// An unclustered node keeps its own place.
			groups = append(groups, &group{members: []*lnode{ln}, sum: float64(i)})
			continue
		}
		clustered = true
		gr, ok := byKey[key]
		if !ok {
			gr = &group{}
			byKey[key] = gr
			groups = append(groups, gr)
		}
		gr.members = append(gr.members, ln)
		gr.sum += float64(i)
	}
	if !clustered {
		return
	}
	sort.SliceStable(groups, func(i, j int) bool {
		return groups[i].sum/float64(len(groups[i].members)) <
			groups[j].sum/float64(len(groups[j].members))
	})
	at := 0
	for _, gr := range groups {
		for _, ln := range gr.members {
			layer[at] = ln
			at++
		}
	}
	for i, ln := range layer {
		ln.pos = i
	}
}
