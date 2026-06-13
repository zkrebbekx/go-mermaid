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
	firstMsgY := lifelineTop + topMargin + msgGap/2
	for i, m := range d.Messages {
		m.Y = firstMsgY + float64(i)*msgGap
	}

	height := lifelineTop + topMargin + float64(len(d.Messages))*msgGap + msgGap/2
	if len(d.Messages) == 0 {
		height = lifelineTop + topMargin + msgGap
	}

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

	return &Layout{
		Diagram:      d,
		Width:        width,
		Height:       height,
		HeaderHeight: headerHeight,
		LifelineTop:  lifelineTop,
		LifelineEnd:  height,
	}
}
