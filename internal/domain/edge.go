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

	// Points is the laid-out polyline from source to target, including
	// any bend points. Empty until layout runs.
	Points []Point

	// LabelPos is the anchor for the edge label on the routed path. Layout
	// sets it to the path midpoint, or to a staggered position when several
	// edges join the same node pair and their labels would otherwise overlap.
	LabelPos Point
}
