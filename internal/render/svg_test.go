package render

import (
	"fmt"
	"math"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/zkrebbekx/go-mermaid/internal/layout"
	"github.com/zkrebbekx/go-mermaid/internal/lexer"
	"github.com/zkrebbekx/go-mermaid/internal/parser"
)

func laidOut(src string) *layout.Result {
	toks, err := lexer.Lex(src)
	if err != nil {
		panic(err)
	}
	g, err := parser.Parse(toks)
	if err != nil {
		panic(err)
	}
	res, err := layout.Compute(g, layout.Options{NodeSep: 50, RankSep: 50, FontSize: 14})
	if err != nil {
		panic(err)
	}
	return res
}

var opts = Options{Theme: "default", FontFace: "sans-serif", FontSize: 14, Padding: 16}

func TestSVG(t *testing.T) {
	Convey("Given a laid-out flowchart", t, func() {
		res := laidOut("graph TD\nA[Start] --> B((End))")
		out, err := SVG(res, opts)
		svg := string(out)

		Convey("Then it produces a well-formed SVG document", func() {
			So(err, ShouldBeNil)
			So(svg, ShouldStartWith, "<svg")
			So(svg, ShouldContainSubstring, "</svg>")
			So(svg, ShouldContainSubstring, "marker id=\"arrow\"")
			So(svg, ShouldContainSubstring, "<rect")   // rectangle node
			So(svg, ShouldContainSubstring, "<circle") // circle node
			So(svg, ShouldContainSubstring, "<path")   // edge
		})
	})

	Convey("Given each node shape", t, func() {
		cases := []struct{ name, src, want string }{
			{"round", "graph TD\nA(R)", "<rect"},
			{"stadium", "graph TD\nA([S])", "<rect"},
			{"diamond", "graph TD\nA{D}", "<polygon"},
		}
		for _, c := range cases {
			c := c
			Convey("When rendering the "+c.name+" shape", func() {
				out, err := SVG(laidOut(c.src), opts)

				Convey("Then the expected SVG primitive appears", func() {
					So(err, ShouldBeNil)
					So(string(out), ShouldContainSubstring, c.want)
				})
			})
		}
	})

	Convey("Given different arrow styles", t, func() {
		Convey("When the arrow is dotted", func() {
			out, _ := SVG(laidOut("graph TD\nA -.-> B"), opts)
			Convey("Then the path is dashed", func() {
				So(string(out), ShouldContainSubstring, "stroke-dasharray")
			})
		})

		Convey("When the arrow is thick", func() {
			out, _ := SVG(laidOut("graph TD\nA ==> B"), opts)
			Convey("Then the stroke width increases", func() {
				So(string(out), ShouldContainSubstring, `stroke-width="3"`)
			})
		})

		Convey("When the link is open (no head)", func() {
			out, _ := SVG(laidOut("graph TD\nA --- B"), opts)
			Convey("Then no arrowhead marker is drawn", func() {
				So(string(out), ShouldNotContainSubstring, "marker-end")
			})
		})
	})

	Convey("Given an edge label", t, func() {
		Convey("When rendering", func() {
			out, _ := SVG(laidOut("graph TD\nA -->|go| B"), opts)
			Convey("Then the label text appears", func() {
				So(string(out), ShouldContainSubstring, ">go<")
			})
		})
	})

	Convey("Given the dark theme", t, func() {
		Convey("When rendering", func() {
			out, _ := SVG(laidOut("graph TD\nA --> B"), Options{Theme: "dark", FontSize: 14, Padding: 16})
			Convey("Then the dark background color is used", func() {
				So(string(out), ShouldContainSubstring, "#1e1e1e")
			})
		})
	})

	Convey("Given an unknown theme", t, func() {
		Convey("When rendering", func() {
			out, _ := SVG(laidOut("graph TD\nA --> B"), Options{Theme: "nope", FontSize: 14})
			Convey("Then it falls back to the default palette", func() {
				So(string(out), ShouldContainSubstring, "#ffffff")
			})
		})
	})

	Convey("Given a label with XML-special characters", t, func() {
		Convey("When rendering", func() {
			res := laidOut("graph TD\nA --> B")
			res.Graph.Nodes[0].Label = `a<b & "c"`
			out, _ := SVG(res, opts)
			Convey("Then the characters are escaped", func() {
				So(string(out), ShouldContainSubstring, "a&lt;b &amp; &quot;c&quot;")
			})
		})
	})
}

func TestNum(t *testing.T) {
	Convey("Given numbers to format for SVG", t, func() {
		cases := []struct {
			in   float64
			want string
		}{
			{12, "12"},
			{12.5, "12.5"},
			{12.25, "12.25"},
			{0, "0"},
			{math.Copysign(0, -1), "0"}, // negative zero must normalize to "0"
			{1.200, "1.2"},
		}
		for _, c := range cases {
			c := c
			Convey(fmt.Sprintf("When formatting %v (want %q)", c.in, c.want), func() {
				Convey("Then trailing zeros and negative zero are normalized", func() {
					So(num(c.in), ShouldEqual, c.want)
				})
			})
		}
	})
}

func TestEscPlain(t *testing.T) {
	Convey("Given text with no special characters", t, func() {
		Convey("When escaped", func() {
			Convey("Then it is returned unchanged", func() {
				So(strings.Contains(esc("plain"), "&"), ShouldBeFalse)
			})
		})
	})
}
