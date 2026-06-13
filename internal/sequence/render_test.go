package sequence

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var ropts = RenderOptions{Theme: "default", FontFace: "sans-serif", FontSize: 14, Padding: 16}

func TestRender(t *testing.T) {
	Convey("Given a sequence diagram", t, func() {
		src := "sequenceDiagram\n" +
			"participant A as Alice\n" +
			"participant B as Bob\n" +
			"A->>B: Hello\n" +
			"B-->>A: Hi"

		Convey("When rendering to SVG", func() {
			out, err := Render(src, ropts)
			svg := string(out)

			Convey("Then the document is well-formed", func() {
				So(err, ShouldBeNil)
				So(svg, ShouldStartWith, "<svg")
				So(svg, ShouldContainSubstring, "</svg>")
			})

			Convey("Then participant labels and lifelines are drawn", func() {
				So(svg, ShouldContainSubstring, ">Alice<")
				So(svg, ShouldContainSubstring, ">Bob<")
				So(svg, ShouldContainSubstring, "stroke-dasharray=\"3,3\"") // lifeline
			})

			Convey("Then messages render with an arrowhead and a label", func() {
				So(svg, ShouldContainSubstring, "marker-end=\"url(#seq-arrow)\"")
				So(svg, ShouldContainSubstring, ">Hello<")
			})

			Convey("Then the reply is dashed", func() {
				So(svg, ShouldContainSubstring, "stroke-dasharray=\"6,4\"")
			})
		})
	})

	Convey("Given a self-message", t, func() {
		Convey("When rendering", func() {
			out, err := Render("sequenceDiagram\nA->>A: retry", ropts)

			Convey("Then a loop path and its label appear", func() {
				So(err, ShouldBeNil)
				So(string(out), ShouldContainSubstring, "<path")
				So(string(out), ShouldContainSubstring, ">retry<")
			})
		})
	})

	Convey("Given a cross-ended message", t, func() {
		Convey("When rendering A-xB", func() {
			out, err := Render("sequenceDiagram\nA-xB: lost", ropts)

			Convey("Then a cross mark is drawn", func() {
				So(err, ShouldBeNil)
				So(string(out), ShouldContainSubstring, "<path")
			})
		})
	})

	Convey("Given invalid source", t, func() {
		Convey("When rendering without a header", func() {
			_, err := Render("A->>B: hi", ropts)

			Convey("Then it returns an error", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}
