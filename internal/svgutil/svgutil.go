// Package svgutil holds small formatting helpers shared by the diagram
// renderers: deterministic number formatting and XML text escaping.
package svgutil

import (
	"fmt"
	"strconv"
	"strings"
)

// Num formats a float for SVG output deterministically: rounded to two
// decimals with trailing zeros trimmed (e.g. 12, 12.5, 12.25).
func Num(f float64) string {
	s := strconv.FormatFloat(f, 'f', 2, 64)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	if s == "" || s == "-0" {
		return "0"
	}
	return s
}

var escaper = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
	">", "&gt;",
	`"`, "&quot;",
	"'", "&#39;",
)

// Esc escapes text for safe inclusion in SVG/XML content and attributes.
func Esc(s string) string { return escaper.Replace(s) }

// helveticaAdvance holds advance widths (per 1000 em units) for printable
// ASCII 32..126, approximating Helvetica/Arial. Used for label sizing so wide
// glyphs (W, m, @) reserve more room than narrow ones (i, l, .).
var helveticaAdvance = [95]int{
	278, 278, 355, 556, 556, 889, 667, 191, 333, 333, 389, 584, 278, 333, 278, 278, // 32..47
	556, 556, 556, 556, 556, 556, 556, 556, 556, 556, 278, 278, 584, 584, 584, 556, // 48..63
	1015, 667, 667, 722, 722, 667, 611, 778, 722, 278, 500, 667, 556, 833, 722, 778, // 64..79
	667, 778, 722, 667, 611, 722, 667, 944, 667, 667, 611, 278, 278, 278, 469, 556, // 80..95
	333, 556, 556, 500, 556, 556, 278, 556, 556, 222, 222, 500, 222, 833, 556, 556, // 96..111
	556, 556, 333, 500, 278, 556, 500, 722, 500, 500, 500, 334, 260, 334, 584, // 112..126
}

// TextWidth estimates the rendered width of s at the given font size using
// per-character advance widths (falling back to 0.6em for non-ASCII runes).
func TextWidth(s string, fontSize float64) float64 {
	var em float64
	for _, r := range s {
		if r >= 32 && r <= 126 {
			em += float64(helveticaAdvance[r-32]) / 1000
		} else {
			em += 0.6
		}
	}
	return em * fontSize
}

// SplitLines splits label text on <br>, <br/>, <br /> and literal or escaped
// newlines, returning at least one line.
func SplitLines(s string) []string {
	for _, sep := range []string{"<br/>", "<br />", "<br>"} {
		s = strings.ReplaceAll(s, sep, "\n")
	}
	s = strings.ReplaceAll(s, "\\n", "\n")
	return strings.Split(s, "\n")
}

// MultilineText writes a centered <text> element whose lines are vertically
// centered around center. extraAttrs is inserted into the <text> tag (e.g.
// ` font-weight="bold"`).
func MultilineText(b *strings.Builder, lines []string, cx, center, lineH float64, fill, extraAttrs string) {
	n := len(lines)
	if n == 0 {
		return
	}
	fmt.Fprintf(b, `<text fill="%s" text-anchor="middle"%s>`, fill, extraAttrs)
	base := center - lineH*float64(n-1)/2
	for i, ln := range lines {
		fmt.Fprintf(b, `<tspan x="%s" y="%s">%s</tspan>`, Num(cx), Num(base+float64(i)*lineH), Esc(ln))
	}
	b.WriteString("</text>")
}

// TitleHeight is the vertical space reserved for a diagram title, or 0 when
// there is no title.
func TitleHeight(title string, fontSize float64) float64 {
	if title == "" {
		return 0
	}
	return fontSize*1.4 + 12
}

// Bounds accumulates a content bounding box in diagram coordinates. A
// renderer feeds it every drawn extent, then uses Offset to shift the
// drawing so nothing sits at a negative coordinate, and Size for the canvas.
// Without this a box that reaches left of the origin, such as a sequence
// note placed left of the first participant, falls outside the canvas and
// the viewer clips it.
type Bounds struct {
	MinX, MinY, MaxX, MaxY float64
	set                    bool
}

// Add grows the bounds to include the point (x, y).
func (b *Bounds) Add(x, y float64) {
	if !b.set {
		b.MinX, b.MinY, b.MaxX, b.MaxY = x, y, x, y
		b.set = true
		return
	}
	if x < b.MinX {
		b.MinX = x
	}
	if y < b.MinY {
		b.MinY = y
	}
	if x > b.MaxX {
		b.MaxX = x
	}
	if y > b.MaxY {
		b.MaxY = y
	}
}

// AddRect grows the bounds to include the box at (x, y) of size w by h.
func (b *Bounds) AddRect(x, y, w, h float64) {
	b.Add(x, y)
	b.Add(x+w, y+h)
}

// Empty reports whether nothing has been added yet.
func (b *Bounds) Empty() bool { return !b.set }

// Offset returns the shift that moves the content so its top-left corner is
// at the origin. Both values are zero or positive.
func (b *Bounds) Offset() (dx, dy float64) {
	if !b.set {
		return 0, 0
	}
	if b.MinX < 0 {
		dx = -b.MinX
	}
	if b.MinY < 0 {
		dy = -b.MinY
	}
	return dx, dy
}

// Size returns the extent of the content after Offset is applied.
func (b *Bounds) Size() (w, h float64) {
	if !b.set {
		return 0, 0
	}
	dx, dy := b.Offset()
	return b.MaxX + dx, b.MaxY + dy
}
