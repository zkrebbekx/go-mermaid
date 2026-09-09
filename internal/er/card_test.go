package er

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCardinalityIsDrawn(t *testing.T) {
	Convey("Given a crow's-foot relationship", t, func() {
		out, err := Render("erDiagram\nCUSTOMER ||--o{ ORDER : places",
			RenderOptions{Theme: "default", FontFace: "sans-serif", FontSize: 14, Padding: 16})

		Convey("When rendering", func() {
			Convey("Then the cardinality text appears, not just the crow's foot", func() {
				So(err, ShouldBeNil)
				So(string(out), ShouldContainSubstring, ">1<")
				So(string(out), ShouldContainSubstring, ">0..N<")
			})
		})
	})
}
