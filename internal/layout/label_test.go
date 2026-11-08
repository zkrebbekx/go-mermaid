package layout

import (
	"maps"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/zkrebbekx/go-mermaid/internal/domain"
	"github.com/zkrebbekx/go-mermaid/internal/parser"
)

// labelCentre returns the centre of e's label box.
func labelCentre(e *domain.Edge) domain.Point {
	r, _ := labelRect(e, opts.face(), opts.FontSize)
	return domain.Point{X: r.Min.X + r.Size.W/2, Y: r.Min.Y + r.Size.H/2}
}

// onPolyline reports whether p lies on one of the axis-aligned segments of pts.
func onPolyline(p domain.Point, pts []domain.Point) bool {
	const eps = 0.01
	for i := 1; i < len(pts); i++ {
		a, b := pts[i-1], pts[i]
		inX := p.X >= math.Min(a.X, b.X)-eps && p.X <= math.Max(a.X, b.X)+eps
		inY := p.Y >= math.Min(a.Y, b.Y)-eps && p.Y <= math.Max(a.Y, b.Y)+eps
		if inX && inY {
			return true
		}
	}
	return false
}

func nodeRect(n *domain.Node) domain.Rect {
	return domain.Rect{Min: n.Pos, Size: n.Size}
}

// labelCollisions returns a description of every label box that overlaps a
// node box or another label box.
func labelCollisions(g *domain.Graph) []string {
	var out []string
	var labelled []*domain.Edge
	for _, e := range g.Edges {
		if _, ok := labelRect(e, opts.face(), opts.FontSize); ok {
			labelled = append(labelled, e)
		}
	}
	for i, e := range labelled {
		r, _ := labelRect(e, opts.face(), opts.FontSize)
		for _, n := range g.Nodes {
			if overlaps(r, nodeRect(n)) {
				out = append(out, "label "+e.Label+" covers node "+n.ID)
			}
		}
		for _, o := range labelled[i+1:] {
			ro, _ := labelRect(o, opts.face(), opts.FontSize)
			if overlaps(r, ro) {
				out = append(out, "label "+e.Label+" covers label "+o.Label)
			}
		}
	}
	return out
}

// isFlowchart reports whether a Mermaid source declares a flowchart.
func isFlowchart(src string) bool {
	for _, line := range strings.Split(src, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "%%") {
			continue
		}
		return strings.HasPrefix(line, "graph") || strings.HasPrefix(line, "flowchart")
	}
	return false
}

func TestLabelsNeverCollideInGoldenCorpus(t *testing.T) {
	Convey("Given every flowchart in the golden corpus", t, func() {
		paths, err := filepath.Glob(filepath.Join("..", "..", "testdata", "golden", "*.mmd"))
		So(err, ShouldBeNil)
		flows := map[string]string{}
		for _, p := range paths {
			src, err := os.ReadFile(p)
			So(err, ShouldBeNil)
			if isFlowchart(string(src)) {
				flows[p] = string(src)
			}
		}
		So(flows, ShouldNotBeEmpty)

		for _, p := range slices.Sorted(maps.Keys(flows)) {
			src := flows[p]
			// Strip class and style directives the way the root renderer does.
			clean, _, _, _ := parser.Preprocess(src)
			g := graphFrom(clean)

			Convey("When laying out "+filepath.Base(p), func() {
				_, err := Compute(g, opts)

				Convey("Then no label covers a node or another label", func() {
					So(err, ShouldBeNil)
					So(labelCollisions(g), ShouldBeEmpty)
				})
			})
		}
	})
}

