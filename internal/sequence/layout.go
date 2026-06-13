package sequence

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
}

const (
	headerPadX   = 16.0 // horizontal padding inside a participant box
	headerHeight = 32.0 // participant box height
	colGap       = 40.0 // minimum gap between participant boxes
	msgGap       = 36.0 // vertical gap between messages
	topMargin    = 12.0 // gap between header and first message
	selfLoopW    = 44.0 // width of a self-message loop
)

// Compute assigns positions to participants and messages.
func Compute(d *Diagram, opts Options) *Layout {
	charW := opts.FontSize * 0.6

	// Participant header widths and X centers, left to right.
	var x float64
	for _, p := range d.Participants {
		w := float64(len([]rune(p.Label)))*charW + headerPadX*2
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

	width := x - colGap // drop trailing gap after the last participant
	if width < 0 {
		width = 0
	}
	// A self-message loop extends to the right of the last lifeline.
	for _, m := range d.Messages {
		if m.From == m.To {
			if p := d.participant(m.From); p != nil && p.X+selfLoopW+20 > width {
				width = p.X + selfLoopW + 20
			}
		}
	}
	// Notes to the right of / over the last participant can extend the width.
	for _, n := range d.Notes {
		if r := noteRight(d, n, opts.FontSize); r > width {
			width = r
		}
	}

	return &Layout{
		Diagram:      d,
		Width:        width,
		Height:       height,
		HeaderHeight: headerHeight,
		LifelineTop:  lifelineTop,
		LifelineEnd:  height,
	}
}

// rowY returns the vertical center of a row (matching message Y).
func rowY(lay *Layout, row int) float64 {
	return lay.LifelineTop + topMargin + msgGap/2 + float64(row)*msgGap
}

// noteWidth estimates a note box width from its text.
func noteWidth(text string, fontSize float64) float64 {
	w := float64(len([]rune(text)))*fontSize*0.6 + 20
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

func noteRight(d *Diagram, n *Note, fontSize float64) float64 {
	x, w := noteBox(d, n, fontSize)
	return x + w
}
