package raster

import (
	"image"
	"image/color"
	"math"
	"regexp"
	"strconv"
)

// oksvg does not render SVG markers, so arrowheads on marker-end edges are
// missing. We synthesize them: find elements with a marker-end or
// marker-start, take the segment at that end, and draw the marker there. The
// flowchart's cross and circle markers get their own shapes; every other
// marker is drawn as a triangle.

var (
	markerElemRe = regexp.MustCompile(`(?s)<(path|line)\b([^>]*marker-(?:end|start)="url\([^"]*"[^>]*)>`)
	markerIDRe   = regexp.MustCompile(`url\(#([^)]*)\)`)
	numRe        = regexp.MustCompile(`-?[0-9]+(?:\.[0-9]+)?`)
)

// drawArrows draws the end markers of marked edges onto img.
func drawArrows(img *image.RGBA, svg string, scale float64) {
	gx, gy := 0.0, 0.0
	if loc := groupRe.FindStringSubmatch(svg); loc != nil {
		gx = numAttr(loc[1], "0")
		gy = numAttr(loc[2], "0")
	}
	for _, m := range markerElemRe.FindAllStringSubmatch(svg, -1) {
		attrs := parseAttrs(m[2])
		col := colorOf(attrs["stroke"])
		var pts []float64
		if m[1] == "line" {
			for _, k := range []string{"x1", "y1", "x2", "y2"} {
				pts = append(pts, numAttr(attrs[k], "0"))
			}
		} else {
			pts = pathNums(attrs["d"])
		}
		if len(pts) < 4 {
			continue
		}
		at := func(i int) (float64, float64) {
			return (pts[i] + gx) * scale, (pts[i+1] + gy) * scale
		}
		// The direction comes from the nearest point that differs from the
		// tip: curved paths repeat the first point as a control point.
		toward := func(tip, step int) (float64, float64) {
			tx, ty := at(tip)
			for i := tip + step; i >= 0 && i+1 < len(pts); i += step {
				if px, py := at(i); px != tx || py != ty {
					return px, py
				}
			}
			return tx, ty
		}
		n := len(pts) &^ 1
		if id := markerID(attrs["marker-end"]); id != "" {
			tx, ty := at(n - 2)
			px, py := toward(n-2, -2)
			drawMarker(img, id, tx, ty, px, py, scale, col)
		}
		if id := markerID(attrs["marker-start"]); id != "" {
			tx, ty := at(0)
			px, py := toward(0, 2)
			drawMarker(img, id, tx, ty, px, py, scale, col)
		}
	}
}

// markerID returns the id in a url(#id) reference, or "" for none.
func markerID(ref string) string {
	if m := markerIDRe.FindStringSubmatch(ref); m != nil {
		return m[1]
	}
	return ""
}

// pathNums returns the coordinates in a path's d, in order.
func pathNums(d string) []float64 {
	nums := numRe.FindAllString(d, -1)
	out := make([]float64, 0, len(nums))
	for _, s := range nums {
		v, _ := strconv.ParseFloat(s, 64)
		out = append(out, v)
	}
	return out
}

// drawMarker draws the marker named id with its tip at (tx,ty), pointing
// away from (px,py). Cross and circle match the flowchart's SVG markers: a
// 10-unit box drawn at 0.7 of the 1.5 stroke, centred 4 units back from the
// tip.
func drawMarker(img *image.RGBA, id string, tx, ty, px, py, scale float64, col color.Color) {
	dx, dy := tx-px, ty-py
	d := math.Hypot(dx, dy)
	if d == 0 {
		return
	}
	dx, dy = dx/d, dy/d
	u := 1.05 * scale
	cx, cy := tx-dx*4*u, ty-dy*4*u
	switch id {
	case "cross":
		for _, s := range []float64{1, -1} {
			ax, ay := (dx-s*dy)*4*u, (dy+s*dx)*4*u
			strokeLine(img, cx-ax, cy-ay, cx+ax, cy+ay, 2*u, col)
		}
	case "circle":
		fillCircle(img, cx, cy, 4*u, col)
	default:
		arrowhead(img, tx, ty, px, py, scale, col)
	}
}

// fillCircle fills a disc of radius r centred on (cx,cy).
func fillCircle(img *image.RGBA, cx, cy, r float64, col color.Color) {
	b := img.Bounds()
	for y := int(math.Floor(cy - r)); y <= int(math.Ceil(cy+r)); y++ {
		for x := int(math.Floor(cx - r)); x <= int(math.Ceil(cx+r)); x++ {
			if math.Hypot(float64(x)+0.5-cx, float64(y)+0.5-cy) <= r && image.Pt(x, y).In(b) {
				img.Set(x, y, col)
			}
		}
	}
}

// strokeLine draws a line of width w from (x0,y0) to (x1,y1) as two
// triangles.
func strokeLine(img *image.RGBA, x0, y0, x1, y1, w float64, col color.Color) {
	dx, dy := x1-x0, y1-y0
	d := math.Hypot(dx, dy)
	if d == 0 {
		return
	}
	nx, ny := -dy/d*w/2, dx/d*w/2
	fillTriangle(img, x0+nx, y0+ny, x1+nx, y1+ny, x1-nx, y1-ny, col)
	fillTriangle(img, x0+nx, y0+ny, x1-nx, y1-ny, x0-nx, y0-ny, col)
}

// arrowhead fills a triangle at (tx,ty) pointing away from (px,py).
func arrowhead(img *image.RGBA, tx, ty, px, py, scale float64, col color.Color) {
	dx, dy := tx-px, ty-py
	d := math.Hypot(dx, dy)
	if d == 0 {
		return
	}
	dx, dy = dx/d, dy/d
	l, half := 9*scale, 4.5*scale
	bx, by := tx-dx*l, ty-dy*l
	perpx, perpy := -dy, dx
	fillTriangle(img,
		tx, ty,
		bx+perpx*half, by+perpy*half,
		bx-perpx*half, by-perpy*half,
		col)
}

// fillTriangle rasterizes a solid triangle with a simple scanline fill.
func fillTriangle(img *image.RGBA, x0, y0, x1, y1, x2, y2 float64, col color.Color) {
	minX := int(math.Floor(min3(x0, x1, x2)))
	maxX := int(math.Ceil(max3(x0, x1, x2)))
	minY := int(math.Floor(min3(y0, y1, y2)))
	maxY := int(math.Ceil(max3(y0, y1, y2)))
	area := edge(x0, y0, x1, y1, x2, y2)
	if area == 0 {
		return
	}
	b := img.Bounds()
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			fx, fy := float64(x)+0.5, float64(y)+0.5
			w0 := edge(x1, y1, x2, y2, fx, fy)
			w1 := edge(x2, y2, x0, y0, fx, fy)
			w2 := edge(x0, y0, x1, y1, fx, fy)
			if (w0 >= 0 && w1 >= 0 && w2 >= 0) || (w0 <= 0 && w1 <= 0 && w2 <= 0) {
				if image.Pt(x, y).In(b) {
					img.Set(x, y, col)
				}
			}
		}
	}
}

func edge(ax, ay, bx, by, cx, cy float64) float64 {
	return (bx-ax)*(cy-ay) - (by-ay)*(cx-ax)
}

func min3(a, b, c float64) float64 { return math.Min(a, math.Min(b, c)) }
func max3(a, b, c float64) float64 { return math.Max(a, math.Max(b, c)) }
