package c4

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestStyleDirectivesAndTech(t *testing.T) {
	Convey("Given a style directive", t, func() {
		d, err := Parse("C4Context\nPerson(a, \"A\")\nUpdateElementStyle(a, $bgColor=\"red\")")

		Convey("When parsing", func() {
			Convey("Then it does not become an element", func() {
				So(err, ShouldBeNil)
				So(len(d.Elements), ShouldEqual, 1)
				So(d.Elements[0].ID, ShouldEqual, "a")
			})
		})
	})

	Convey("Given a relationship with a technology argument", t, func() {
		src := "C4Context\nPerson(a,\"A\")\nSystem(s,\"S\")\nRel(a, s, \"uses\", \"HTTPS\")"
		d, err := Parse(src)

		Convey("When parsing", func() {
			Convey("Then the technology is kept, not dropped", func() {
				So(err, ShouldBeNil)
				So(d.Rels[0].Label, ShouldEqual, "uses")
				So(d.Rels[0].Tech, ShouldEqual, "HTTPS")
			})
		})

		Convey("When rendering", func() {
			out, rErr := Render(src, RenderOptions{Theme: "default", FontFace: "sans-serif", FontSize: 14, Padding: 16})

			Convey("Then it is drawn under the label", func() {
				So(rErr, ShouldBeNil)
				So(string(out), ShouldContainSubstring, "[HTTPS]")
			})
		})
	})
}
