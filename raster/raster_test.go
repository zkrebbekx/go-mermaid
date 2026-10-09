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

func TestLinkHeadMarkers(t *testing.T) {
	Convey("Given lines ending in each marker", t, func() {
		line := func(y int, start, end string) string {
			return fmt.Sprintf(`<path d="M10,%d L90,%d" stroke="#000000" stroke-width="1.5" marker-start="url(#%s)" marker-end="url(#%s)"/>`, y, y, start, end)
		}
		svg := `<svg xmlns="http://www.w3.org/2000/svg" width="100" height="60" viewBox="0 0 100 60">` +
			line(10, "circle", "circle") + line(30, "cross", "cross") + line(50, "arrow", "arrow") +
			`</svg>`
		// Curved edges repeat each end point as a control point.
		curved := `<svg xmlns="http://www.w3.org/2000/svg" width="100" height="20" viewBox="0 0 100 20">` +
			`<path d="M10,10 Q10,10 50,10 Q90,10 90,10" stroke="#000000" stroke-width="1.5" marker-start="url(#circle)" marker-end="url(#circle)"/>` +
			`</svg>`

		Convey("When rasterizing it", func() {
			out, err := raster.RasterizeSVG([]byte(svg), 1)
			So(err, ShouldBeNil)
			img, err := png.Decode(bytes.NewReader(out))
			So(err, ShouldBeNil)

			Convey("Then a circle is drawn 4 units back from both ends", func() {
				So(opacityAt(img, image.Pt(85, 13)), ShouldEqual, 0xff)
				So(opacityAt(img, image.Pt(14, 7)), ShouldEqual, 0xff)
			})

			Convey("Then a cross's arms reach off the line at both ends", func() {
				So(opacityAt(img, image.Pt(83, 33)), ShouldBeGreaterThan, 0x80)
				So(opacityAt(img, image.Pt(16, 27)), ShouldBeGreaterThan, 0x80)
			})

			Convey("Then a curved path gets markers at both ends", func() {
				out, err := raster.RasterizeSVG([]byte(curved), 1)
				So(err, ShouldBeNil)
				img, err := png.Decode(bytes.NewReader(out))
				So(err, ShouldBeNil)
				So(opacityAt(img, image.Pt(85, 13)), ShouldEqual, 0xff)
				So(opacityAt(img, image.Pt(14, 13)), ShouldEqual, 0xff)
			})

			Convey("Then an arrowhead fills the line near both tips", func() {
				So(opacityAt(img, image.Pt(84, 51)), ShouldEqual, 0xff)
				So(opacityAt(img, image.Pt(15, 48)), ShouldEqual, 0xff)
			})
		})
	})
}

// opacityAt returns the alpha of the pixel at p; the page itself is clear.
func opacityAt(img image.Image, p image.Point) uint32 {
	_, _, _, a := img.At(p.X, p.Y).RGBA()
	return a >> 8
}
