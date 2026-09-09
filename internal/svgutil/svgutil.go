// Package svgutil holds small formatting helpers shared by the diagram
// renderers: deterministic number formatting and XML text escaping.
package svgutil

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
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

// courierAdvance is the advance width of every Courier glyph. Courier is a
// fixed-pitch face, so one value covers the whole set.
const courierAdvance = 600

// timesAdvance holds advance widths (per 1000 em units) for printable ASCII
// 32..126 in Times Roman, the standard serif core font.
var timesAdvance = [95]int{
	250, 333, 408, 500, 500, 833, 778, 180, 333, 333, 500, 564, 250, 333, 250, 278, // 32..47
	500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 278, 278, 564, 564, 564, 444, // 48..63
	921, 722, 667, 667, 722, 611, 556, 722, 722, 333, 389, 722, 611, 889, 722, 722, // 64..79
	556, 722, 667, 556, 611, 722, 722, 944, 722, 722, 611, 333, 278, 333, 469, 500, // 80..95
	333, 444, 500, 444, 500, 444, 333, 500, 500, 278, 278, 500, 278, 778, 500, 500, // 96..111
	500, 500, 333, 389, 278, 500, 500, 722, 500, 500, 444, 480, 200, 480, 541, // 112..126
}

// Face selects the metric table used to estimate text width. It mirrors the
// three CSS generic families, because the width table must match the family
// the SVG asks the viewer to render with. Measuring Helvetica while the
// viewer draws Courier makes every box the wrong size.
type Face int

const (
	// FaceSans measures with Helvetica metrics, which Arial also matches.
	FaceSans Face = iota
	// FaceSerif measures with Times Roman metrics.
	FaceSerif
	// FaceMono measures with Courier metrics.
	FaceMono
)

// FaceFor maps a CSS font-family value to the closest metric table. It reads
// the first family in the list and falls back to FaceSans, which is what an
// unknown family most often resolves to.
func FaceFor(family string) Face {
	first := family
	if i := strings.IndexByte(first, ','); i >= 0 {
		first = first[:i]
	}
	first = strings.ToLower(strings.TrimSpace(strings.Trim(strings.TrimSpace(first), `"'`)))
	switch first {
	case "monospace", "courier", "courier new", "consolas", "menlo", "monaco", "ui-monospace":
		return FaceMono
	case "serif", "times", "times new roman", "georgia", "garamond", "ui-serif":
		return FaceSerif
	default:
		return FaceSans
	}
}

// asciiAdvance returns the advance of a printable ASCII rune in 1/1000 em.
func (f Face) asciiAdvance(r rune) int {
	switch f {
	case FaceMono:
		return courierAdvance
	case FaceSerif:
		return timesAdvance[r-32]
	default:
		return helveticaAdvance[r-32]
	}
}

// East Asian wide ranges. A character in one of these occupies a full em,
// not the 0.6 em that a non-ASCII fallback would otherwise charge.
var wideRanges = [][2]rune{
	{0x1100, 0x115F}, {0x2E80, 0x303E}, {0x3041, 0x33FF},
	{0x3400, 0x4DBF}, {0x4E00, 0x9FFF}, {0xA000, 0xA4CF},
	{0xAC00, 0xD7A3}, {0xF900, 0xFAFF}, {0xFE30, 0xFE6F},
	{0xFF00, 0xFF60}, {0xFFE0, 0xFFE6},
	{0x20000, 0x2FFFD}, {0x30000, 0x3FFFD},
}

// Emoji ranges, which render wider than a full em in most fonts.
var emojiRanges = [][2]rune{
	{0x1F000, 0x1F0FF}, {0x1F300, 0x1F5FF}, {0x1F600, 0x1F64F},
	{0x1F680, 0x1F6FF}, {0x1F900, 0x1FAFF},
}

func inRanges(rs [][2]rune, r rune) bool {
	for _, p := range rs {
		if r >= p[0] && r <= p[1] {
			return true
		}
	}
	return false
}

// runeWidth returns the advance of one rune as a fraction of the font size.
func (f Face) runeWidth(r rune) float64 {
	if r >= 32 && r < 127 {
		return float64(f.asciiAdvance(r)) / 1000
	}
	switch {
	case unicode.Is(unicode.Mn, r), unicode.Is(unicode.Me, r), unicode.Is(unicode.Cf, r):
		// Combining marks, enclosing marks and format controls such as the
		// zero-width joiner and the variation selectors take no room of
		// their own; they modify the glyph beside them.
		return 0
	case r >= 0x1F3FB && r <= 0x1F3FF:
		// Skin tone modifiers merge into the emoji they follow.
		return 0
	case f == FaceMono:
		// A fixed-pitch face still doubles the cell for wide characters.
		if inRanges(wideRanges, r) {
			return float64(courierAdvance) / 1000 * 2
		}
		return float64(courierAdvance) / 1000
	case inRanges(wideRanges, r):
		return 1.0
	case inRanges(emojiRanges, r):
		return 1.2
	}
	return 0.6
}

// zeroWidthJoiner combines the emoji on either side of it into one glyph.
const zeroWidthJoiner = 0x200D

// Width estimates the rendered width of s at the given font size.
func (f Face) Width(s string, fontSize float64) float64 {
	var em float64
	joined := false
	for _, r := range s {
		if joined {
			// The previous rune was a zero-width joiner, so this one merges
			// into the glyph before it and adds no width of its own.
			joined = r == zeroWidthJoiner
			continue
		}
		joined = r == zeroWidthJoiner
		em += f.runeWidth(r)
	}
	return em * fontSize
}

// TextWidth estimates the rendered width of s at the given font size using
// sans-serif (Helvetica) metrics. Prefer Face.Width when the font family is
// known; this is the default for callers that have no family to hand.
func TextWidth(s string, fontSize float64) float64 {
	return FaceSans.Width(s, fontSize)
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
