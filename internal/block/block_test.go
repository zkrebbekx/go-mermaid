package block

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestParse(t *testing.T) {
	Convey("Given a block diagram with columns and spans", t, func() {
		src := "block-beta\ncolumns 3\na b c\nd[\"wide block\"]:2 e"

		Convey("When parsing", func() {
			d, err := Parse(src)

			Convey("Then columns and rows parse", func() {
				So(err, ShouldBeNil)
				So(d.Columns, ShouldEqual, 3)
				So(len(d.Rows), ShouldEqual, 2)
				So(len(d.Rows[0]), ShouldEqual, 3)
			})

			Convey("Then a labeled spanning block parses", func() {
				wide := d.Rows[1][0]
				So(wide.ID, ShouldEqual, "d")
				So(wide.Label, ShouldEqual, "wide block")
				So(wide.Span, ShouldEqual, 2)
			})
		})
	})

	Convey("Given no explicit columns", t, func() {
		Convey("When parsing", func() {
			d, err := Parse("block-beta\na b c d")

			Convey("Then columns default to the widest row", func() {
				So(err, ShouldBeNil)
				So(d.Columns, ShouldEqual, 4)
			})
		})
	})

	Convey("Given no header", t, func() {
		Convey("When parsing", func() {
			_, err := Parse("columns 2")

			Convey("Then it returns an error", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestRender(t *testing.T) {
	Convey("Given a block diagram, when rendering", t, func() {
		out, err := Render("block-beta\ncolumns 2\na b\nc[\"wide\"]:2",
			RenderOptions{Theme: "default", FontSize: 14, Padding: 16})
		svg := string(out)

		Convey("Then it draws block cells and labels", func() {
			So(err, ShouldBeNil)
			So(svg, ShouldStartWith, "<svg")
			So(svg, ShouldContainSubstring, ">a<")
			So(svg, ShouldContainSubstring, ">wide<")
		})
	})
}
