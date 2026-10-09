package parser

import (
	"errors"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/zkrebbekx/go-mermaid/internal/domain"
	"github.com/zkrebbekx/go-mermaid/internal/lexer"
)

func parse(src string) (*domain.Graph, error) {
	toks, err := lexer.Lex(src)
	if err != nil {
		return nil, err
	}
	return Parse(toks)
}

func TestParse(t *testing.T) {
	Convey("Given a flowchart source", t, func() {

		Convey("When parsing a header with direction and one edge", func() {
			g, err := parse("graph LR\nA --> B")

			Convey("Then direction, nodes and edge are captured", func() {
				So(err, ShouldBeNil)
				So(g.Direction, ShouldEqual, domain.LeftRight)
				So(len(g.Nodes), ShouldEqual, 2)
				So(len(g.Edges), ShouldEqual, 1)
				So(g.Edges[0].From, ShouldEqual, "A")
				So(g.Edges[0].To, ShouldEqual, "B")
				So(g.Edges[0].Arrow, ShouldEqual, domain.ArrowNormal)
			})
		})

		Convey("When a node declares a shape and label", func() {
			g, err := parse("graph TD\nA{Decision} --> B([Done])")

			Convey("Then shape and label are recorded and merged by ID", func() {
				So(err, ShouldBeNil)
				So(g.NodeByID("A").Shape, ShouldEqual, domain.ShapeDiamond)
				So(g.NodeByID("A").Label, ShouldEqual, "Decision")
				So(g.NodeByID("B").Shape, ShouldEqual, domain.ShapeStadium)
			})
		})

		Convey("When an edge carries an inline label", func() {
			g, err := parse("graph TD\nA -->|yes| B")

			Convey("Then the edge label is set", func() {
				So(err, ShouldBeNil)
				So(g.Edges[0].Label, ShouldEqual, "yes")
			})
		})

		Convey("When the header is missing", func() {
			_, err := parse("A --> B")

			Convey("Then a ParseError with position is returned", func() {
				So(err, ShouldNotBeNil)
				var pe *ParseError
				So(errors.As(err, &pe), ShouldBeTrue)
				So(pe.Line, ShouldEqual, 1)
				So(pe.Col, ShouldEqual, 1)
			})
		})
	})
}

func TestLinkHeads(t *testing.T) {
	Convey("Given links with different ends", t, func() {
		cases := []struct {
			in         string
			arrow      domain.Arrow
			tail, head domain.Head
		}{
			{"A --> B", domain.ArrowNormal, domain.HeadNone, domain.HeadArrow},
			{"A --- B", domain.ArrowOpen, domain.HeadNone, domain.HeadNone},
			{"A --x B", domain.ArrowOpen, domain.HeadNone, domain.HeadCross},
			{"A --o B", domain.ArrowOpen, domain.HeadNone, domain.HeadCircle},
			{"A x--x B", domain.ArrowOpen, domain.HeadCross, domain.HeadCross},
			{"A o--o B", domain.ArrowOpen, domain.HeadCircle, domain.HeadCircle},
			{"A <--> B", domain.ArrowNormal, domain.HeadArrow, domain.HeadArrow},
			{"A -.- B", domain.ArrowDotted, domain.HeadNone, domain.HeadNone},
			{"A -.-> B", domain.ArrowDotted, domain.HeadNone, domain.HeadArrow},
			{"A <-.-> B", domain.ArrowDotted, domain.HeadArrow, domain.HeadArrow},
			{"A === B", domain.ArrowThick, domain.HeadNone, domain.HeadNone},
			{"A o==o B", domain.ArrowThick, domain.HeadCircle, domain.HeadCircle},
			{"A x-- text --o B", domain.ArrowOpen, domain.HeadCross, domain.HeadCircle},
			{"A <-- text --> B", domain.ArrowNormal, domain.HeadArrow, domain.HeadArrow},
			{"A -->|text| B", domain.ArrowNormal, domain.HeadNone, domain.HeadArrow},
			{"A x-- x --x B", domain.ArrowOpen, domain.HeadCross, domain.HeadCross},
		}
		for _, c := range cases {
			Convey("When parsing "+c.in, func() {
				g, err := parse("flowchart LR\n" + c.in)

				Convey("Then there are two nodes, one edge and the right ends", func() {
					So(err, ShouldBeNil)
					So(len(g.Nodes), ShouldEqual, 2)
					So(len(g.Edges), ShouldEqual, 1)
					e := g.Edges[0]
					So(e.Arrow, ShouldEqual, c.arrow)
					So(e.Tail, ShouldEqual, c.tail)
					So(e.Head, ShouldEqual, c.head)
				})
			})
		}
	})

	Convey("Given a label that is just x", t, func() {
		g, err := parse("flowchart LR\nA -- x --> B")

		Convey("When parsing", func() {
			Convey("Then x is the label, not a head or a node", func() {
				So(err, ShouldBeNil)
				So(len(g.Nodes), ShouldEqual, 2)
				So(g.Edges[0].Label, ShouldEqual, "x")
				So(g.Edges[0].Head, ShouldEqual, domain.HeadArrow)
			})
		})
	})

	Convey("Given a link ending in x or o with no target", t, func() {
		Convey("When parsing", func() {
			// As in Mermaid.js, a trailing x or o is a head, so a node named
			// x or o must be written with a space: A --- x.
			for _, src := range []string{"A --x", "A --o", "A ---o[Label]"} {
				_, err := parse("flowchart LR\n" + src)
				var pe *ParseError

				Convey("Then "+src+" reports a missing node", func() {
					So(errors.As(err, &pe), ShouldBeTrue)
					So(pe.Msg, ShouldContainSubstring, "expected node identifier")
				})
			}
		})
	})

	Convey("Given an open link followed by another link", t, func() {
		g, err := parse("flowchart LR\nA --- B --> C")

		Convey("When parsing", func() {
			Convey("Then it is a chain of two links, not a label", func() {
				So(err, ShouldBeNil)
				So(len(g.Nodes), ShouldEqual, 3)
				So(len(g.Edges), ShouldEqual, 2)
				So(g.Edges[0].Label, ShouldBeEmpty)
				So(g.Edges[1].From, ShouldEqual, "B")
			})
		})
	})
}
