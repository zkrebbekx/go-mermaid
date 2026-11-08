package domain

// Arrow is the head/line style of an edge.
type Arrow string

const (
	// ArrowNormal is a solid line with an arrowhead (-->).
	ArrowNormal Arrow = "normal"
	// ArrowOpen is a solid line with no arrowhead (---).
	ArrowOpen Arrow = "open"
	// ArrowDotted is a dotted line with an arrowhead (-.->).
	ArrowDotted Arrow = "dotted"
	// ArrowThick is a thick line with an arrowhead (==>).
	ArrowThick Arrow = "thick"
)

// Edge connects two nodes by ID.
type Edge struct {
	From  string
	To    string
	Label string
	Arrow Arrow

	// Style holds optional per-edge overrides from a linkStyle directive.
	// nil means use the theme and the arrow kind.
	Style *Style

	// Points is the laid-out polyline from source to target, including
	// any bend points. Empty until layout runs.
	Points []Point

	// LabelPos is the baseline anchor of the edge label's last line. Layout
	// reserves room for each label and places it on the routed path, or just
	// outside the loop for a self-edge.
	LabelPos Point
}
