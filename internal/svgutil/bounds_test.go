package svgutil

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestBounds(t *testing.T) {
	Convey("Given a fresh Bounds", t, func() {
		var b Bounds

		Convey("When nothing is added", func() {
			Convey("Then it is empty and yields no shift or size", func() {
				So(b.Empty(), ShouldBeTrue)
				dx, dy := b.Offset()
				So(dx, ShouldEqual, 0)
				So(dy, ShouldEqual, 0)
				w, h := b.Size()
				So(w, ShouldEqual, 0)
				So(h, ShouldEqual, 0)
			})
		})

		Convey("When content sits entirely at positive coordinates", func() {
			b.AddRect(10, 20, 30, 40)

			Convey("Then no shift is needed and the size reaches the far corner", func() {
				dx, dy := b.Offset()
				So(dx, ShouldEqual, 0)
				So(dy, ShouldEqual, 0)
				w, h := b.Size()
				So(w, ShouldEqual, 40)
				So(h, ShouldEqual, 60)
			})
		})

		Convey("When content reaches left of and above the origin", func() {
			b.Add(0, 0)
			b.AddRect(-25, -10, 5, 5)

			Convey("Then the shift moves it back onto the canvas", func() {
				dx, dy := b.Offset()
				So(dx, ShouldEqual, 25)
				So(dy, ShouldEqual, 10)
			})

			Convey("Then the size covers the shifted content", func() {
				w, h := b.Size()
				So(w, ShouldEqual, 25)
				So(h, ShouldEqual, 10)
			})
		})

		Convey("When a single point is added", func() {
			b.Add(7, 9)

			Convey("Then the bounds collapse onto that point", func() {
				So(b.MinX, ShouldEqual, 7)
				So(b.MaxX, ShouldEqual, 7)
				So(b.MinY, ShouldEqual, 9)
				So(b.MaxY, ShouldEqual, 9)
			})
		})
	})
}