func TestLabelPlacement(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		{"fan-out with labels", "flowchart TD\nP[poll status] -- still running --> P\nP -- finished --> A[exit 0]\nP -- failed --> B[exit 1]\nP -- waiting for approval --> C[exit 3]\nP -- timeout or network error --> D[exit 4]"},
		{"LR labels wider than the rank gap", "flowchart LR\nA[Client] -->|signed token| B[Gateway] --> C[Authorizer]\nC -->|\"new route<br/>(field dropped)\"| D[Handler]"},
		{"TD labelled self-loop with a right neighbour", "flowchart TD\nR --> A\nR --> B\nA -->|a rather long loop label| A"},
		{"LR labelled self-loop with a neighbour below", "flowchart LR\nR --> A\nR --> B\nA -->|a rather long loop label| A"},
		{"opposite labelled parallel edges", "flowchart TD\nA -->|request| B\nB -->|response| A"},
		{"long labelled edge", "flowchart TD\nA --> B --> C --> D\nA -->|skip ahead| D"},
		{"BT labels", "flowchart BT\nA -->|up one| B\nA -->|up two| C"},
		{"RL labels", "flowchart RL\nA -->|left one| B\nA -->|left two| C"},
		{"labels inside a subgraph", "flowchart TD\nsubgraph S\nA -->|a very long label indeed| B\nend\nB -->|out| C"},
		{"mixed labelled and unlabelled parallels", "flowchart TD\nA -->|one| B\nA --> B\nB -->|back| A"},
		{"TD node with two labelled self-loops", "flowchart TD\nA -->|loop label that is long| A\nA -->|loop again| A\nA --> B"},
		{"LR node with two labelled self-loops", "flowchart LR\nA -->|loop label that is long| A\nA -->|loop again| A\nA --> B"},
	}
	for _, tc := range cases {
		Convey("Given "+tc.name, t, func() {
			g := graphFrom(tc.src)

			Convey("When computing the layout", func() {
				res, err := Compute(g, opts)

				Convey("Then no label covers a node or another label", func() {
					So(err, ShouldBeNil)
					So(labelCollisions(g), ShouldBeEmpty)
				})

				Convey("Then every label sits inside the bounds", func() {
					So(err, ShouldBeNil)
					for _, e := range g.Edges {
						r, ok := labelRect(e, opts.face(), opts.FontSize)
						if !ok {
							continue
						}
						So(r.Min.X, ShouldBeGreaterThanOrEqualTo, 0)
						So(r.Min.Y, ShouldBeGreaterThanOrEqualTo, 0)
						So(r.Min.X+r.Size.W, ShouldBeLessThanOrEqualTo, res.Width+0.01)
						So(r.Min.Y+r.Size.H, ShouldBeLessThanOrEqualTo, res.Height+0.01)
					}
				})

				Convey("Then no path repeats a point", func() {
					So(err, ShouldBeNil)
					for _, e := range g.Edges {
						for i := 1; i < len(e.Points); i++ {
							So(e.Points[i], ShouldNotResemble, e.Points[i-1])
						}
					}
				})

				Convey("Then every non-loop label is centred on its own edge", func() {
					So(err, ShouldBeNil)
					for _, e := range g.Edges {
						if e.Label == "" || e.From == e.To {
							continue
						}
						So(onPolyline(labelCentre(e), e.Points), ShouldBeTrue)
					}
				})
			})
		})
	}
}

func TestOppositeLabelledEdgesStayAligned(t *testing.T) {
	for _, dir := range []string{"TD", "LR"} {
		Convey("Given two "+dir+" nodes joined both ways by labelled edges", t, func() {
			g := graphFrom("flowchart " + dir + "\nA -->|request| B\nB -->|response| A")

			Convey("When computing the layout", func() {
				_, err := Compute(g, opts)

				Convey("Then the two nodes line up across the gap", func() {
					So(err, ShouldBeNil)
					a, b := g.NodeByID("A").Center(), g.NodeByID("B").Center()
					if dir == "TD" {
						So(a.X, ShouldAlmostEqual, b.X, 0.01)
					} else {
						So(a.Y, ShouldAlmostEqual, b.Y, 0.01)
					}
				})
			})
		})
	}
}

