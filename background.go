package mermaid

import (
	"fmt"
	"strings"

	"github.com/zkrebbekx/go-mermaid/internal/svgutil"
)

// applyBackground post-processes the full-canvas background rect that every
// renderer emits: it recolors it (WithBackground) or removes it for a
// transparent background (WithTransparentBackground). With neither option set
// the SVG is returned unchanged.
func applyBackground(svg []byte, cfg config) []byte {
	if !cfg.bgTransparent && cfg.bgColor == "" {
		return svg
	}
	s := string(svg)
	marker := `width="100%" height="100%" fill="`
	m := strings.Index(s, marker)
	if m < 0 {
		return svg
	}
	start := strings.LastIndex(s[:m], "<rect")
	rel := strings.Index(s[m:], "/>")
	if start < 0 || rel < 0 {
		return svg
	}
	end := m + rel + 2

	if cfg.bgTransparent {
		return []byte(strings.TrimRight(s[:start], " ") + strings.TrimPrefix(s[end:], "\n"))
	}
	rect := fmt.Sprintf(`<rect width="100%%" height="100%%" fill="%s"/>`, svgutil.Esc(cfg.bgColor))
	return []byte(s[:start] + rect + s[end:])
}
