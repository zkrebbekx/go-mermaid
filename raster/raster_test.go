package raster_test

import (
	"bytes"
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
