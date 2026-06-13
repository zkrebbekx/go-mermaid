package mermaid

import (
	"github.com/Zac300/go-mermaid/internal/layout"
	"github.com/Zac300/go-mermaid/internal/render"
	"github.com/Zac300/go-mermaid/internal/theme"
)

// Theme selects a built-in color palette for rendering.
type Theme string

const (
	// Default is the light theme used when no theme is set.
	Default Theme = "default"
	// Dark is a dark-background palette.
	Dark Theme = "dark"
	// Neutral is a grayscale palette suitable for print.
	Neutral Theme = "neutral"
	// Forest is a green palette.
	Forest Theme = "forest"
	// Base is a muted neutral palette.
	Base Theme = "base"
)

// Themes returns the names of all built-in themes.
func Themes() []string { return theme.Names() }

// DiagramTypes returns the diagram header keywords this library can render.
func DiagramTypes() []string {
	return []string{
		"graph / flowchart",
		"sequenceDiagram",
		"classDiagram",
		"stateDiagram-v2",
		"erDiagram",
		"pie",
		"journey",
		"quadrantChart",
		"gitGraph",
		"timeline",
		"mindmap",
		"gantt",
		"C4Context / C4Container",
	}
}

// Option configures a Render call. Options are applied in order; later
// options override earlier ones.
type Option func(*config)

// config holds resolved rendering settings. It is unexported; callers
// mutate it only through Option values.
type config struct {
	theme    Theme
	fontFace string
	fontSize float64
	padding  float64
	nodeSep  float64 // horizontal gap between nodes in a rank
	rankSep  float64 // gap between ranks
	curved   bool    // render flowchart edges as smooth curves
}

func defaultConfig() config {
	return config{
		theme:    Default,
		fontFace: "sans-serif",
		fontSize: 14,
		padding:  16,
		nodeSep:  50,
		rankSep:  50,
	}
}

func (c config) layout() layout.Options {
	return layout.Options{
		NodeSep:  c.nodeSep,
		RankSep:  c.rankSep,
		FontSize: c.fontSize,
	}
}

func (c config) render() render.Options {
	return render.Options{
		Theme:    string(c.theme),
		FontFace: c.fontFace,
		FontSize: c.fontSize,
		Padding:  c.padding,
		Curved:   c.curved,
	}
}

// WithTheme sets the color palette.
func WithTheme(t Theme) Option {
	return func(c *config) { c.theme = t }
}

// WithFont sets the font family and base size (in pixels) for labels.
func WithFont(face string, size float64) Option {
	return func(c *config) {
		c.fontFace = face
		c.fontSize = size
	}
}

// WithPadding sets the outer padding (in pixels) around the diagram.
func WithPadding(px float64) Option {
	return func(c *config) { c.padding = px }
}

// WithSpacing sets the gap between sibling nodes (nodeSep) and between
// ranks (rankSep), in pixels.
func WithSpacing(nodeSep, rankSep float64) Option {
	return func(c *config) {
		c.nodeSep = nodeSep
		c.rankSep = rankSep
	}
}

// WithCurvedEdges renders flowchart edges as smooth curves instead of
// straight orthogonal segments.
func WithCurvedEdges(on bool) Option {
	return func(c *config) { c.curved = on }
}
