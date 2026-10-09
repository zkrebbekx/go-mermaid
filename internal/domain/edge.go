package domain

// Arrow is the line style of an edge. What each end draws is set by the
// edge's Head and Tail.
type Arrow string

const (
	// ArrowNormal is a solid line (-->).
	ArrowNormal Arrow = "normal"
	// ArrowOpen is a solid line written without a head (---, --x, --o).
	ArrowOpen Arrow = "open"
	// ArrowDotted is a dotted line (-.->, -.-).
	ArrowDotted Arrow = "dotted"
	// ArrowThick is a thick line (==>, ===).
	ArrowThick Arrow = "thick"
)

// Head is the marker drawn at one end of an edge.
type Head string

const (
	// HeadNone draws nothing (---).
	HeadNone Head = ""
	// HeadArrow is a triangular arrowhead (-->, <--).
	HeadArrow Head = "arrow"
	// HeadCross is a cross (--x, x--).
	HeadCross Head = "cross"
	// HeadCircle is a filled circle (--o, o--).
	HeadCircle Head = "circle"
)

// Edge connects two nodes by ID.
type Edge struct {
	From  string
	To    string
	Label string
	Arrow Arrow
	// Head is drawn where the edge meets To, Tail where it leaves From.
	Head Head
	Tail Head

	// Style holds optional per-edge overrides from a linkStyle directive.
	// nil means use the theme and the arrow kind.
	Style *Style

	// Points is the laid-out polyline from source to target, including
	// any bend points. Empty until layout runs.
	Points []Point

	// LabelPos is the anchor for the edge label on the routed path. Layout
	// sets it to the path midpoint, or to a staggered position when several
	// edges join the same node pair and their labels would otherwise overlap.
	LabelPos Point
}
