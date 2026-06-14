// Package raster rasterizes go-mermaid SVG output to PNG. It is a separate
// package so the core mermaid library stays dependency-free: only programs
// that need PNG pull in the SVG rasterizer.
//
//	png, err := raster.PNG("graph TD\n A --> B", mermaid.WithTheme(mermaid.Dark))
package raster

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"regexp"
	"strconv"

	mermaid "github.com/Zac300/go-mermaid"
	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

// PNG renders Mermaid source to SVG and rasterizes it to a PNG image. The
// scale multiplies the SVG's intrinsic pixel size (1 = same size, 2 = 2x).
func PNG(src string, scale float64, opts ...mermaid.Option) ([]byte, error) {
	svg, err := mermaid.Render(src, opts...)
	if err != nil {
		return nil, err
	}
	return RasterizeSVG(svg, scale)
}

// RasterizeSVG converts an SVG document to PNG bytes at the given scale.
func RasterizeSVG(svg []byte, scale float64) ([]byte, error) {
	if scale <= 0 {
		scale = 1
	}
	icon, err := oksvg.ReadIconStream(bytes.NewReader(svg))
	if err != nil {
		return nil, fmt.Errorf("rasterize: parse svg: %w", err)
	}
	w := int(icon.ViewBox.W * scale)
	h := int(icon.ViewBox.H * scale)
	if w <= 0 || h <= 0 {
		return nil, fmt.Errorf("rasterize: svg has no size")
	}
	icon.SetTarget(0, 0, float64(w), float64(h))

	img := image.NewRGBA(image.Rect(0, 0, w, h))
	scanner := rasterx.NewScannerGV(w, h, img, img.Bounds())
	raster := rasterx.NewDasher(w, h, scanner)
	icon.Draw(raster, 1.0)

	// oksvg ignores <text>; draw labels ourselves.
	drawText(img, string(svg), scale, rootFontSize(svg))

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("rasterize: encode png: %w", err)
	}
	return buf.Bytes(), nil
}

var rootSizeRe = regexp.MustCompile(`<svg\b[^>]*\bfont-size="([0-9.]+)"`)

// rootFontSize reads the root <svg> font-size, defaulting to 14.
func rootFontSize(svg []byte) float64 {
	if m := rootSizeRe.FindSubmatch(svg); m != nil {
		if v, err := strconv.ParseFloat(string(m[1]), 64); err == nil && v > 0 {
			return v
		}
	}
	return 14
}
