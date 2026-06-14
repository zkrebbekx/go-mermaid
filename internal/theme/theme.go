// Package theme holds the color palettes shared by all diagram renderers.
// It is a dependency-free leaf so every render stage can use one source of
// truth for colors.
package theme

import "sync"

// Palette holds the colors used to render a diagram.
type Palette struct {
	Background string
	NodeFill   string
	NodeStroke string
	Text       string
	Edge       string
}

// palettes maps theme names to palettes. Unknown names fall back to default.
var palettes = map[string]Palette{
	"default": {
		Background: "#ffffff",
		NodeFill:   "#ECECFF",
		NodeStroke: "#9370DB",
		Text:       "#333333",
		Edge:       "#333333",
	},
	"dark": {
		Background: "#1e1e1e",
		NodeFill:   "#2b2b40",
		NodeStroke: "#8888bb",
		Text:       "#e6e6e6",
		Edge:       "#bbbbbb",
	},
	"neutral": {
		Background: "#ffffff",
		NodeFill:   "#eeeeee",
		NodeStroke: "#999999",
		Text:       "#222222",
		Edge:       "#555555",
	},
	"forest": {
		Background: "#ffffff",
		NodeFill:   "#cde498",
		NodeStroke: "#13540c",
		Text:       "#13540c",
		Edge:       "#3a7a2a",
	},
	"base": {
		Background: "#ffffff",
		NodeFill:   "#e8e8e8",
		NodeStroke: "#666666",
		Text:       "#1a1a1a",
		Edge:       "#444444",
	},
}

// Names returns the available theme names in a stable order.
func Names() []string {
	return []string{"default", "dark", "neutral", "forest", "base"}
}

var (
	mu     sync.RWMutex
	custom = map[string]Palette{}
)

// Register adds or replaces a custom palette under name.
func Register(name string, p Palette) {
	mu.Lock()
	custom[name] = p
	mu.Unlock()
}

// For returns the palette for name: a registered custom palette, then a
// built-in, otherwise the default palette.
func For(name string) Palette {
	mu.RLock()
	p, ok := custom[name]
	mu.RUnlock()
	if ok {
		return p
	}
	if p, ok := palettes[name]; ok {
		return p
	}
	return palettes["default"]
}
