package raster

import (
	"image"
	"image/color"
	"regexp"
	"strconv"
	"strings"

	xhtml "html"

	"github.com/Zac300/go-mermaid/internal/svgutil"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

// oksvg does not render <text>, so we draw labels ourselves. We control the
// SVG format, so a light regex pass over <text>/<tspan> elements is enough.

var (
	textRe     = regexp.MustCompile(`(?s)<text\b([^>]*)>(.*?)</text>`)
	tspanRe    = regexp.MustCompile(`(?s)<tspan\b([^>]*)>(.*?)</tspan>`)
	attrRe     = regexp.MustCompile(`([\w-]+)="([^"]*)"`)
	groupRe    = regexp.MustCompile(`<g transform="translate\(([0-9.-]+),([0-9.-]+)\)">`)
	regularFnt *opentype.Font
	boldFnt    *opentype.Font
)

func init() {
	regularFnt, _ = opentype.Parse(goregular.TTF)
	boldFnt, _ = opentype.Parse(gobold.TTF)
}

// drawText overlays the SVG's text elements onto img at the given scale.
// Text inside the content group inherits that group's translate; text before
// it (e.g. the title) does not.
func drawText(img *image.RGBA, svg string, scale, rootSize float64) {
	gx, gy, gIdx := 0.0, 0.0, -1
	if loc := groupRe.FindStringSubmatchIndex(svg); loc != nil {
		gx = numAttr(svg[loc[2]:loc[3]], "0")
		gy = numAttr(svg[loc[4]:loc[5]], "0")
		gIdx = loc[0]
	}
	for _, loc := range textRe.FindAllStringSubmatchIndex(svg, -1) {
		ox, oy := 0.0, 0.0
		if gIdx >= 0 && loc[0] > gIdx {
			ox, oy = gx, gy
		}
		attrs := parseAttrs(svg[loc[2]:loc[3]])
		inner := svg[loc[4]:loc[5]]
		fill := colorOf(attrs["fill"])
		size := sizeOf(attrs["font-size"], rootSize)
		bold := attrs["font-weight"] == "bold"
		anchor := attrs["text-anchor"]

		if strings.Contains(inner, "<tspan") {
			for _, ts := range tspanRe.FindAllStringSubmatch(inner, -1) {
				ta := parseAttrs(ts[1])
				x := numAttr(ta["x"], attrs["x"])
				y := numAttr(ta["y"], attrs["y"])
				drawString(img, xhtml.UnescapeString(ts[2]), x+ox, y+oy, size, scale, fill, bold, anchor)
			}
			continue
		}
		x := numAttr(attrs["x"], "0")
		y := numAttr(attrs["y"], "0")
		drawString(img, xhtml.UnescapeString(inner), x+ox, y+oy, size, scale, fill, bold, anchor)
	}
}

func drawString(img *image.RGBA, s string, x, y, size, scale float64, fill color.Color, bold bool, anchor string) {
	if strings.TrimSpace(s) == "" {
		return
	}
	fnt := regularFnt
	if bold {
		fnt = boldFnt
	}
	if fnt == nil {
		return
	}
	px := size * scale
	face, err := opentype.NewFace(fnt, &opentype.FaceOptions{Size: px, DPI: 72})
	if err != nil {
		return
	}
	adv := float64(font.MeasureString(face, s)) / 64

	// The layout reserved space using svgutil.TextWidth (sans-serif metrics).
	// The Go font is wider, so shrink to fit the reserved width and keep PNG
	// labels inside their boxes.
	if reserved := svgutil.TextWidth(s, size) * scale; reserved > 0 && adv > reserved {
		_ = face.Close()
		px *= reserved / adv
		face, err = opentype.NewFace(fnt, &opentype.FaceOptions{Size: px, DPI: 72})
		if err != nil {
			return
		}
		adv = float64(font.MeasureString(face, s)) / 64
	}
	defer func() { _ = face.Close() }()

	d := &font.Drawer{Dst: img, Src: image.NewUniform(fill), Face: face}
	dx := x * scale
	switch anchor {
	case "middle":
		dx -= adv / 2
	case "end":
		dx -= adv
	}
	d.Dot = fixed.Point26_6{X: fixed.Int26_6(dx * 64), Y: fixed.Int26_6(y * scale * 64)}
	d.DrawString(s)
}

func parseAttrs(s string) map[string]string {
	m := map[string]string{}
	for _, a := range attrRe.FindAllStringSubmatch(s, -1) {
		m[a[1]] = a[2]
	}
	return m
}

func numAttr(v, fallback string) float64 {
	if v == "" {
		v = fallback
	}
	f, _ := strconv.ParseFloat(v, 64)
	return f
}

func sizeOf(v string, root float64) float64 {
	if v == "" {
		if root > 0 {
			return root
		}
		return 14
	}
	f, _ := strconv.ParseFloat(v, 64)
	if f <= 0 {
		return 14
	}
	return f
}

// colorOf parses #rgb/#rrggbb (defaulting to black). "none" stays transparent.
func colorOf(s string) color.Color {
	s = strings.TrimSpace(s)
	if s == "" {
		return color.Black
	}
	if !strings.HasPrefix(s, "#") {
		return color.Black
	}
	hex := s[1:]
	if len(hex) == 3 {
		hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
	}
	if len(hex) != 6 {
		return color.Black
	}
	v, err := strconv.ParseUint(hex, 16, 32)
	if err != nil {
		return color.Black
	}
	return color.RGBA{R: byte(v >> 16), G: byte(v >> 8), B: byte(v), A: 255}
}
