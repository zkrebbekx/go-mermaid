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
	"math"
	"regexp"
	"strconv"

	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
	mermaid "github.com/zkrebbekx/go-mermaid"
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
	icon, err := oksvg.ReadIconStream(bytes.NewReader(defaultStrokeWidth(svg)))
	if err != nil {
		return nil, fmt.Errorf("rasterize: parse svg: %w", err)
	}
	w := int(icon.ViewBox.W * scale)
	h := int(icon.ViewBox.H * scale)
	if w <= 0 || h <= 0 {
		return nil, fmt.Errorf("rasterize: svg has no size")
	}
	icon.SetTarget(0, 0, float64(w), float64(h))
	scaleStrokes(icon)

	img := image.NewRGBA(image.Rect(0, 0, w, h))
	scanner := rasterx.NewScannerGV(w, h, img, img.Bounds())
	raster := rasterx.NewDasher(w, h, scanner)
	icon.Draw(raster, 1.0)

	// oksvg ignores markers and <text>; draw arrowheads and labels ourselves.
	drawArrows(img, string(svg), scale)
	drawText(img, string(svg), scale, rootFontSize(svg))

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("rasterize: encode png: %w", err)
	}
	return buf.Bytes(), nil
}

var (
	rootTagRe     = regexp.MustCompile(`<svg(?:\s[^>]*)?>`)
	commentRe     = regexp.MustCompile(`(?s)<!--.*?-->`)
	strokeWidthRe = regexp.MustCompile(`(?:^|[\s;"'])stroke-width\s*[=:]`)
)

// defaultStrokeWidth gives the root <svg> the SVG initial stroke-width of 1
// when it sets none. oksvg's own default is 2, which drew every border
// without an explicit width twice as heavy as a browser does.
func defaultStrokeWidth(svg []byte) []byte {
	loc := rootTag(svg)
	if loc == nil || strokeWidthRe.Match(svg[loc[0]+len("<svg"):loc[1]]) {
		return svg
	}
	at := loc[0] + len("<svg")
	out := make([]byte, 0, len(svg)+len(` stroke-width="1"`))
	out = append(out, svg[:at]...)
	out = append(out, ` stroke-width="1"`...)
	return append(out, svg[at:]...)
}

// rootTag returns the span of the first <svg> tag outside a comment.
func rootTag(svg []byte) []int {
	comments := commentRe.FindAllIndex(svg, -1)
	for _, loc := range rootTagRe.FindAllIndex(svg, -1) {
		inside := false
		for _, c := range comments {
			if loc[0] >= c[0] && loc[0] < c[1] {
				inside = true
				break
			}
		}
		if !inside {
			return loc
		}
	}
	return nil
}

// scaleStrokes widens every stroke and its dash pattern by the icon's
// transform. oksvg scales path geometry but hands line widths and dashes to
// the rasterizer as written, so at scale 2 lines were drawn at half weight.
func scaleStrokes(icon *oksvg.SvgIcon) {
	k := math.Sqrt(math.Abs(icon.Transform.A * icon.Transform.D))
	for i := range icon.SVGPaths {
		p := &icon.SVGPaths[i]
		p.LineWidth *= k
		p.DashOffset *= k
		if p.Dash == nil {
			continue
		}
		// Paths can share one Dash slice, so scale a copy.
		dash := make([]float64, len(p.Dash))
		for j, v := range p.Dash {
			dash[j] = v * k
		}
		p.Dash = dash
	}
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
