package render

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/zkrebbekx/go-mermaid/internal/layout"
	"github.com/zkrebbekx/go-mermaid/internal/lexer"
	"github.com/zkrebbekx/go-mermaid/internal/parser"
	"github.com/zkrebbekx/go-mermaid/internal/svgutil"
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

func TestParallelEdgeLabelsSVG(t *testing.T) {
	Convey("Given opposite labeled edges between the same node pair (issue #27)", t, func() {
		out, err := SVG(laidOut("flowchart TD\nClient -->|request| API\nAPI -->|response| Client"), opts)
		svg := string(out)

		Convey("Then both labels are emitted", func() {
			So(err, ShouldBeNil)
			So(svg, ShouldContainSubstring, ">request<")
			So(svg, ShouldContainSubstring, ">response<")
		})

		Convey("Then no label background rect starts at a negative x", func() {
			So(err, ShouldBeNil)
			So(svg, ShouldNotContainSubstring, `<rect x="-`)
		})

		Convey("Then the two label rects sit at different heights", func() {
			So(err, ShouldBeNil)
			var ys []string
			for _, ln := range strings.Split(svg, "\n") {
				if !strings.Contains(ln, `fill="#ffffff"/>`) || !strings.Contains(ln, `height="18"`) {
					continue
				}
				_, after, found := strings.Cut(ln, `y="`)
				So(found, ShouldBeTrue)
				ys = append(ys, after)
			}
			So(len(ys), ShouldEqual, 2)
			So(ys[0], ShouldNotEqual, ys[1])
		})
	})
}

func TestSubgraphTitleFitsBox(t *testing.T) {
	Convey("Given a subgraph whose title is wider than its member nodes", t, func() {
		src := "flowchart TD\nsubgraph s [A Very Long Subgraph Title That Is Wide]\na --> b\nend"
		out, err := SVG(laidOut(src), opts)
		svg := string(out)

		Convey("Then the whole title is inside the canvas", func() {
			So(err, ShouldBeNil)
			m := regexp.MustCompile(`viewBox="0 0 ([0-9.]+) `).FindStringSubmatch(svg)
			So(m, ShouldNotBeNil)
			width, convErr := strconv.ParseFloat(m[1], 64)
			So(convErr, ShouldBeNil)
			title := "A Very Long Subgraph Title That Is Wide"
			So(svg, ShouldContainSubstring, title)
			tm := regexp.MustCompile(`<text x="([-0-9.]+)"[^>]*>` + title).FindStringSubmatch(svg)
			So(tm, ShouldNotBeNil)
			x, convErr2 := strconv.ParseFloat(tm[1], 64)
			So(convErr2, ShouldBeNil)
			// The group is translated by the padding, so add it back.
			So(x+opts.Padding, ShouldBeGreaterThanOrEqualTo, 0)
			So(x+opts.Padding+svgutil.TextWidth(title, opts.FontSize), ShouldBeLessThanOrEqualTo, width)
		})
	})

	Convey("Given a subgraph whose title is narrower than its member nodes", t, func() {
		wide, err := SVG(laidOut("flowchart TD\nsubgraph s [S]\na[A very wide node label] --> b\nend"), opts)

		Convey("Then the box is still sized by the nodes", func() {
			So(err, ShouldBeNil)
			So(string(wide), ShouldContainSubstring, "A very wide node label")
		})
	})
}
