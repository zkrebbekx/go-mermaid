package pie

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTitleOnItsOwnLine(t *testing.T) {
	Convey("Given a title on the line after the header", t, func() {
		d, err := Parse("pie\ntitle Pets adopted\n\"Dogs\" : 386\n\"Cats\" : 85")

		Convey("When parsing", func() {
			Convey("Then it is read instead of failing as a slice", func() {
				So(err, ShouldBeNil)
				So(d.Title, ShouldEqual, "Pets adopted")
				So(len(d.Slices), ShouldEqual, 2)
			})
		})
	})

	Convey("Given a title on the header line", t, func() {
		d, err := Parse("pie title Pets adopted\n\"Dogs\" : 1")

		Convey("When parsing", func() {
			Convey("Then it still works", func() {
				So(err, ShouldBeNil)
				So(d.Title, ShouldEqual, "Pets adopted")
			})
		})
	})

	Convey("Given showData on the header", t, func() {
		d, err := Parse("pie showData title T\n\"Dogs\" : 1")

		Convey("When parsing", func() {
			Convey("Then the flag is recorded and the title still reads", func() {
				So(err, ShouldBeNil)
				So(d.ShowData, ShouldBeTrue)
				So(d.Title, ShouldEqual, "T")
			})
		})
	})
}
