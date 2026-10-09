package raster

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFillRadii(t *testing.T) {
	Convey("Given rect tags with some corner radii", t, func() {
		cases := []struct {
			name, in, want string
		}{
			{"rx only", `<rect x="0" rx="21"/>`, `<rect ry="21" x="0" rx="21"/>`},
			{"ry only", `<rect x="0" ry="6"/>`, `<rect rx="6" x="0" ry="6"/>`},
			{"single quotes", `<rect rx='4'/>`, `<rect ry="4" rx='4'/>`},
			{"both set", `<rect rx="4" ry="2"/>`, `<rect rx="4" ry="2"/>`},
			{"neither", `<rect x="0" width="5"/>`, `<rect x="0" width="5"/>`},
			{"look-alike attribute", `<rect data-rx="4"/>`, `<rect data-rx="4"/>`},
			{"multi-line tag", "<rect\n x=\"0\"\n rx=\"4\"/>", "<rect ry=\"4\"\n x=\"0\"\n rx=\"4\"/>"},
			{"spaces around =", `<rect rx = "4"/>`, `<rect ry="4" rx = "4"/>`},
			{"closing tag", `<rect rx="4"></rect>`, `<rect ry="4" rx="4"></rect>`},
			{"custom element", `<rect-foo rx="4"/>`, `<rect-foo rx="4"/>`},
			{"other elements", `<ellipse rx="4"/><path d="M0,0"/>`, `<ellipse rx="4"/><path d="M0,0"/>`},
		}
		for _, c := range cases {
			Convey("When the tag has "+c.name, func() {
				got := string(fillRadii([]byte(c.in)))

				Convey("Then the missing radius copies the given one", func() {
					So(got, ShouldEqual, c.want)
				})
			})
		}
	})
}
