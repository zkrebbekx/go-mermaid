package block

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConnectorsAndSpaces(t *testing.T) {
	Convey("Given a row containing an arrow", t, func() {
		d, err := Parse("block-beta\ncolumns 3\nA --> B")

		Convey("When parsing", func() {
			Convey("Then the arrow is not turned into a block", func() {
				So(err, ShouldBeNil)
				So(len(d.Rows[0]), ShouldEqual, 2)
				So(d.Rows[0][0].Label, ShouldEqual, "A")
				So(d.Rows[0][1].Label, ShouldEqual, "B")
			})
		})
	})

	Convey("Given a row containing a space", t, func() {
		d, err := Parse("block-beta\ncolumns 3\na space b")

		Convey("When parsing", func() {
			Convey("Then the space is a gap, not a labelled block", func() {
				So(err, ShouldBeNil)
				So(len(d.Rows[0]), ShouldEqual, 3)
				So(d.Rows[0][1].Space, ShouldBeTrue)
				So(d.Rows[0][1].Label, ShouldEqual, "")
			})
		})

		Convey("When rendering", func() {
			out, rErr := Render("block-beta\ncolumns 3\na space b",
				RenderOptions{Theme: "default", FontFace: "sans-serif", FontSize: 14, Padding: 16})

			Convey("Then no block is drawn for it", func() {
				So(rErr, ShouldBeNil)
				So(string(out), ShouldNotContainSubstring, ">space<")
			})
		})
	})

	Convey("Given a multi-column space", t, func() {
		d, err := Parse("block-beta\ncolumns 3\na space:2")

		Convey("When parsing", func() {
			Convey("Then it spans the requested columns", func() {
				So(err, ShouldBeNil)
				So(d.Rows[0][1].Space, ShouldBeTrue)
				So(d.Rows[0][1].Span, ShouldEqual, 2)
			})
		})
	})
}
