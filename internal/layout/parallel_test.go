package layout

import (
	"math"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/zkrebbekx/go-mermaid/internal/domain"
)

// overlaps reports whether two axis-aligned rects intersect.
func overlaps(a, b domain.Rect) bool {
	return a.Min.X < b.Min.X+b.Size.W && b.Min.X < a.Min.X+a.Size.W &&
		a.Min.Y < b.Min.Y+b.Size.H && b.Min.Y < a.Min.Y+a.Size.H
}

func TestReversedEdgeDirection(t *testing.T) {
	Convey("Given a cycle A->B, B->A", t, func() {
		g := graphFrom("graph TD\nA --> B\nB --> A")

		Convey("When computing the layout", func() {
			_, err := Compute(g, opts)

			Convey("Then each edge's polyline starts at its source and ends at its target", func() {
				So(err, ShouldBeNil)
				a, b := g.NodeByID("A"), g.NodeByID("B")
				ab, ba := g.Edges[0], g.Edges[1]
				// A->B leaves the bottom of A and lands on the top of B.
				So(ab.Points[0].Y, ShouldEqual, a.Pos.Y+a.Size.H)
				So(ab.Points[len(ab.Points)-1].Y, ShouldEqual, b.Pos.Y)
				// B->A was reversed to break the cycle: it must still leave the
				// top of B and land on the bottom of A, so the arrowhead is on A.
				So(ba.Points[0].Y, ShouldEqual, b.Pos.Y)
				So(ba.Points[len(ba.Points)-1].Y, ShouldEqual, a.Pos.Y+a.Size.H)
			})

			Convey("Then the two edges are still drawn apart", func() {
				So(err, ShouldBeNil)
				ab, ba := g.Edges[0], g.Edges[1]
				So(math.Abs(ab.Points[0].X-ba.Points[0].X), ShouldAlmostEqual, parallelSep, 0.01)
			})
		})
	})

	Convey("Given a longer cycle A->B->C->A", t, func() {
		g := graphFrom("graph TD\nA --> B\nB --> C\nC --> A")

		Convey("When computing the layout", func() {
			_, err := Compute(g, opts)

			Convey("Then the back edge C->A ends on A", func() {
				So(err, ShouldBeNil)
				a, c := g.NodeByID("A"), g.NodeByID("C")
				ca := g.Edges[2]
				first, last := ca.Points[0], ca.Points[len(ca.Points)-1]
				So(first.Y, ShouldEqual, c.Pos.Y)
				So(last.Y, ShouldEqual, a.Pos.Y+a.Size.H)
			})
		})
	})
}

