package sankey

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestParse(t *testing.T) {
	Convey("Given a sankey-beta diagram", t, func() {
		src := "sankey-beta\nA,B,10\nB,C,4\nB,D,6"

		Convey("When parsing", func() {
			d, err := Parse(src)

			Convey("Then flows and first-seen node order parse", func() {
				So(err, ShouldBeNil)
				So(len(d.Flows), ShouldEqual, 3)
				So(d.Flows[0].Source, ShouldEqual, "A")
				So(d.Flows[0].Value, ShouldEqual, 10)
				So(d.Nodes, ShouldResemble, []string{"A", "B", "C", "D"})
			})
		})
	})

	Convey("Given a quoted field with a comma", t, func() {
		Convey("When parsing", func() {
			d, err := Parse("sankey-beta\n\"A, Inc\",B,5")

			Convey("Then the quoted source is preserved", func() {
				So(err, ShouldBeNil)
				So(d.Flows[0].Source, ShouldEqual, "A, Inc")
			})
		})
	})

	Convey("Given an invalid value", t, func() {
		Convey("When parsing", func() {
			_, err := Parse("sankey-beta\nA,B,x")

			Convey("Then it returns an error", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Given no header", t, func() {
		Convey("When parsing", func() {
			_, err := Parse("A,B,1")

			Convey("Then it returns an error", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestRender(t *testing.T) {
	Convey("Given a sankey diagram, when rendering", t, func() {
		out, err := Render("sankey-beta\nA,B,10\nB,C,5\nB,D,5",
			RenderOptions{Theme: "default", FontSize: 14, Padding: 16})
		svg := string(out)

		Convey("Then it draws node bars, flow bands, and labels", func() {
			So(err, ShouldBeNil)
			So(svg, ShouldStartWith, "<svg")
			So(svg, ShouldContainSubstring, "fill-opacity=\"0.4\"") // band
			So(svg, ShouldContainSubstring, ">A<")
		})
	})
}
