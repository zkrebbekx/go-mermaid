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

func TestQuotedMiddleLabels(t *testing.T) {
	Convey("Given middle-form edge labels in quotes", t, func() {
		cases := []struct {
			in    string
			label string
			arrow domain.Arrow
		}{
			{`A -- "quoted label" --> B`, "quoted label", domain.ArrowNormal},
			{`A -. "dotted quoted" .-> B`, "dotted quoted", domain.ArrowDotted},
			{`A == "thick" ==> B`, "thick", domain.ArrowThick},
			{`A --"tight"--> B`, "tight", domain.ArrowNormal},
			{`A -- "a<br/>b" --> B`, "a<br/>b", domain.ArrowNormal},
			{`A -- "has (parens)" --> B`, "has (parens)", domain.ArrowNormal},
			{`A -- "open" --- B`, "open", domain.ArrowOpen},
			{`A -- "" --> B`, "", domain.ArrowNormal},
			// Mermaid.js rejects words around a quoted string; we join them.
			{`A -- foo "bar" baz --> B`, "foo bar baz", domain.ArrowNormal},
		}
		for _, c := range cases {
			Convey("When parsing "+c.in, func() {
				g, err := parse("flowchart LR\n" + c.in)

				Convey("Then one edge carries the unquoted label and the closing style", func() {
					So(err, ShouldBeNil)
					So(len(g.Edges), ShouldEqual, 1)
					So(g.Edges[0].Label, ShouldEqual, c.label)
					So(g.Edges[0].Arrow, ShouldEqual, c.arrow)
				})
			})
		}

		Convey("When the quote is not closed", func() {
			_, err := parse("flowchart LR\nA -- \"unterminated --> B")

			Convey("Then a ParseError points at the opening quote", func() {
				var pe *ParseError
				So(errors.As(err, &pe), ShouldBeTrue)
				So(pe.Line, ShouldEqual, 2)
				So(pe.Col, ShouldEqual, 6)
			})
		})
	})
}

func TestQuotedSubgraphTitle(t *testing.T) {
	Convey("Given a subgraph with a quoted title", t, func() {
		g, err := parse("flowchart LR\nsubgraph \"Quoted Title\"\nA --> B\nend")

		Convey("When parsing", func() {
			Convey("Then the title is the unquoted text and holds both nodes", func() {
				So(err, ShouldBeNil)
				So(len(g.Subgraphs), ShouldEqual, 1)
				So(g.Subgraphs[0].Title, ShouldEqual, "Quoted Title")
				So(g.Subgraphs[0].NodeIDs, ShouldResemble, []string{"A", "B"})
			})
		})
	})
}
