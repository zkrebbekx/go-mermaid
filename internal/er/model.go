// Package er parses and renders Mermaid entity-relationship diagrams
// (erDiagram) to SVG, reusing the shared layered layout engine. Entities are
// boxes with an attribute compartment; relationships carry cardinality labels.
package er

// Entity is a table-like box with attribute rows ("type name").
type Entity struct {
	Name       string
	Attributes []string
}

// Card is a crow's-foot cardinality kind.
type Card int

const (
	// CardOne is exactly one (||).
	CardOne Card = iota
	// CardZeroOne is zero or one (|o / o|).
	CardZeroOne
	// CardMany is many ({ / }).
	CardMany
	// CardOneMany is one or more (}| / |{).
	CardOneMany
	// CardZeroMany is zero or more (}o / o{).
	CardZeroMany
)

// Relationship connects two entities with cardinality at each end.
type Relationship struct {
	From      string
	To        string
	Label     string
	LeftCard  string // human-readable cardinality at the From end
	RightCard string // ... at the To end
	LeftKind  Card   // crow's-foot kind at the From end
	RightKind Card   // ... at the To end
	Dashed    bool   // non-identifying relationship
}

// Diagram is a parsed ER diagram.
type Diagram struct {
	Entities      []*Entity
	Relationships []*Relationship
}

func (d *Diagram) entity(name string) *Entity {
	for _, e := range d.Entities {
		if e.Name == name {
			return e
		}
	}
	return nil
}

func (d *Diagram) ensureEntity(name string) *Entity {
	if e := d.entity(name); e != nil {
		return e
	}
	e := &Entity{Name: name}
	d.Entities = append(d.Entities, e)
	return e
}