func TestLongLabelledEdgeUsesMiddleRank(t *testing.T) {
	Convey("Given a labelled edge spanning three ranks", t, func() {
		g := graphFrom("flowchart TD\nA --> B --> C --> D\nA -->|skip ahead| D")

		Convey("When computing the layout", func() {
			_, err := Compute(g, opts)

			Convey("Then the label sits in the gap between the two middle ranks", func() {
				So(err, ShouldBeNil)
				b, c := g.NodeByID("B"), g.NodeByID("C")
				y := labelCentre(g.Edges[3]).Y
				So(y, ShouldBeGreaterThan, b.Pos.Y+b.Size.H)
				So(y, ShouldBeLessThan, c.Pos.Y)
			})
		})
	})
}

func TestNestedSelfLoops(t *testing.T) {
	Convey("Given a node with two self-loops", t, func() {
		g := graphFrom("flowchart TD\nA -->|first| A\nA -->|second| A")

		Convey("When computing the layout", func() {
			_, err := Compute(g, opts)

			Convey("Then the second loop wraps around the first", func() {
				So(err, ShouldBeNil)
				inner, outer := g.Edges[0].Points, g.Edges[1].Points
				So(outer[1].X, ShouldBeGreaterThan, inner[1].X)
				So(outer[1].Y, ShouldBeLessThan, inner[1].Y)
				So(outer[2].Y, ShouldBeGreaterThan, inner[2].Y)
			})

			Convey("Then both labels sit outside the outer loop", func() {
				So(err, ShouldBeNil)
				for _, e := range g.Edges {
					r, _ := labelRect(e, opts.face(), opts.FontSize)
					So(r.Min.X, ShouldBeGreaterThan, g.Edges[1].Points[1].X)
				}
			})
		})
	})
}

func TestSelfLoopLabel(t *testing.T) {
	Convey("Given a TD node with a labelled self-loop", t, func() {
		g := graphFrom("flowchart TD\nA -->|again| A")

		Convey("When computing the layout", func() {
			_, err := Compute(g, opts)

			Convey("Then the label sits to the right of the loop, not on it", func() {
				So(err, ShouldBeNil)
				e := g.Edges[0]
				r, _ := labelRect(e, opts.face(), opts.FontSize)
				So(r.Min.X, ShouldBeGreaterThan, e.Points[1].X)
			})
		})
	})

	Convey("Given an LR node with a labelled self-loop", t, func() {
		g := graphFrom("flowchart LR\nA -->|again| A --> B")

		Convey("When computing the layout", func() {
			_, err := Compute(g, opts)

			Convey("Then the loop hangs below the node, clear of the rank gap", func() {
				So(err, ShouldBeNil)
				a, e := g.NodeByID("A"), g.Edges[0]
				for _, p := range e.Points {
					So(p.X, ShouldBeBetween, a.Pos.X, a.Pos.X+a.Size.W)
					So(p.Y, ShouldBeGreaterThanOrEqualTo, a.Pos.Y+a.Size.H)
				}
				r, _ := labelRect(e, opts.face(), opts.FontSize)
				So(r.Min.Y, ShouldBeGreaterThan, e.Points[1].Y)
			})
		})
	})
}

func TestRankGap(t *testing.T) {
	gapBetween := func(g *domain.Graph, from, to string) float64 {
		a, b := g.NodeByID(from), g.NodeByID(to)
		return b.Pos.Y - (a.Pos.Y + a.Size.H)
	}
	gap := func(g *domain.Graph) float64 { return gapBetween(g, "A", "B") }

	Convey("Given a graph with no edge labels", t, func() {
		g := graphFrom("flowchart TD\nA --> B")

		Convey("Then the gap between ranks is RankSep", func() {
			_, err := Compute(g, opts)
			So(err, ShouldBeNil)
			So(gap(g), ShouldAlmostEqual, opts.RankSep, 0.01)
		})
	})

	Convey("Given an unlabelled edge beside a labelled one", t, func() {
		g := graphFrom("flowchart TD\nA --> B\nC -->|x| D")

		Convey("Then both cross the same middle rank, so their gaps match", func() {
			_, err := Compute(g, opts)
			So(err, ShouldBeNil)
			So(gap(g), ShouldAlmostEqual, gapBetween(g, "C", "D"), 0.01)
		})
	})

	Convey("Given a labelled edge between adjacent ranks", t, func() {
		g := graphFrom("flowchart TD\nA -->|label| B")

		Convey("Then the gap grows by the label height", func() {
			_, err := Compute(g, opts)
			So(err, ShouldBeNil)
			h := labelSize("label", opts.face(), opts.FontSize).H
			So(gap(g), ShouldAlmostEqual, opts.RankSep+h, 0.01)
		})
	})
}

