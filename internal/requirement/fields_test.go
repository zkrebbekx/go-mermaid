package requirement

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRequirementFieldsAreDrawn(t *testing.T) {
	Convey("Given a requirement with a text body", t, func() {
		src := "requirementDiagram\nrequirement r {\nid: 1\ntext: the important requirement text\nrisk: high\nverifymethod: test\n}"
		out, err := Render(src, RenderOptions{Theme: "default", FontFace: "sans-serif", FontSize: 14, Padding: 16})

		Convey("When rendering", func() {
			Convey("Then the text body appears, not only id and risk", func() {
				So(err, ShouldBeNil)
				So(string(out), ShouldContainSubstring, "text:")
				So(string(out), ShouldContainSubstring, "important")
				So(string(out), ShouldContainSubstring, "verifymethod: test")
			})
		})
	})

	Convey("Given a long field value", t, func() {
		Convey("When wrapping it", func() {
			rows := wrapField("text", "one two three four five six seven eight nine ten")

			Convey("Then it is split over several rows, the first one labelled", func() {
				So(len(rows), ShouldBeGreaterThan, 1)
				So(rows[0], ShouldStartWith, "text: ")
				for _, r := range rows {
					So(len([]rune(r)), ShouldBeLessThanOrEqualTo, fieldWrap+8)
				}
			})
		})

		Convey("When the value is empty", func() {
			Convey("Then it yields no rows", func() {
				So(wrapField("text", "   "), ShouldBeNil)
			})
		})
	})
}
