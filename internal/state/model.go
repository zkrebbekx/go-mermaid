// Package state parses and renders Mermaid state diagrams (stateDiagram-v2)
// to SVG, reusing the shared layered layout engine. States are rounded boxes;
// [*] start/end pseudostates render as filled and ringed circles.
package state

const (
	startID = "__start__"
	endID   = "__end__"
)

// Kind distinguishes an ordinary state from the pseudostates that Mermaid
// writes with a stereotype, such as `state f <<fork>>`.
type Kind int

const (
	// KindNormal is an ordinary state box.
	KindNormal Kind = iota
	// KindFork is a <<fork>> bar, one incoming and several outgoing arrows.
	KindFork
	// KindJoin is a <<join>> bar, several incoming and one outgoing arrow.
	KindJoin
	// KindChoice is a <<choice>> diamond, a guarded branch.
	KindChoice
)

// State is a state node. Start/End mark the [*] pseudostates.
type State struct {
	ID    string
	Label string
	Start bool
	End   bool
	Kind  Kind

	// Parent is the ID of the composite state that encloses this one, or
	// the empty string at the top level.
	Parent string
}

// Transition is an arrow between two states.
type Transition struct {
	From  string
	To    string
	Label string
}

// Side is where a note sits relative to the state it annotates.
type Side int

const (
	// SideRight places the note to the right of its state.
	SideRight Side = iota
	// SideLeft places the note to the left of its state.
	SideLeft
)

// Note is a free-text annotation attached to a state.
type Note struct {
	Target string
	Side   Side
	Text   string
}

// Composite is a state that encloses a nested machine. Members lists the
// states declared inside it, in source order.
type Composite struct {
	ID      string
	Label   string
	Members []string
}

// Diagram is a parsed state diagram.
type Diagram struct {
	States      []*State
	Transitions []*Transition
	Notes       []*Note
	Composites  []*Composite

	// Direction is the layout direction requested by a `direction` line.
	// It is empty when the source does not ask for one.
	Direction string
}

func (d *Diagram) state(id string) *State {
	for _, s := range d.States {
		if s.ID == id {
			return s
		}
	}
	return nil
}

func (d *Diagram) ensureState(id string) *State {
	if s := d.state(id); s != nil {
		return s
	}
	s := &State{ID: id, Label: id}
	d.States = append(d.States, s)
	return s
}

// composite returns the composite with the given ID, or nil.
func (d *Diagram) composite(id string) *Composite {
	for _, c := range d.Composites {
		if c.ID == id {
			return c
		}
	}
	return nil
}

// pseudoID names the start or end pseudostate belonging to a scope. Mermaid
// scopes [*] to the enclosing composite, so a nested machine gets its own
// entry and exit rather than sharing the diagram's.
func pseudoID(parent string, end bool) string {
	base := startID
	if end {
		base = endID
	}
	if parent == "" {
		return base
	}
	return base + parent
}

// ensurePseudo returns the start or end pseudostate for a scope, creating it
// once.
func (d *Diagram) ensurePseudo(parent string, end bool) *State {
	id := pseudoID(parent, end)
	if s := d.state(id); s != nil {
		return s
	}
	s := &State{ID: id, Start: !end, End: end, Parent: parent}
	d.States = append(d.States, s)
	return s
}
