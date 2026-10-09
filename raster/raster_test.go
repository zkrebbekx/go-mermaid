package raster_test

import (
	"bytes"
	"image"
	"image/png"
	"math"
	"regexp"
	"strconv"
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

// alpha returns the opacity of the pixel at p. The page itself is clear,
// because oksvg skips the percentage-sized background rect.
func alpha(img image.Image, p image.Point) uint32 {
	_, _, _, a := img.At(p.X, p.Y).RGBA()
	return a >> 8
}

func TestRoundedCorners(t *testing.T) {
	Convey("Given a hand-written SVG with corner radii", t, func() {
		svg := `<svg xmlns="http://www.w3.org/2000/svg" width="300" height="60" viewBox="0 0 300 60">` +
			`<rect x="10" y="10" width="80" height="40" rx="20" fill="#ff0000"/>` +
			`<rect x="110" y="10" width="80" height="40" ry="20" fill="#ff0000"/>` +
			`<rect x="210" y="10" width="80" height="40" fill="#ff0000"/>` +
			`</svg>`

		Convey("When rasterizing it", func() {
			out, err := raster.RasterizeSVG([]byte(svg), 1)
			So(err, ShouldBeNil)
			img, err := png.Decode(bytes.NewReader(out))
			So(err, ShouldBeNil)

			Convey("Then a lone rx or ry rounds the corners", func() {
				So(alpha(img, image.Pt(11, 11)), ShouldBeLessThan, 0x80)
				So(alpha(img, image.Pt(111, 11)), ShouldBeLessThan, 0x80)
				So(alpha(img, image.Pt(50, 30)), ShouldEqual, 0xff)
			})

			Convey("Then a rect without radii keeps square corners", func() {
				So(alpha(img, image.Pt(211, 11)), ShouldEqual, 0xff)
			})
		})
	})

	Convey("Given stadium, round and rect nodes", t, func() {
		const src = "flowchart LR\nA([stadium]) --> B(round) --> C[rect]"
		const scale = 2
		svg, err := mermaid.Render(src)
		So(err, ShouldBeNil)
		corners := nodeCorners(string(svg), scale)
		So(len(corners), ShouldEqual, 3)

		Convey("When rendering to PNG at scale 2", func() {
			out, err := raster.PNG(src, scale)
			So(err, ShouldBeNil)
			img, err := png.Decode(bytes.NewReader(out))
			So(err, ShouldBeNil)

			Convey("Then the stadium and round corners are rounded", func() {
				So(alpha(img, corners[0]), ShouldBeLessThan, 0x80)
				So(alpha(img, corners[1]), ShouldBeLessThan, 0x80)
			})

			Convey("Then the rect corner is square", func() {
				So(alpha(img, corners[2]), ShouldEqual, 0xff)
			})
		})
	})
}

var (
	translateRe = regexp.MustCompile(`<g transform="translate\(([0-9.]+),([0-9.]+)\)">`)
	nodeRectRe  = regexp.MustCompile(`<rect x="([0-9.]+)" y="([0-9.]+)"[^>]*fill="#ECECFF"`)
)

// nodeCorners returns, for each node box in document order, the output pixel
// just inside its top-left corner at the given scale.
func nodeCorners(svg string, scale float64) []image.Point {
	var gx, gy float64
	if m := translateRe.FindStringSubmatch(svg); m != nil {
		gx, _ = strconv.ParseFloat(m[1], 64)
		gy, _ = strconv.ParseFloat(m[2], 64)
	}
	px := func(v float64) int { return int(math.Floor(v*scale)) + 1 }
	var out []image.Point
	for _, m := range nodeRectRe.FindAllStringSubmatch(svg, -1) {
		x, _ := strconv.ParseFloat(m[1], 64)
		y, _ := strconv.ParseFloat(m[2], 64)
		out = append(out, image.Pt(px(x+gx), px(y+gy)))
	}
	return out
}
