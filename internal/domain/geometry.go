package domain

import "math"

// Point is a 2D coordinate in SVG user units (pixels).
type Point struct {
	X, Y float64
}

// PolylineLength returns the total length of the polyline through pts.
func PolylineLength(pts []Point) float64 {
	total := 0.0
	for i := 1; i < len(pts); i++ {
		total += math.Hypot(pts[i].X-pts[i-1].X, pts[i].Y-pts[i-1].Y)
	}
	return total
}

// PolylinePointAt returns the point at distance dist along the polyline
// through pts, measured from pts[0]. A distance past either end clamps to
// that endpoint.
func PolylinePointAt(pts []Point, dist float64) Point {
	switch len(pts) {
	case 0:
		return Point{}
	case 1:
		return pts[0]
	}
	if dist <= 0 {
		return pts[0]
	}
	for i := 1; i < len(pts); i++ {
		seg := math.Hypot(pts[i].X-pts[i-1].X, pts[i].Y-pts[i-1].Y)
		if seg == 0 {
			continue
		}
		if dist <= seg {
			t := dist / seg
			return Point{
				X: pts[i-1].X + (pts[i].X-pts[i-1].X)*t,
				Y: pts[i-1].Y + (pts[i].Y-pts[i-1].Y)*t,
			}
		}
		dist -= seg
	}
	return pts[len(pts)-1]
}

// PolylineMidpoint returns the point at half the total length along the
// polyline through pts. Unlike averaging the endpoints, this lands on the
// actual routed path, so labels stay on bent (orthogonal) edges instead of
// floating in the gap between them.
func PolylineMidpoint(pts []Point) Point {
	return PolylinePointAt(pts, PolylineLength(pts)/2)
}

// Size is a width/height pair in SVG user units (pixels).
type Size struct {
	W, H float64
}

// Rect is an axis-aligned box defined by its top-left corner and size.
type Rect struct {
	Min  Point
	Size Size
}
