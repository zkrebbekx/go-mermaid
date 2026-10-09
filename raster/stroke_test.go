package raster

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDefaultStrokeWidth(t *testing.T) {
	Convey("Given root svg tags", t, func() {
		cases := []struct {
			name, in, want string
		}{
			{"no stroke width", `<svg width="1"><rect/></svg>`, `<svg stroke-width="1" width="1"><rect/></svg>`},
			{"no attributes", `<svg><rect/></svg>`, `<svg stroke-width="1"><rect/></svg>`},
			{"its own width", `<svg stroke-width="3"><rect/></svg>`, `<svg stroke-width="3"><rect/></svg>`},
			{"width in style", `<svg style="stroke-width:3"><rect/></svg>`, `<svg style="stroke-width:3"><rect/></svg>`},
			{"after a prolog", `<?xml version="1.0"?><svg>`, `<?xml version="1.0"?><svg stroke-width="1">`},
			{"namespaced root", `<svg:svg>`, `<svg:svg>`},
			{"look-alike attribute", `<svg data-stroke-width="2">`, `<svg stroke-width="1" data-stroke-width="2">`},
			{"commented-out tag first", `<!-- <svg> --><svg>`, `<!-- <svg> --><svg stroke-width="1">`},
			{"spaces around =", `<svg stroke-width = "3">`, `<svg stroke-width = "3">`},
			{"no root", `<g/>`, `<g/>`},
		}
		for _, c := range cases {
			Convey("When the tag has "+c.name, func() {
				got := string(defaultStrokeWidth([]byte(c.in)))

				Convey("Then only a missing width is set to 1", func() {
					So(got, ShouldEqual, c.want)
				})
			})
		}
	})
}
