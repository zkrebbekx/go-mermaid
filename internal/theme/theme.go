// Package theme holds the color palettes shared by all diagram renderers.
// It is a dependency-free leaf so every render stage can use one source of
// truth for colors.
package theme

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
}

// For returns the palette for name, or the default palette if name is unknown.
func For(name string) Palette {
	if p, ok := palettes[name]; ok {
		return p
	}
	return palettes["default"]
}
