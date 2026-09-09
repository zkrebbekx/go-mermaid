package svgutil

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFaceFor(t *testing.T) {
	Convey("Given CSS font-family values", t, func() {
		cases := []struct {
			in   string
			want Face
		}{
			{"sans-serif", FaceSans},
			{"Helvetica", FaceSans},
			{"", FaceSans},
			{"Nonesuch Display", FaceSans},
			{"monospace", FaceMono},
			{"Courier New", FaceMono},
			{"serif", FaceSerif},
			{"Georgia, serif", FaceSerif},
			{`"Times New Roman", serif`, FaceSerif},
			{"  MONOSPACE  ", FaceMono},
		}
		for _, c := range cases {
			c := c
			Convey("When mapping "+c.in, func() {
				Convey("Then it picks the matching metric table", func() {
					So(FaceFor(c.in), ShouldEqual, c.want)
				})
			})
		}
	})
}

func TestFaceWidth(t *testing.T) {
	Convey("Given a fixed-pitch face", t, func() {
		Convey("When measuring narrow and wide ASCII", func() {
			Convey("Then every glyph takes the same room", func() {
				So(FaceMono.Width("iiiii", 14), ShouldAlmostEqual, 42, 0.01)
				So(FaceMono.Width("WWWWW", 14), ShouldAlmostEqual, 42, 0.01)
			})
		})
	})

	Convey("Given a proportional face", t, func() {
		Convey("When measuring narrow and wide ASCII", func() {
			Convey("Then wide glyphs reserve more room", func() {
				So(FaceSans.Width("WWWWW", 14), ShouldBeGreaterThan, FaceSans.Width("iiiii", 14))
			})
		})

		Convey("When comparing serif to sans for the same text", func() {
			Convey("Then Times measures narrower than Helvetica", func() {
				So(FaceSerif.Width("Hello World", 14), ShouldBeLessThan, FaceSans.Width("Hello World", 14))
			})
		})
	})

	Convey("Given non-Latin text", t, func() {
		Convey("When measuring CJK characters", func() {
			Convey("Then each takes a full em, not the 0.6 em fallback", func() {
				So(FaceSans.Width("日本語", 14), ShouldAlmostEqual, 42, 0.01)
				So(FaceSans.Width("한국어", 14), ShouldAlmostEqual, 42, 0.01)
			})
		})

		Convey("When measuring a combining mark", func() {
			Convey("Then the mark adds no width of its own", func() {
				So(FaceSans.Width("é", 14), ShouldAlmostEqual, FaceSans.Width("e", 14), 0.01)
			})
		})

		Convey("When measuring an emoji joined by zero-width joiners", func() {
			Convey("Then the sequence counts as one glyph", func() {
				one := FaceSans.Width("\U0001F468", 14)
				family := FaceSans.Width("\U0001F468\u200d\U0001F469\u200d\U0001F467", 14)
				So(family, ShouldAlmostEqual, one, 0.01)
			})
		})

		Convey("When measuring an emoji with a skin tone modifier", func() {
			Convey("Then the modifier adds no width", func() {
				So(FaceSans.Width("\U0001F44D\U0001F3FD", 14), ShouldAlmostEqual, FaceSans.Width("\U0001F44D", 14), 0.01)
			})
		})
	})

	Convey("Given an empty string", t, func() {
		Convey("Then every face measures zero", func() {
			So(FaceSans.Width("", 14), ShouldEqual, 0)
			So(FaceMono.Width("", 14), ShouldEqual, 0)
			So(FaceSerif.Width("", 14), ShouldEqual, 0)
		})
	})
}
