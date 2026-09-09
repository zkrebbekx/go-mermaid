package mindmap

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDecorationsAreNotNodes(t *testing.T) {
	Convey("Given an icon decoration under a node", t, func() {
		d, err := Parse("mindmap\nroot((r))\n  A\n  ::icon(fa fa-book)")

		Convey("When parsing", func() {
			Convey("Then it does not become a child node", func() {
				So(err, ShouldBeNil)
				So(len(d.Root.Children), ShouldEqual, 1)
				So(d.Root.Children[0].Text, ShouldEqual, "A")
			})
		})
	})

	Convey("Given a class decoration under a node", t, func() {
		d, err := Parse("mindmap\nroot((r))\n  A\n  :::urgent")

		Convey("When parsing", func() {
			Convey("Then it is skipped too", func() {
				So(err, ShouldBeNil)
				So(len(d.Root.Children), ShouldEqual, 1)
			})
		})
	})
}
