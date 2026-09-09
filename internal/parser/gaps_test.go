package parser

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/zkrebbekx/go-mermaid/internal/domain"
	"github.com/zkrebbekx/go-mermaid/internal/lexer"
)

func graphOf(src string) *domain.Graph {
	clean, styles, _, links := Preprocess(src)
	toks, err := lexer.Lex(clean)
	if err != nil {
		panic(err)
	}
	g, err := Parse(toks)
	if err != nil {
		panic(err)
	}
	for id, st := range styles {
		if n := g.NodeByID(id); n != nil {
			n.Style = st
		}
	}
	for i, e := range g.Edges {
		if st := links.For(i); st != nil {
			e.Style = st
		}
	}
	return g
}

func TestMultiNodeLinks(t *testing.T) {
	Convey("Given several targets joined by &", t, func() {
		g := graphOf("flowchart TD\nA --> B & C")

		Convey("When parsing", func() {
			Convey("Then one link is made to each target", func() {
				So(len(g.Edges), ShouldEqual, 2)
				So(g.Edges[0].From, ShouldEqual, "A")
				So(g.Edges[0].To, ShouldEqual, "B")
				So(g.Edges[1].From, ShouldEqual, "A")
				So(g.Edges[1].To, ShouldEqual, "C")
			})
		})
	})

	Convey("Given several sources joined by &", t, func() {
		g := graphOf("flowchart TD\nA & B --> C")

		Convey("When parsing", func() {
			Convey("Then each source links to the target", func() {
				So(len(g.Edges), ShouldEqual, 2)
				So(g.Edges[0].From, ShouldEqual, "A")
				So(g.Edges[1].From, ShouldEqual, "B")
			})
		})
	})

	Convey("Given lists on both sides", t, func() {
		g := graphOf("flowchart TD\nA & B --> C & D")

		Convey("When parsing", func() {
			Convey("Then every pair is linked", func() {
				So(len(g.Edges), ShouldEqual, 4)
			})
		})
	})

	Convey("Given a chain that uses &", t, func() {
		g := graphOf("flowchart LR\na --> b & c --> d")

		Convey("When parsing", func() {
			Convey("Then the second link starts from both middle nodes", func() {
				So(len(g.Edges), ShouldEqual, 4)
			})
		})
	})

	Convey("Given a labelled multi-node link", t, func() {
		g := graphOf("flowchart TD\nA -->|go| B & C")

		Convey("When parsing", func() {
			Convey("Then both links carry the label", func() {
				So(len(g.Edges), ShouldEqual, 2)
				So(g.Edges[0].Label, ShouldEqual, "go")
				So(g.Edges[1].Label, ShouldEqual, "go")
			})
		})
	})
}

func TestLinkStyle(t *testing.T) {
	Convey("Given a linkStyle for one edge index", t, func() {
		g := graphOf("flowchart LR\nA --> B\nA --> C\nlinkStyle 0 stroke:#ff0000,stroke-width:4px")

		Convey("When parsing", func() {
			Convey("Then only that edge is styled", func() {
				So(g.Edges[0].Style, ShouldNotBeNil)
				So(g.Edges[0].Style.Stroke, ShouldEqual, "#ff0000")
				So(g.Edges[0].Style.StrokeWidth, ShouldEqual, "4")
				So(g.Edges[1].Style, ShouldBeNil)
			})
		})
	})

	Convey("Given a linkStyle default", t, func() {
		g := graphOf("flowchart LR\nA --> B\nA --> C\nlinkStyle default stroke:#00ff00")

		Convey("When parsing", func() {
			Convey("Then every edge takes it", func() {
				So(g.Edges[0].Style.Stroke, ShouldEqual, "#00ff00")
				So(g.Edges[1].Style.Stroke, ShouldEqual, "#00ff00")
			})
		})
	})

	Convey("Given a linkStyle listing several indexes", t, func() {
		g := graphOf("flowchart LR\nA --> B\nA --> C\nA --> D\nlinkStyle 0,2 stroke:#00f")

		Convey("When parsing", func() {
			Convey("Then only the listed edges are styled", func() {
				So(g.Edges[0].Style, ShouldNotBeNil)
				So(g.Edges[1].Style, ShouldBeNil)
				So(g.Edges[2].Style, ShouldNotBeNil)
			})
		})
	})

	Convey("Given classDef with stroke width and dashes", t, func() {
		g := graphOf("flowchart LR\nA:::big --> B\nclassDef big stroke-width:5px,stroke-dasharray:3 3")

		Convey("When parsing", func() {
			Convey("Then both properties survive instead of being dropped", func() {
				n := g.NodeByID("A")
				So(n.Style, ShouldNotBeNil)
				So(n.Style.StrokeWidth, ShouldEqual, "5")
				So(n.Style.StrokeDash, ShouldEqual, "3 3")
			})
		})
	})
}

func TestIdentifiersAndShapeMetadata(t *testing.T) {
	Convey("Given identifiers containing a hyphen or a dot", t, func() {
		g := graphOf("flowchart LR\nnode-1 --> node-2\na.b --> c")

		Convey("When parsing", func() {
			Convey("Then they are single identifiers, not split at the punctuation", func() {
				So(g.NodeByID("node-1"), ShouldNotBeNil)
				So(g.NodeByID("node-2"), ShouldNotBeNil)
				So(g.NodeByID("a.b"), ShouldNotBeNil)
			})
		})
	})

	Convey("Given a link written without spaces", t, func() {
		g := graphOf("flowchart LR\nA-->B")

		Convey("When parsing", func() {
			Convey("Then the arrow is still a connector, not part of the name", func() {
				So(g.NodeByID("A"), ShouldNotBeNil)
				So(g.NodeByID("B"), ShouldNotBeNil)
				So(len(g.Edges), ShouldEqual, 1)
			})
		})
	})

	Convey("Given the Mermaid 11 shape metadata form", t, func() {
		cases := []struct {
			src   string
			shape domain.Shape
		}{
			{`flowchart LR
A@{ shape: rect, label: "Hi" }`, domain.ShapeRect},
			{`flowchart LR
A@{ shape: circle, label: "Hi" }`, domain.ShapeCircle},
			{`flowchart LR
A@{ shape: diamond, label: "Hi" }`, domain.ShapeDiamond},
			{`flowchart LR
A@{ shape: cylinder, label: "Hi" }`, domain.ShapeCylinder},
		}
		for _, c := range cases {
			c := c
			Convey("When parsing "+string(c.shape), func() {
				g := graphOf(c.src)

				Convey("Then the node takes that shape and label", func() {
					n := g.NodeByID("A")
					So(n, ShouldNotBeNil)
					So(n.Shape, ShouldEqual, c.shape)
					So(n.Label, ShouldEqual, "Hi")
				})
			})
		}

		Convey("When the shape is unknown", func() {
			g := graphOf("flowchart LR\nA@{ shape: nonesuch, label: \"Hi\" }")

			Convey("Then it falls back to a rectangle and keeps the label", func() {
				So(g.NodeByID("A").Shape, ShouldEqual, domain.ShapeRect)
				So(g.NodeByID("A").Label, ShouldEqual, "Hi")
			})
		})

		Convey("When the label contains a comma", func() {
			g := graphOf(`flowchart LR
A@{ shape: rect, label: "one, two" }`)

			Convey("Then the label is not split at it", func() {
				So(g.NodeByID("A").Label, ShouldEqual, "one, two")
			})
		})
	})
}
