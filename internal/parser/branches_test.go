package parser

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/zkrebbekx/go-mermaid/internal/domain"
)

func TestDirections(t *testing.T) {
	Convey("Given each direction keyword", t, func() {
		cases := map[string]domain.Direction{
			"graph TD\nA": domain.TopBottom,
			"graph TB\nA": domain.TopBottom,
			"graph BT\nA": domain.BottomTop,
			"graph LR\nA": domain.LeftRight,
			"graph RL\nA": domain.RightLeft,
		}
		for src, want := range cases {
			src, want := src, want
			Convey("When parsing "+src, func() {
				g, err := parse(src)

				Convey("Then the graph direction matches", func() {
					So(err, ShouldBeNil)
					So(g.Direction, ShouldEqual, want)
				})
			})
		}
	})
}

func TestArrowKinds(t *testing.T) {
	Convey("Given each arrow style", t, func() {
		cases := map[string]domain.Arrow{
			"graph TD\nA --> B":  domain.ArrowNormal,
			"graph TD\nA --- B":  domain.ArrowOpen,
			"graph TD\nA -.-> B": domain.ArrowDotted,
			"graph TD\nA ==> B":  domain.ArrowThick,
		}
		for src, want := range cases {
			src, want := src, want
			Convey("When parsing "+src, func() {
				g, err := parse(src)

				Convey("Then the edge arrow kind matches", func() {
					So(err, ShouldBeNil)
					So(g.Edges[0].Arrow, ShouldEqual, want)
				})
			})
		}
	})
}

func TestParseErrors(t *testing.T) {
	Convey("Given invalid source", t, func() {
		cases := map[string]string{
			"unknown direction":  "graph XY\nA --> B",
			"junk after header":  "graph TD A --> B",
			"missing node id":    "graph TD\n--> B",
			"missing target":     "graph TD\nA -->",
			"unterminated label": "graph TD\nA -->|x B",
		}
		for name, src := range cases {
			src := src
			Convey("When parsing the "+name+" case", func() {
				_, err := parse(src)

				Convey("Then it returns an error", func() {
					So(err, ShouldNotBeNil)
				})
			})
		}
	})
}

func TestNodeShapes(t *testing.T) {
	Convey("Given each node shape syntax", t, func() {
		cases := []struct {
			src   string
			shape domain.Shape
		}{
			{"graph TD\nA[r]", domain.ShapeRect},
			{"graph TD\nA(r)", domain.ShapeRound},
			{"graph TD\nA([r])", domain.ShapeStadium},
			{"graph TD\nA((r))", domain.ShapeCircle},
			{"graph TD\nA{r}", domain.ShapeDiamond},
			{"graph TD\nA{{r}}", domain.ShapeHexagon},
			{"graph TD\nA[[r]]", domain.ShapeSubroutine},
			{"graph TD\nA[(r)]", domain.ShapeCylinder},
			{"graph TD\nA[/r/]", domain.ShapeParallelogram},
			{"graph TD\nA[\\r\\]", domain.ShapeParallelogramAlt},
			{"graph TD\nA[/r\\]", domain.ShapeTrapezoid},
			{"graph TD\nA[\\r/]", domain.ShapeTrapezoidAlt},
		}
		for _, c := range cases {
			c := c
			Convey("When parsing "+c.src, func() {
				g, err := parse(c.src)

				Convey("Then the node shape matches", func() {
					So(err, ShouldBeNil)
					So(g.NodeByID("A").Shape, ShouldEqual, c.shape)
				})
			})
		}
	})
}

func TestMiddleFormLabels(t *testing.T) {
	Convey("Given middle-form edge labels", t, func() {
		cases := []struct {
			src   string
			label string
			arrow domain.Arrow
		}{
			{"graph LR\nA -- yes --> B", "yes", domain.ArrowNormal},
			{"graph LR\nA == no ==> B", "no", domain.ArrowThick},
			{"graph LR\nA -. maybe .-> B", "maybe", domain.ArrowDotted},
			{"graph LR\nA -- two words --> B", "two words", domain.ArrowNormal},
		}
		for _, c := range cases {
			c := c
			Convey("When parsing "+c.src, func() {
				g, err := parse(c.src)

				Convey("Then the label and arrow are captured on a single edge", func() {
					So(err, ShouldBeNil)
					So(len(g.Edges), ShouldEqual, 1)
					So(g.Edges[0].Label, ShouldEqual, c.label)
					So(g.Edges[0].Arrow, ShouldEqual, c.arrow)
					So(len(g.Nodes), ShouldEqual, 2)
				})
			})
		}
	})

	Convey("Given a plain open link A -- B", t, func() {
		Convey("When parsing", func() {
			g, err := parse("graph LR\nA -- B")

			Convey("Then it is a single open edge between two nodes", func() {
				So(err, ShouldBeNil)
				So(len(g.Nodes), ShouldEqual, 2)
				So(len(g.Edges), ShouldEqual, 1)
				So(g.Edges[0].Arrow, ShouldEqual, domain.ArrowOpen)
			})
		})
	})
}

func TestSubgraphs(t *testing.T) {
	Convey("Given subgraphs with titles", t, func() {
		src := "graph TD\nsubgraph one [Group One]\nA --> B\nend\nsubgraph two\nC --> D\nend\nB --> C"

		Convey("When parsing", func() {
			g, err := parse(src)

			Convey("Then subgraphs capture their titles and members", func() {
				So(err, ShouldBeNil)
				So(len(g.Subgraphs), ShouldEqual, 2)
				So(g.Subgraphs[0].Title, ShouldEqual, "Group One")
				So(g.Subgraphs[0].NodeIDs, ShouldResemble, []string{"A", "B"})
				So(g.Subgraphs[1].Title, ShouldEqual, "two")
				So(g.Subgraphs[1].NodeIDs, ShouldResemble, []string{"C", "D"})
			})

			Convey("Then all nodes and the cross-subgraph edge exist", func() {
				So(len(g.Nodes), ShouldEqual, 4)
				So(len(g.Edges), ShouldEqual, 3)
			})
		})
	})
}

func TestStatementSeparators(t *testing.T) {
	Convey("Given semicolon-separated statements", t, func() {
		Convey("When parsing a single line and a header with inline statements", func() {
			g, err := parse("graph TD; A --> B; B --> C")

			Convey("Then all nodes and edges are parsed", func() {
				So(err, ShouldBeNil)
				So(len(g.Nodes), ShouldEqual, 3)
				So(len(g.Edges), ShouldEqual, 2)
			})
		})
	})
}

func TestFlowchartKeyword(t *testing.T) {
	Convey("Given the 'flowchart' keyword", t, func() {
		g, err := parse("flowchart LR\nA --> B")

		Convey("When parsed", func() {
			Convey("Then it behaves like 'graph' with a direction", func() {
				So(err, ShouldBeNil)
				So(g.Direction, ShouldEqual, domain.LeftRight)
			})
		})
	})
}
