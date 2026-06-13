// Package sequence parses, lays out, and renders Mermaid sequence diagrams
// to SVG. It is self-contained: source text in, SVG bytes out.
package sequence

// Head is the arrowhead style at the end of a message.
type Head int

const (
	// HeadNone is a line with no arrowhead (-> / -->).
	HeadNone Head = iota
	// HeadArrow is a solid triangular arrowhead (->> / -->>).
	HeadArrow
	// HeadCross is an X at the end, denoting a lost/async-end message (-x / --x).
	HeadCross
)

// Arrow describes a message line's style.
type Arrow struct {
	Dashed bool
	Head   Head
}

// Participant is an actor with a vertical lifeline.
type Participant struct {
	ID    string
	Label string

	// X is the lifeline's horizontal center; set during layout.
	X float64
	// Width is the header box width; set during layout.
	Width float64
}

// Message is a single arrow from one participant to another.
type Message struct {
	From  string
	To    string
	Text  string
	Arrow Arrow

	// Y is the vertical position of the message line; set during layout.
	Y float64
}

// Diagram is a parsed sequence diagram.
type Diagram struct {
	Participants []*Participant
	Messages     []*Message
}

// participant returns the participant with id, or nil.
func (d *Diagram) participant(id string) *Participant {
	for _, p := range d.Participants {
		if p.ID == id {
			return p
		}
	}
	return nil
}
