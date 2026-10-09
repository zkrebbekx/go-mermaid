package raster_test

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	mermaid "github.com/zkrebbekx/go-mermaid"
	"github.com/zkrebbekx/go-mermaid/raster"
)

func TestPNG(t *testing.T) {
	Convey("Given diagram source", t, func() {
		Convey("When rendering to PNG", func() {
			out, err := raster.PNG("graph TD\nA[Start] --> B[End]", 1)

			Convey("Then it returns a decodable PNG with positive size", func() {
				So(err, ShouldBeNil)
				So(out[:4], ShouldResemble, []byte{0x89, 'P', 'N', 'G'})
				img, derr := png.Decode(bytes.NewReader(out))
				So(derr, ShouldBeNil)
				So(img.Bounds().Dx(), ShouldBeGreaterThan, 0)
				So(img.Bounds().Dy(), ShouldBeGreaterThan, 0)
			})
		})

		Convey("When scaling 2x", func() {
			one, _ := raster.PNG("graph TD\nA --> B", 1)
			two, _ := raster.PNG("graph TD\nA --> B", 2)
			i1, _ := png.Decode(bytes.NewReader(one))
			i2, _ := png.Decode(bytes.NewReader(two))

			Convey("Then the 2x image is larger", func() {
				So(i2.Bounds().Dx(), ShouldBeGreaterThan, i1.Bounds().Dx())
			})
		})

		Convey("When the source is invalid", func() {
			_, err := raster.PNG("graph TD\nA[oops", 1)

			Convey("Then it propagates the render error", func() {
				So(err, ShouldNotBeNil)
			})
		})

		Convey("When passing options through", func() {
			out, err := raster.PNG("graph TD\nA --> B", 1, mermaid.WithTheme(mermaid.Dark))

			Convey("Then it still produces a valid PNG", func() {
				So(err, ShouldBeNil)
				_, derr := png.Decode(bytes.NewReader(out))
				So(derr, ShouldBeNil)
			})
		})
	})
}

// coverage sums the opacity down column x from y0 to y1, in pixels: a line
// of width w crossing the column scores about w however it is antialiased.
func coverage(img image.Image, x, y0, y1 int) float64 {
	var sum float64
	for y := y0; y <= y1; y++ {
		_, _, _, a := img.At(x, y).RGBA()
		sum += float64(a) / 0xffff
	}
	return sum
}

// runs returns the lengths of the alternating opaque and clear stretches
// along row y, from x0 to x1, starting at the first opaque pixel.
func runs(img image.Image, y, x0, x1 int) []int {
	var out []int
	for x := x0; x <= x1; x++ {
		_, _, _, a := img.At(x, y).RGBA()
		on := a > 0x8000
		// Odd-length out means the last run is an opaque one.
		lastOn := len(out)%2 == 1
		switch {
		case len(out) == 0 && !on:
		case on == lastOn:
			out[len(out)-1]++
		default:
			out = append(out, 1)
		}
	}
	return out
}

func TestStrokesScale(t *testing.T) {
	Convey("Given a 2-unit line, a dashed line and a line with no width", t, func() {
		svg := `<svg xmlns="http://www.w3.org/2000/svg" width="100" height="40" viewBox="0 0 100 40">` +
			`<path d="M0,10 L100,10" stroke="#000000" stroke-width="2"/>` +
			`<path d="M2,30 L98,30" stroke="#000000" stroke-width="2" stroke-dasharray="4,4"/>` +
			`<path d="M0,20 L100,20" stroke="#000000"/>` +
			`</svg>`

		for _, scale := range []float64{1, 2, 4} {
			Convey(fmt.Sprintf("When rasterizing at scale %g", scale), func() {
				out, err := raster.RasterizeSVG([]byte(svg), scale)
				So(err, ShouldBeNil)
				img, err := png.Decode(bytes.NewReader(out))
				So(err, ShouldBeNil)
				px := func(v float64) int { return int(v * scale) }

				Convey("Then the line is 2 units wide in output pixels", func() {
					So(coverage(img, px(50), px(5), px(15)), ShouldAlmostEqual, 2*scale, 0.5)
				})

				Convey("Then a line with no width uses SVG's default of 1 unit", func() {
					So(coverage(img, px(50), px(16), px(24)), ShouldAlmostEqual, 1*scale, 0.5)
				})

				Convey("Then the dashes are 4 units on and 4 off", func() {
					r := runs(img, px(30), px(2), px(98))
					So(len(r), ShouldBeGreaterThan, 4)
					for _, n := range r[:4] {
						So(float64(n), ShouldAlmostEqual, 4*scale, 1)
					}
				})
			})
		}
	})
}
