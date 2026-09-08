package domain

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestPolylineMidpoint(t *testing.T) {
	Convey("Given an L-shaped polyline", t, func() {
		pts := []Point{{X: 0, Y: 0}, {X: 0, Y: 10}, {X: 10, Y: 10}}

		Convey("When taking the midpoint", func() {
			mid := PolylineMidpoint(pts)

			Convey("Then it lands on the path at half the total length, not the chord", func() {
				So(mid.X, ShouldEqual, 0)
				So(mid.Y, ShouldEqual, 10)
			})
		})
	})

	Convey("Given a straight two-point segment", t, func() {
		mid := PolylineMidpoint([]Point{{X: 0, Y: 0}, {X: 10, Y: 0}})
		Convey("Then the midpoint is the centre", func() {
			So(mid.X, ShouldEqual, 5)
			So(mid.Y, ShouldEqual, 0)
		})
	})

	Convey("Given an empty polyline", t, func() {
		Convey("Then it returns the zero point without panicking", func() {
			So(PolylineMidpoint(nil), ShouldResemble, Point{})
		})
	})
}

func TestPolylinePointAt(t *testing.T) {
	Convey("Given an L-shaped polyline of length 20", t, func() {
		pts := []Point{{X: 0, Y: 0}, {X: 0, Y: 10}, {X: 10, Y: 10}}

		Convey("When measuring its length", func() {
			Convey("Then it sums the segments", func() {
				So(PolylineLength(pts), ShouldEqual, 20)
			})
		})

		Convey("When taking the point at a quarter of the length", func() {
			p := PolylinePointAt(pts, 5)

			Convey("Then it lies on the first segment", func() {
				So(p.X, ShouldEqual, 0)
				So(p.Y, ShouldEqual, 5)
			})
		})

		Convey("When taking the point at three quarters of the length", func() {
			p := PolylinePointAt(pts, 15)

			Convey("Then it lies on the second segment", func() {
				So(p.X, ShouldEqual, 5)
				So(p.Y, ShouldEqual, 10)
			})
		})

		Convey("When the distance is negative or past the end", func() {
			Convey("Then it clamps to the endpoints", func() {
				So(PolylinePointAt(pts, -3), ShouldResemble, pts[0])
				So(PolylinePointAt(pts, 99), ShouldResemble, pts[2])
			})
		})
	})

	Convey("Given an empty polyline", t, func() {
		Convey("Then the length is zero and the point is the origin", func() {
			So(PolylineLength(nil), ShouldEqual, 0)
			So(PolylinePointAt(nil, 5), ShouldResemble, Point{})
		})
	})
}
