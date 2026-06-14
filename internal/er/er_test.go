package er

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestParse(t *testing.T) {
	Convey("Given an ER diagram with a relationship and attributes", t, func() {
		src := "erDiagram\nCUSTOMER ||--o{ ORDER : places\nCUSTOMER {\nstring name\nstring email\n}"

		Convey("When parsing", func() {
			d, err := Parse(src)

			Convey("Then entities, attributes, and the relationship are captured", func() {
				So(err, ShouldBeNil)
				So(len(d.Entities), ShouldEqual, 2)
				So(d.entity("CUSTOMER").Attributes, ShouldResemble, []string{"string name", "string email"})
				So(len(d.Relationships), ShouldEqual, 1)
				So(d.Relationships[0].Label, ShouldEqual, "places")
			})

			Convey("Then cardinalities are decoded", func() {
				r := d.Relationships[0]
				So(r.LeftCard, ShouldEqual, "1")
				So(r.RightCard, ShouldEqual, "0..N")
			})
		})
	})

	Convey("Given each cardinality token", t, func() {
		cases := []struct {
			tok  string
			want string
		}{
			{"||", "1"},
			{"o{", "0..N"},
			{"|{", "1..N"},
			{"o|", "0..1"},
		}
		for _, c := range cases {
			c := c
			Convey("When mapping "+c.tok, func() {
				Convey("Then the label is "+c.want, func() {
					So(cardLabel(c.tok), ShouldEqual, c.want)
				})
			})
		}
	})

	Convey("Given a non-identifying relationship", t, func() {
		Convey("When parsing with ..", func() {
			d, err := Parse("erDiagram\nA ||..|| B : has")

			Convey("Then it is marked dashed", func() {
				So(err, ShouldBeNil)
				So(d.Relationships[0].Dashed, ShouldBeTrue)
			})
		})
	})

	Convey("Given source without the header", t, func() {
		Convey("When parsing", func() {
			_, err := Parse("CUSTOMER ||--o{ ORDER")

			Convey("Then it returns an error", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestRender(t *testing.T) {
	Convey("Given an ER diagram", t, func() {
		Convey("When rendering", func() {
			out, err := Render("erDiagram\nCUSTOMER ||--o{ ORDER : places\nORDER {\nint id\n}",
				RenderOptions{Theme: "default", FontFace: "sans-serif", FontSize: 14, Padding: 16})
			svg := string(out)

			Convey("Then it draws entity boxes, attributes, and crow's-foot markers", func() {
				So(err, ShouldBeNil)
				So(svg, ShouldStartWith, "<svg")
				So(svg, ShouldContainSubstring, ">CUSTOMER<")
				So(svg, ShouldContainSubstring, ">ORDER<")
				So(svg, ShouldContainSubstring, "int id")
				So(svg, ShouldContainSubstring, "<circle") // zero-or-many crow's foot
			})
		})
	})
}