func TestShiftPortKeepsInteriorPoints(t *testing.T) {
	Convey("Given an edge that runs straight from its source through a dummy", t, func() {
		e := &domain.Edge{Points: []domain.Point{{X: 50, Y: 0}, {X: 50, Y: 40}, {X: 50, Y: 80}}}

		Convey("When its start moves along the face", func() {
			shiftPort(e, false, 30, true, 0)

			Convey("Then a jog is inserted and the dummy stays put", func() {
				So(e.Points, ShouldResemble, []domain.Point{
					{X: 30, Y: 0}, {X: 30, Y: 20}, {X: 50, Y: 20}, {X: 50, Y: 40}, {X: 50, Y: 80},
				})
			})
		})

		Convey("When its end moves along the face", func() {
			shiftPort(e, true, 70, true, 0)

			Convey("Then the jog is inserted before the end", func() {
				So(e.Points, ShouldResemble, []domain.Point{
					{X: 50, Y: 0}, {X: 50, Y: 40}, {X: 50, Y: 60}, {X: 70, Y: 60}, {X: 70, Y: 80},
				})
			})
		})
	})

	Convey("Given an edge that runs straight from its source through its label", t, func() {
		e := &domain.Edge{
			Points:   []domain.Point{{X: 50, Y: 0}, {X: 50, Y: 40}, {X: 50, Y: 80}},
			LabelPos: domain.Point{X: 50, Y: 40},
		}

		Convey("When its start moves along the face", func() {
			shiftPort(e, false, 30, true, 10)

			Convey("Then the jog sits halfway to the label's edge, not its centre", func() {
				So(e.Points[1], ShouldResemble, domain.Point{X: 30, Y: 15})
				So(e.Points[2], ShouldResemble, domain.Point{X: 50, Y: 15})
			})
		})
	})

	Convey("Given an edge that runs straight through its label into its target", t, func() {
		e := &domain.Edge{
			Points:   []domain.Point{{X: 50, Y: 0}, {X: 50, Y: 40}, {X: 50, Y: 80}},
			LabelPos: domain.Point{X: 50, Y: 40},
		}

		Convey("When its end moves along the face", func() {
			shiftPort(e, true, 70, true, 10)

			Convey("Then the jog sits halfway between the label's edge and the end", func() {
				So(e.Points[2], ShouldResemble, domain.Point{X: 50, Y: 65})
				So(e.Points[3], ShouldResemble, domain.Point{X: 70, Y: 65})
			})
		})
	})

	Convey("Given an end that already sits at its port", t, func() {
		pts := []domain.Point{{X: 50, Y: 0}, {X: 50, Y: 40}, {X: 50, Y: 80}}
		e := &domain.Edge{Points: slices.Clone(pts)}

		Convey("When it is moved to the same port", func() {
			shiftPort(e, false, 50, true, 0)

			Convey("Then the path is unchanged", func() {
				So(e.Points, ShouldResemble, pts)
			})
		})
	})

	Convey("Given an edge whose next point is an elbow", t, func() {
		e := &domain.Edge{Points: []domain.Point{{X: 50, Y: 0}, {X: 50, Y: 20}, {X: 90, Y: 20}, {X: 90, Y: 40}}}

		Convey("When its start moves along the face", func() {
			shiftPort(e, false, 30, true, 0)

			Convey("Then the elbow moves with it", func() {
				So(e.Points, ShouldResemble, []domain.Point{
					{X: 30, Y: 0}, {X: 30, Y: 20}, {X: 90, Y: 20}, {X: 90, Y: 40},
				})
			})
		})
	})
}
