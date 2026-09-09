package sequence

import (
	"fmt"

	"github.com/zkrebbekx/go-mermaid/internal/svgutil"
)

// Options tunes sequence diagram spacing and metrics.
type Options struct {
	FontSize float64
	Padding  float64
}

// Layout holds computed geometry for rendering.
type Layout struct {
	Diagram      *Diagram
	Width        float64
	Height       float64
	HeaderHeight float64 // height of the participant header boxes
	LifelineTop  float64 // y where lifelines start (header bottom)
	LifelineEnd  float64 // y where lifelines stop

	// OffsetX is the horizontal shift the renderer must apply so that
	// content reaching left of the first lifeline stays on the canvas.
	// A note placed left of the first participant, and the frame boxes,
	// both start at a negative x before this shift.
	OffsetX float64
}

const (
	headerPadX   = 16.0 // horizontal padding inside a participant box
	headerHeight = 32.0 // participant box height
	colGap       = 40.0 // minimum gap between participant boxes
	msgGap       = 36.0 // vertical gap between messages
	frameInset   = 10.0 // how far a frame box sits outside the outer lifelines
	selfLabelGap = 6.0  // gap between a self-loop and its label
	topMargin    = 12.0 // gap between header and first message
	selfLoopW    = 44.0 // width of a self-message loop
)

// Compute assigns positions to participants and messages.
func Compute(d *Diagram, opts Options) *Layout {
	// Participant header widths and X centers, left to right.
	var x float64
	for _, p := range d.Participants {
		w := svgutil.TextWidth(p.Label, opts.FontSize) + headerPadX*2
		if w < 60 {
			w = 60
		}
		p.Width = w
		p.X = x + w/2
		x += w + colGap
	}

	lifelineTop := headerHeight
	firstY := lifelineTop + topMargin + msgGap/2
	for _, m := range d.Messages {
		m.Y = firstY + float64(m.Row)*msgGap
	}
	for _, n := range d.Notes {
		n.Y = firstY + float64(n.Row)*msgGap
	}

	rows := d.rows
	if rows == 0 {
		rows = 1
	}
	height := lifelineTop + topMargin + float64(rows)*msgGap + msgGap/2

	// Collect every horizontal extent, including the ones that reach left of
	// the origin. The vertical extent stays row-driven above, so only the X
	// axis of the bounds is consumed here.
	var bd svgutil.Bounds
	bd.Add(0, 0)
	if right := x - colGap; right > 0 { // drop trailing gap after the last participant
		bd.Add(right, 0)
	}
	// Message labels extend past the arrow they belong to. A self-message
	// draws its label to the right of the loop; a normal message centers it
	// between the two lifelines.
	for _, m := range d.Messages {
		lw := svgutil.TextWidth(MessageLabel(m), opts.FontSize)
		if m.From == m.To {
			if p := d.participant(m.From); p != nil {
				bd.Add(p.X+selfLoopW+selfLabelGap+lw, 0)
			}
			continue
		}
		from, to := d.participant(m.From), d.participant(m.To)
		if from == nil || to == nil {
			continue
		}
		mid := (from.X + to.X) / 2
		bd.Add(mid-lw/2, 0)
		bd.Add(mid+lw/2, 0)
	}
	// A note can sit left of the first lifeline or right of the last.
	for _, n := range d.Notes {
		nx, nw := noteBox(d, n, opts.FontSize)
		bd.Add(nx, 0)
		bd.Add(nx+nw, 0)
	}
	// Frame boxes sit outside the outer lifelines on both sides.
	if len(d.Frames) > 0 {
		lo, hi := participantSpan(d)
		bd.Add(lo-frameInset, 0)
		bd.Add(hi+frameInset, 0)
	}
	offsetX, _ := bd.Offset()
	width, _ := bd.Size()

	return &Layout{
		Diagram:      d,
		Width:        width,
		Height:       height,
		HeaderHeight: headerHeight,
		LifelineTop:  lifelineTop,
		LifelineEnd:  height,
		OffsetX:      offsetX,
	}
}

// participantSpan returns the left and right edges of the outer participant
// boxes. It returns zeroes when the diagram has no participants.
func participantSpan(d *Diagram) (lo, hi float64) {
	ps := d.Participants
	if len(ps) == 0 {
		return 0, 0
	}
	lo, hi = ps[0].X-ps[0].Width/2, ps[0].X+ps[0].Width/2
	for _, p := range ps {
		if l := p.X - p.Width/2; l < lo {
			lo = l
		}
		if r := p.X + p.Width/2; r > hi {
			hi = r
		}
	}
	return lo, hi
}

// rowY returns the vertical center of a row (matching message Y).
func rowY(lay *Layout, row int) float64 {
	return lay.LifelineTop + topMargin + msgGap/2 + float64(row)*msgGap
}

// noteWidth estimates a note box width from its text.
func noteWidth(text string, fontSize float64) float64 {
	w := svgutil.TextWidth(text, fontSize) + 20
	if w < 60 {
		w = 60
	}
	return w
}

// noteBox returns the left x and width of a note's box.
func noteBox(d *Diagram, n *Note, fontSize float64) (x, w float64) {
	w = noteWidth(n.Text, fontSize)
	switch n.Pos {
	case NoteRight:
		if p := d.participant(n.Of[0]); p != nil {
			x = p.X + 12
		}
	case NoteLeft:
		if p := d.participant(n.Of[0]); p != nil {
			x = p.X - 12 - w
		}
	default: // NoteOver
		p1 := d.participant(n.Of[0])
		p2 := d.participant(n.Of[len(n.Of)-1])
		if p1 == nil || p2 == nil {
			return x, w
		}
		lo, hi := min(p1.X, p2.X), max(p1.X, p2.X)
		if span := hi - lo + 40; span > w {
			w = span
		}
		x = (lo+hi)/2 - w/2
	}
	return x, w
}

// MessageLabel returns the text drawn for a message, including the autonumber
// prefix when numbering is on. The layout measures it and the renderer draws
// it, so both must derive it the same way.
func MessageLabel(m *Message) string {
	if m.Num > 0 {
		return fmt.Sprintf("%d. %s", m.Num, m.Text)
	}
	return m.Text
}
