package xychart

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestYAxisTitleIsDrawn(t *testing.T) {
	Convey("Given a chart with a y-axis title", t, func() {
		src := "xychart-beta\nx-axis [a, b]\ny-axis \"Revenue USD\" 0 --> 10\nbar [1, 2]"
		out, err := Render(src, RenderOptions{Theme: "default", FontFace: "sans-serif", FontSize: 14, Padding: 16})

		Convey("When rendering", func() {
			Convey("Then the title appears, rotated into the left margin", func() {
				So(err, ShouldBeNil)
				So(string(out), ShouldContainSubstring, "Revenue USD")
				So(string(out), ShouldContainSubstring, "rotate(-90")
			})
		})
	})

	Convey("Given a chart with no y-axis title", t, func() {
		out, err := Render("xychart-beta\nx-axis [a, b]\nbar [1, 2]",
			RenderOptions{Theme: "default", FontFace: "sans-serif", FontSize: 14, Padding: 16})

		Convey("When rendering", func() {
			Convey("Then no rotated title is emitted", func() {
				So(err, ShouldBeNil)
				So(string(out), ShouldNotContainSubstring, "rotate(-90")
			})
		})
	})
}
