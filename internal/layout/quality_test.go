package layout

import (
	"math"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/zkrebbekx/go-mermaid/internal/domain"
)

// primarySpan returns the extent of a node along the layout's rank axis.
func primarySpan(n *domain.Node, vertical bool) (lo, hi float64) {
	if vertical {
		return n.Pos.Y, n.Pos.Y + n.Size.H
	}
	return n.Pos.X, n.Pos.X + n.Size.W
}

func TestElbowSitsInTheGap(t *testing.T) {
	Convey("Given a source node far taller than the rank gap", t, func() {
		g := graphFrom("graph TD\nHuge --> S\nHuge --> T")
		g.NodeByID("Huge").Size = domain.Size{W: 90, H: 460}

		Convey("When computing the layout", func() {
			_, err := Compute(g, opts)

			Convey("Then no edge point falls inside the box it leaves", func() {
				So(err, ShouldBeNil)
				huge := g.NodeByID("Huge")
				top, bottom := primarySpan(huge, true)
				for _, e := range g.Edges {
					for i, p := range e.Points {
						if i == 0 {
							continue // the endpoint sits on the boundary
						}
						inside := p.Y > top+0.01 && p.Y < bottom-0.01 &&
							p.X > huge.Pos.X+0.01 && p.X < huge.Pos.X+huge.Size.W-0.01
						So(inside, ShouldBeFalse)
					}
				}
			})

			Convey("Then the path never doubles back along the rank axis", func() {
				So(err, ShouldBeNil)
				for _, e := range g.Edges {
					first, last := e.Points[0], e.Points[len(e.Points)-1]
					forward := last.Y > first.Y
					for i := 1; i < len(e.Points); i++ {
						d := e.Points[i].Y - e.Points[i-1].Y
						if math.Abs(d) < 0.01 {
							continue
						}
						So(d > 0, ShouldEqual, forward)
					}
				}
			})
		})
	})

	Convey("Given two boxes of ordinary height", t, func() {
		g := graphFrom("graph TD\nA --> B\nA --> C")

		Convey("When computing the layout", func() {
			_, err := Compute(g, opts)

			Convey("Then the elbow still sits between the ranks", func() {
				So(err, ShouldBeNil)
				a := g.NodeByID("A")
				aBottom := a.Pos.Y + a.Size.H
				e := g.Edges[0]
				So(len(e.Points), ShouldBeGreaterThanOrEqualTo, 3)
				So(e.Points[1].Y, ShouldBeGreaterThanOrEqualTo, aBottom-0.01)
			})
		})
	})
}

func TestGapMid(t *testing.T) {
	Convey("Given two boxes with clear space between them", t, func() {
		Convey("When the second is further along the axis", func() {
			Convey("Then the midpoint lies in the gap, not between the centers", func() {
				// A spans 0..100 (center 50, half 50); B spans 150..170.
				So(gapMid(50, 50, 160, 10), ShouldEqual, 125)
			})
		})

		Convey("When the second is earlier along the axis", func() {
			Convey("Then the gap is measured the other way round", func() {
				So(gapMid(160, 10, 50, 50), ShouldEqual, 125)
			})
		})
	})

	Convey("Given two boxes that overlap", t, func() {
		Convey("Then it falls back to the midpoint of the centers", func() {
			So(gapMid(50, 50, 60, 50), ShouldEqual, 55)
		})
	})
}

func TestSubgraphMembersStayTogether(t *testing.T) {
	Convey("Given a subgraph whose members are pulled apart by their parents", t, func() {
		src := "flowchart TD\nroot --> s1\nroot --> mid\nroot --> s2\nsubgraph g [Group]\ns1 --> t1\ns2 --> t2\nend\nmid --> u"
		g := graphFrom(src)

		Convey("When computing the layout", func() {
			_, err := Compute(g, opts)

			Convey("Then no outsider sits between two members on the same rank", func() {
				So(err, ShouldBeNil)
				members := map[string]bool{}
				for _, id := range g.Subgraphs[0].NodeIDs {
					members[id] = true
				}
				// Group nodes by rank using their y position.
				byRank := map[float64][]*domain.Node{}
				for _, n := range g.Nodes {
					byRank[n.Pos.Y] = append(byRank[n.Pos.Y], n)
				}
				for _, row := range byRank {
					var lo, hi float64
					seen := false
					for _, n := range row {
						if !members[n.ID] {
							continue
						}
						if !seen {
							lo, hi, seen = n.Pos.X, n.Pos.X+n.Size.W, true
							continue
						}
						lo = math.Min(lo, n.Pos.X)
						hi = math.Max(hi, n.Pos.X+n.Size.W)
					}
					if !seen {
						continue
					}
					for _, n := range row {
						if members[n.ID] {
							continue
						}
						centre := n.Pos.X + n.Size.W/2
						So(centre > lo && centre < hi, ShouldBeFalse)
					}
				}
			})
		})
	})

	Convey("Given two subgraphs interleaved by their parents", t, func() {
		src := "flowchart TD\nroot --> a\nroot --> b\nroot --> c\nroot --> d\nsubgraph one [One]\na --> p1\nc --> p2\nend\nsubgraph two [Two]\nb --> q1\nd --> q2\nend"
		g := graphFrom(src)

		Convey("When computing the layout", func() {
			_, err := Compute(g, opts)

			Convey("Then each subgraph's members are contiguous on their rank", func() {
				So(err, ShouldBeNil)
				So(len(g.Subgraphs), ShouldEqual, 2)
				for _, sg := range g.Subgraphs {
					members := map[string]bool{}
					for _, id := range sg.NodeIDs {
						members[id] = true
					}
					byRank := map[float64][]*domain.Node{}
					for _, n := range g.Nodes {
						byRank[n.Pos.Y] = append(byRank[n.Pos.Y], n)
					}
					for _, row := range byRank {
						var lo, hi float64
						seen := false
						for _, n := range row {
							if !members[n.ID] {
								continue
							}
							if !seen {
								lo, hi, seen = n.Pos.X, n.Pos.X+n.Size.W, true
								continue
							}
							lo = math.Min(lo, n.Pos.X)
							hi = math.Max(hi, n.Pos.X+n.Size.W)
						}
						if !seen {
							continue
						}
						for _, n := range row {
							if members[n.ID] {
								continue
							}
							centre := n.Pos.X + n.Size.W/2
							So(centre > lo && centre < hi, ShouldBeFalse)
						}
					}
				}
			})
		})
	})

	Convey("Given a graph with no subgraphs", t, func() {
		g := graphFrom("flowchart TD\nA --> B\nA --> C\nB --> D\nC --> D")

		Convey("When computing the layout", func() {
			_, err := Compute(g, opts)

			Convey("Then the ordering pass leaves it alone", func() {
				So(err, ShouldBeNil)
				So(g.NodeByID("A").Center().X, ShouldAlmostEqual,
					(g.NodeByID("B").Center().X+g.NodeByID("C").Center().X)/2, 0.5)
			})
		})
	})
}
