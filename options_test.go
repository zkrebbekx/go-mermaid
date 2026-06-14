package mermaid_test

import (
	"strconv"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	mermaid "github.com/zkrebbekx/go-mermaid"
)

func TestOptions(t *testing.T) {
	const src = "graph TD\nA --> B"

	Convey("Given Render options", t, func() {

		Convey("When setting font face and size", func() {
			out, err := mermaid.Render(src, mermaid.WithFont("Inter", 20))
			So(err, ShouldBeNil)
			So(string(out), ShouldContainSubstring, `font-family="Inter"`)
			So(string(out), ShouldContainSubstring, `font-size="20"`)
		})

		Convey("When setting padding", func() {
			out, err := mermaid.Render(src, mermaid.WithPadding(40))
			So(err, ShouldBeNil)
			So(string(out), ShouldContainSubstring, "translate(40,40)")
		})

		Convey("When increasing rank spacing", func() {
			tight, _ := mermaid.Render(src, mermaid.WithSpacing(50, 20))
			loose, _ := mermaid.Render(src, mermaid.WithSpacing(50, 200))
			So(height(loose), ShouldBeGreaterThan, height(tight))
		})

		Convey("When enabling curved edges on a graph that needs bends", func() {
			out, err := mermaid.Render("graph TD\nA-->B\nA-->C", mermaid.WithCurvedEdges(true))

			Convey("Then edges use quadratic curve commands", func() {
				So(err, ShouldBeNil)
				So(string(out), ShouldContainSubstring, "Q")
			})
		})

		Convey("When setting a custom background", func() {
			out, err := mermaid.Render(src, mermaid.WithBackground("#101820"))

			Convey("Then the background rect uses that color", func() {
				So(err, ShouldBeNil)
				So(string(out), ShouldContainSubstring, `fill="#101820"`)
			})
		})

		Convey("When requesting a transparent background", func() {
			out, err := mermaid.Render(src, mermaid.WithTransparentBackground())

			Convey("Then no full-canvas background rect is emitted", func() {
				So(err, ShouldBeNil)
				So(string(out), ShouldNotContainSubstring, `width="100%" height="100%"`)
			})
		})

		Convey("When using the neutral theme", func() {
			out, err := mermaid.Render(src, mermaid.WithTheme(mermaid.Neutral))
			So(err, ShouldBeNil)
			So(string(out), ShouldContainSubstring, "#eeeeee")
		})

		Convey("When using a custom theme", func() {
			out, err := mermaid.Render(src, mermaid.WithCustomTheme("brand", mermaid.Palette{
				Background: "#102030", NodeFill: "#204060", NodeStroke: "#88aaff",
				Text: "#ffffff", Edge: "#aaccff",
			}))

			Convey("Then the custom palette colors are used", func() {
				So(err, ShouldBeNil)
				So(string(out), ShouldContainSubstring, "#204060")
				So(string(out), ShouldContainSubstring, "#102030")
			})
		})
	})
}

// height extracts the SVG height attribute as a number.
func height(svg []byte) float64 {
	s := string(svg)
	i := strings.Index(s, `height="`)
	if i < 0 {
		return 0
	}
	s = s[i+len(`height="`):]
	v, _ := strconv.ParseFloat(s[:strings.IndexByte(s, '"')], 64)
	return v
}
