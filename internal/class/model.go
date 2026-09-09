// Package class parses and renders Mermaid class diagrams to SVG. It reuses
// the shared layered layout engine for positioning and edge routing, drawing
// UML class boxes (name / attributes / methods compartments) and relationship
// markers (inheritance, composition, aggregation, association, dependency).
package class

// Class is a UML class with attribute and method members.
type Class struct {
	Name       string
	Attributes []string
	Methods    []string

	// Display is the name as written, generic parameters included, such as
	// "Box~T~". It falls back to Name when the source used no generics.
	Display string

	// Annotation is a stereotype written as <<interface>> or <<abstract>>
	// inside the class body, without the angle brackets.
	Annotation string

	// Namespace is the name of the enclosing namespace block, or empty.
	Namespace string
}

// Label returns the name to draw for the class.
func (c *Class) Label() string {
	if c.Display != "" {
		return c.Display
	}
	return c.Name
}

// headKind is a relationship line-end decoration.
type headKind int

const (
	headNone headKind = iota
	headArrow
	headTriangle      // inheritance / realization (hollow triangle)
	headDiamondFilled // composition
	headDiamondHollow // aggregation
)

// Relation is a relationship between two classes.
type Relation struct {
	From   string
	To     string
	Label  string
	Dashed bool
	Left   headKind // decoration at the From end
	Right  headKind // decoration at the To end

	// LeftCard and RightCard are the multiplicity labels written in quotes
	// beside each end, such as "1" and "0..*".
	LeftCard  string
	RightCard string
}

// Namespace groups classes declared inside a namespace block.
type Namespace struct {
	Name    string
	Members []string
}

// Diagram is a parsed class diagram.
type Diagram struct {
	Classes    []*Class
	Relations  []*Relation
	Namespaces []*Namespace

	// Direction is the layout direction requested by a `direction` line.
	Direction string
}

// namespace returns the namespace with the given name, or nil.
func (d *Diagram) namespace(name string) *Namespace {
	for _, n := range d.Namespaces {
		if n.Name == name {
			return n
		}
	}
	return nil
}

func (d *Diagram) class(name string) *Class {
	for _, c := range d.Classes {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func (d *Diagram) ensureClass(name string) *Class {
	if c := d.class(name); c != nil {
		return c
	}
	c := &Class{Name: name}
	d.Classes = append(d.Classes, c)
	return c
}
