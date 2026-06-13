package pie

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestParse(t *testing.T) {
	Convey("Given pie source with a title and slices", t, func() {
		src := "pie title Pets\n\"Dogs\" : 386\n\"Cats\" : 85"

		Convey("When parsing", func() {
			d, err := Parse(src)

			Convey("Then the title and slices are captured", func() {
				So(err, ShouldBeNil)
				So(d.Title, ShouldEqual, "Pets")
				So(len(d.Slices), ShouldEqual, 2)
				So(d.Slices[0].Label, ShouldEqual, "Dogs")
				So(d.Slices[0].Value, ShouldEqual, 386)
				So(d.Total(), ShouldEqual, 471)
			})
		})
	})

	Convey("Given pie with showData and no quotes", t, func() {
		Convey("When parsing", func() {
			d, err := Parse("pie showData\nApples : 10\nPears : 5")

			Convey("Then showData is ignored and slices parse", func() {
				So(err, ShouldBeNil)
				So(len(d.Slices), ShouldEqual, 2)
				So(d.Slices[1].Label, ShouldEqual, "Pears")
			})
		})
	})

	Convey("Given invalid pie source", t, func() {
		cases := map[string]string{
			"missing header": "\"Dogs\" : 1",
			"bad value":      "pie\n\"Dogs\" : abc",
			"no colon":       "pie\nDogs 5",
			"negative":       "pie\n\"Dogs\" : -3",
		}
		for name, src := range cases {
			src := src
			Convey("When parsing the "+name+" case", func() {
				_, err := Parse(src)

				Convey("Then it returns an error", func() {
					So(err, ShouldNotBeNil)
				})
			})
		}
	})
}

func TestRender(t *testing.T) {
	Convey("Given a pie chart", t, func() {
		Convey("When rendering", func() {
			out, err := Render("pie title Fruit\n\"Apples\" : 70\n\"Pears\" : 30",
				RenderOptions{Theme: "default", FontFace: "sans-serif", FontSize: 14, Padding: 16})
			svg := string(out)

			Convey("Then it draws wedges, a legend, and the title", func() {
				So(err, ShouldBeNil)
				So(svg, ShouldStartWith, "<svg")
				So(svg, ShouldContainSubstring, "<path")
				So(svg, ShouldContainSubstring, ">Fruit<")
				So(svg, ShouldContainSubstring, "Apples: 70 (70.0%)")
			})
		})
	})

	Convey("Given a single-slice pie", t, func() {
		Convey("When rendering", func() {
			out, err := Render("pie\n\"Only\" : 1",
				RenderOptions{Theme: "default", FontSize: 14, Padding: 16})

			Convey("Then it draws a full circle", func() {
				So(err, ShouldBeNil)
				So(string(out), ShouldContainSubstring, "<circle")
			})
		})
	})
}
