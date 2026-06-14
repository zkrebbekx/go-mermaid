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