func TestParallelEdgeLabels(t *testing.T) {
	Convey("Given a top-down pair of opposite labeled edges (issue #27)", t, func() {
		g := graphFrom("flowchart TD\nClient -->|request| API\nAPI -->|response| Client")

		Convey("When computing the layout", func() {
			res, err := Compute(g, opts)

			Convey("Then the label boxes do not overlap", func() {
				So(err, ShouldBeNil)
				r0, ok0 := labelRect(g.Edges[0], opts.face(), opts.FontSize)
				r1, ok1 := labelRect(g.Edges[1], opts.face(), opts.FontSize)
				So(ok0, ShouldBeTrue)
				So(ok1, ShouldBeTrue)
				So(overlaps(r0, r1), ShouldBeFalse)
			})

			Convey("Then each label sits on its own edge", func() {
				So(err, ShouldBeNil)
				for _, e := range g.Edges {
					So(e.LabelPos.X, ShouldEqual, e.Points[0].X)
					lo := math.Min(e.Points[0].Y, e.Points[1].Y)
					hi := math.Max(e.Points[0].Y, e.Points[1].Y)
					So(e.LabelPos.Y, ShouldBeBetween, lo, hi)
				}
			})

			Convey("Then nothing is placed at a negative coordinate and the bounds cover the labels", func() {
				So(err, ShouldBeNil)
				for _, e := range g.Edges {
					r, _ := labelRect(e, opts.face(), opts.FontSize)
					So(r.Min.X, ShouldBeGreaterThanOrEqualTo, 0)
					So(r.Min.Y, ShouldBeGreaterThanOrEqualTo, 0)
					So(r.Min.X+r.Size.W, ShouldBeLessThanOrEqualTo, res.Width)
					So(r.Min.Y+r.Size.H, ShouldBeLessThanOrEqualTo, res.Height)
				}
				for _, n := range g.Nodes {
					So(n.Pos.X, ShouldBeGreaterThanOrEqualTo, 0)
				}
			})
		})
	})

	Convey("Given a left-right pair of opposite labeled edges", t, func() {
		g := graphFrom("flowchart LR\nClient -->|request| API\nAPI -->|response| Client")

		Convey("When computing the layout", func() {
			_, err := Compute(g, opts)

			Convey("Then the labels already clear each other and stay side by side at the midpoint", func() {
				So(err, ShouldBeNil)
				r0, _ := labelRect(g.Edges[0], opts.face(), opts.FontSize)
				r1, _ := labelRect(g.Edges[1], opts.face(), opts.FontSize)
				So(overlaps(r0, r1), ShouldBeFalse)
				for _, e := range g.Edges {
					mid := domain.PolylineMidpoint(e.Points)
					So(e.LabelPos, ShouldResemble, mid)
				}
			})
		})
	})

	Convey("Given three labeled edges between the same pair", t, func() {
		g := graphFrom("graph TD\nA -->|alpha| B\nA -->|beta| B\nB -->|gamma| A")

		Convey("When computing the layout", func() {
			_, err := Compute(g, opts)

			Convey("Then no two label boxes overlap", func() {
				So(err, ShouldBeNil)
				for i := range g.Edges {
					for j := i + 1; j < len(g.Edges); j++ {
						ri, _ := labelRect(g.Edges[i], opts.face(), opts.FontSize)
						rj, _ := labelRect(g.Edges[j], opts.face(), opts.FontSize)
						So(overlaps(ri, rj), ShouldBeFalse)
					}
				}
			})
		})
	})

	Convey("Given a lone bent edge with a label", t, func() {
		g := graphFrom("graph TD\nA -->|yes| B\nA --> C")

		Convey("When computing the layout", func() {
			_, err := Compute(g, opts)

			Convey("Then the label anchors on the routed path midpoint", func() {
				So(err, ShouldBeNil)
				e := g.Edges[0]
				So(e.LabelPos, ShouldResemble, domain.PolylineMidpoint(e.Points))
			})
		})
	})
}

func TestEdgeLabelLineBreaks(t *testing.T) {
	Convey("Given an edge label containing a line break", t, func() {
		g := graphFrom("flowchart TD\nA -->|line one<br/>line two| B")

		Convey("When computing the layout", func() {
			_, err := Compute(g, opts)

			Convey("Then the label box is sized for two lines, not one long one", func() {
				So(err, ShouldBeNil)
				sz := labelSize(g.Edges[0].Label, opts.face(), opts.FontSize)
				So(sz.H, ShouldAlmostEqual, opts.FontSize*2+4, 0.01)
				oneLine := labelSize("line one", opts.face(), opts.FontSize)
				So(sz.W, ShouldAlmostEqual, oneLine.W, 0.01)
			})
		})
	})
}

func TestFontFaceReachesMeasurement(t *testing.T) {
	Convey("Given the same node label under two font families", t, func() {
		narrow := graphFrom("flowchart TD\nA[iiiiiiiiii] --> B")
		wide := graphFrom("flowchart TD\nA[iiiiiiiiii] --> B")

		Convey("When laying out with sans and with monospace", func() {
			sansOpts := Options{NodeSep: 50, RankSep: 50, FontSize: 14, FontFace: "sans-serif"}
			monoOpts := Options{NodeSep: 50, RankSep: 50, FontSize: 14, FontFace: "monospace"}
			_, err1 := Compute(narrow, sansOpts)
			_, err2 := Compute(wide, monoOpts)

			Convey("Then the monospace box is wider, because the metrics differ", func() {
				So(err1, ShouldBeNil)
				So(err2, ShouldBeNil)
				So(wide.NodeByID("A").Size.W, ShouldBeGreaterThan, narrow.NodeByID("A").Size.W)
			})
		})
	})
}
