package layout

import (
	"testing"

	"github.com/Zac300/go-mermaid/internal/domain"
	"github.com/Zac300/go-mermaid/internal/lexer"
	"github.com/Zac300/go-mermaid/internal/parser"
	. "github.com/smartystreets/goconvey/convey"
)

func graphFrom(src string) *domain.Graph {
	toks, err := lexer.Lex(src)
	if err != nil {
		panic(err)
	}
	g, err := parser.Parse(toks)
	if err != nil {
		panic(err)
	}
	return g
}

var opts = Options{NodeSep: 50, RankSep: 50, FontSize: 14}

func TestCompute(t *testing.T) {
	Convey("Given a top-down chain A->B->C", t, func() {
		g := graphFrom("graph TD\nA --> B --> C")

		Convey("When computing the layout", func() {
			res, err := Compute(g, opts)

			Convey("Then it reports positive bounds", func() {
				So(err, ShouldBeNil)
				So(res.Width, ShouldBeGreaterThan, 0)
				So(res.Height, ShouldBeGreaterThan, 0)
			})

			Convey("Then nodes are ranked top to bottom", func() {
				a, b, c := g.NodeByID("A"), g.NodeByID("B"), g.NodeByID("C")
				So(a.Pos.Y, ShouldBeLessThan, b.Pos.Y)
				So(b.Pos.Y, ShouldBeLessThan, c.Pos.Y)
			})

			Convey("Then every edge is routed with at least two points", func() {
				for _, e := range g.Edges {
					So(len(e.Points), ShouldBeGreaterThanOrEqualTo, 2)
				}
			})
		})
	})

	Convey("Given a top-down edge A->B", t, func() {
		g := graphFrom("graph TD\nA --> B")

		Convey("When computing the layout", func() {
			_, err := Compute(g, opts)

			Convey("Then the edge endpoints sit on the node boundaries, not centers", func() {
				So(err, ShouldBeNil)
				a, b := g.NodeByID("A"), g.NodeByID("B")
				e := g.Edges[0]
				start, end := e.Points[0], e.Points[len(e.Points)-1]
				// Leaves the bottom of A, lands on the top of B.
				So(start.Y, ShouldEqual, a.Pos.Y+a.Size.H)
				So(end.Y, ShouldEqual, b.Pos.Y)
				So(end.Y, ShouldNotEqual, b.Center().Y)
			})
		})
	})

	Convey("Given a left-right chain", t, func() {
		g := graphFrom("graph LR\nA --> B --> C")

		Convey("When computing the layout", func() {
			_, err := Compute(g, opts)

			Convey("Then ranks advance along X", func() {
				So(err, ShouldBeNil)
				a, b, c := g.NodeByID("A"), g.NodeByID("B"), g.NodeByID("C")
				So(a.Pos.X, ShouldBeLessThan, b.Pos.X)
				So(b.Pos.X, ShouldBeLessThan, c.Pos.X)
			})
		})
	})

	Convey("Given a cyclic graph A->B->A", t, func() {
		g := graphFrom("graph TD\nA --> B\nB --> A")

		Convey("When computing the layout", func() {
			_, err := Compute(g, opts)

			Convey("Then it terminates and restores edge directions", func() {
				So(err, ShouldBeNil)
				So(g.Edges[0].From, ShouldEqual, "A")
				So(g.Edges[0].To, ShouldEqual, "B")
				So(g.Edges[1].From, ShouldEqual, "B")
				So(g.Edges[1].To, ShouldEqual, "A")
			})
		})
	})

	Convey("Given an isolated node", t, func() {
		g := graphFrom("graph TD\nA")

		Convey("When computing the layout", func() {
			res, err := Compute(g, opts)

			Convey("Then it is placed at the first rank", func() {
				So(err, ShouldBeNil)
				So(res.Width, ShouldBeGreaterThan, 0)
				So(g.NodeByID("A").Pos.Y, ShouldEqual, 0)
			})
		})
	})

	Convey("Given a circle node", t, func() {
		g := graphFrom("graph TD\nA((Round))")

		Convey("When computing the layout", func() {
			_, err := Compute(g, opts)

			Convey("Then its box is squared", func() {
				So(err, ShouldBeNil)
				n := g.NodeByID("A")
				So(n.Size.W, ShouldEqual, n.Size.H)
			})
		})
	})
}
