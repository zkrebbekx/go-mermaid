package sequence

import (
	"fmt"
	"strings"

	"github.com/zkrebbekx/go-mermaid/internal/svgutil"
	"github.com/zkrebbekx/go-mermaid/internal/theme"
)

// RenderOptions controls sequence diagram appearance.
type RenderOptions struct {
	Theme    string
	FontFace string
	FontSize float64
	Padding  float64
	Title    string
}

// Render parses, lays out, and renders sequence diagram source to SVG.
func Render(src string, o RenderOptions) ([]byte, error) {
	d, err := Parse(src)
	if err != nil {
		return nil, err
	}
	lay := Compute(d, Options{FontSize: o.FontSize, Padding: o.Padding})
	return svg(lay, o), nil
}

func svg(lay *Layout, o RenderOptions) []byte {
	pal := theme.For(o.Theme)
	pad := o.Padding
	titleH := svgutil.TitleHeight(o.Title, o.FontSize)
	contentW := lay.Width
	if tw := svgutil.TextWidth(o.Title, o.FontSize); tw > contentW {
		contentW = tw
	}
	w := contentW + pad*2
	h := lay.Height + titleH + pad*2

	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%s" height="%s" viewBox="0 0 %s %s" font-family="%s" font-size="%s">`,
		svgutil.Num(w), svgutil.Num(h), svgutil.Num(w), svgutil.Num(h), svgutil.Esc(o.FontFace), svgutil.Num(o.FontSize))
	b.WriteByte('\n')

	fmt.Fprintf(&b, `  <defs><marker id="seq-arrow" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="7" markerHeight="7" orient="auto-start-reverse"><path d="M0,0 L10,5 L0,10 z" fill="%s"/></marker></defs>`, pal.Edge)
	b.WriteByte('\n')
	fmt.Fprintf(&b, `  <rect width="100%%" height="100%%" fill="%s"/>`, pal.Background)
	b.WriteByte('\n')
	if o.Title != "" {
		fmt.Fprintf(&b, `  <text x="%s" y="%s" fill="%s" text-anchor="middle" font-weight="bold">%s</text>`,
			svgutil.Num(w/2), svgutil.Num(pad+o.FontSize), pal.Text, svgutil.Esc(o.Title))
		b.WriteByte('\n')
	}
	fmt.Fprintf(&b, `  <g transform="translate(%s,%s)">`, svgutil.Num(pad), svgutil.Num(pad+titleH))
	b.WriteByte('\n')

	for _, p := range lay.Diagram.Participants {
		writeLifeline(&b, p, lay, pal)
	}
	for _, bar := range lay.Diagram.Bars {
		writeBar(&b, bar, lay, pal)
	}
	for _, f := range lay.Diagram.Frames {
		writeFrame(&b, f, lay, pal, o)
	}
	for _, p := range lay.Diagram.Participants {
		writeHeader(&b, p, lay, pal, o)
	}
	for _, m := range lay.Diagram.Messages {
		writeMessage(&b, m, lay, pal)
	}
	for _, n := range lay.Diagram.Notes {
		writeNote(&b, n, lay, pal, o)
	}

	b.WriteString("  </g>\n</svg>\n")
	return []byte(b.String())
}

func writeLifeline(b *strings.Builder, p *Participant, lay *Layout, pal theme.Palette) {
	fmt.Fprintf(b, `    <line x1="%s" y1="%s" x2="%s" y2="%s" stroke="%s" stroke-dasharray="3,3"/>`,
		svgutil.Num(p.X), svgutil.Num(lay.LifelineTop), svgutil.Num(p.X), svgutil.Num(lay.LifelineEnd), pal.Edge)
	b.WriteByte('\n')
}

func writeHeader(b *strings.Builder, p *Participant, lay *Layout, pal theme.Palette, o RenderOptions) {
	x := p.X - p.Width/2
	fmt.Fprintf(b, `    <rect x="%s" y="0" width="%s" height="%s" rx="3" fill="%s" stroke="%s"/>`,
		svgutil.Num(x), svgutil.Num(p.Width), svgutil.Num(lay.HeaderHeight), pal.NodeFill, pal.NodeStroke)
	b.WriteByte('\n')
	fmt.Fprintf(b, `    <text x="%s" y="%s" fill="%s" text-anchor="middle">%s</text>`,
		svgutil.Num(p.X), svgutil.Num(lay.HeaderHeight/2+o.FontSize*0.35), pal.Text, svgutil.Esc(p.Label))
	b.WriteByte('\n')
}

func writeMessage(b *strings.Builder, m *Message, lay *Layout, pal theme.Palette) {
	from := lay.Diagram.participant(m.From)
	to := lay.Diagram.participant(m.To)
	if from == nil || to == nil {
		return
	}
	dash := ""
	if m.Arrow.Dashed {
		dash = ` stroke-dasharray="6,4"`
	}
	marker := ""
	if m.Arrow.Head == HeadArrow {
		marker = ` marker-end="url(#seq-arrow)"`
	}

	label := m.Text
	if m.Num > 0 {
		label = fmt.Sprintf("%d. %s", m.Num, m.Text)
	}

	if m.From == m.To {
		// Self-message: a small loop to the right of the lifeline.
		x := from.X
		y := m.Y
		path := fmt.Sprintf("M%s,%s L%s,%s L%s,%s L%s,%s",
			svgutil.Num(x), svgutil.Num(y),
			svgutil.Num(x+selfLoopW), svgutil.Num(y),
			svgutil.Num(x+selfLoopW), svgutil.Num(y+12),
			svgutil.Num(x), svgutil.Num(y+12))
		fmt.Fprintf(b, `    <path d="%s" fill="none" stroke="%s"%s%s/>`, path, pal.Edge, dash, marker)
		b.WriteByte('\n')
		writeMsgText(b, label, x+selfLoopW+6, y+6, pal, "start")
		return
	}

	fmt.Fprintf(b, `    <line x1="%s" y1="%s" x2="%s" y2="%s" stroke="%s"%s%s/>`,
		svgutil.Num(from.X), svgutil.Num(m.Y), svgutil.Num(to.X), svgutil.Num(m.Y), pal.Edge, dash, marker)
	b.WriteByte('\n')

	if m.Arrow.Head == HeadCross {
		writeCross(b, to.X, m.Y, from.X < to.X, pal)
	}
	writeMsgText(b, label, (from.X+to.X)/2, m.Y-4, pal, "middle")
}

func writeMsgText(b *strings.Builder, text string, x, y float64, pal theme.Palette, anchor string) {
	if text == "" {
		return
	}
	fmt.Fprintf(b, `    <text x="%s" y="%s" fill="%s" text-anchor="%s">%s</text>`,
		svgutil.Num(x), svgutil.Num(y), pal.Text, anchor, svgutil.Esc(text))
	b.WriteByte('\n')
}

// writeFrame draws a grouping frame (loop/alt/opt/...) with a label tab and
// any else/and section dividers.
func writeFrame(b *strings.Builder, f *Frame, lay *Layout, pal theme.Palette, o RenderOptions) {
	if f.EndRow < f.StartRow {
		return
	}
	ps := lay.Diagram.Participants
	if len(ps) == 0 {
		return
	}
	minX := ps[0].X - ps[0].Width/2
	maxX := ps[0].X + ps[0].Width/2
	for _, p := range ps {
		if l := p.X - p.Width/2; l < minX {
			minX = l
		}
		if r := p.X + p.Width/2; r > maxX {
			maxX = r
		}
	}
	x, w := minX-10, (maxX-minX)+20
	top := rowY(lay, f.StartRow) - 16
	bot := rowY(lay, f.EndRow) + 14

	fmt.Fprintf(b, `    <rect x="%s" y="%s" width="%s" height="%s" fill="none" stroke="%s" stroke-dasharray="2,2"/>`,
		svgutil.Num(x), svgutil.Num(top), svgutil.Num(w), svgutil.Num(bot-top), pal.NodeStroke)
	b.WriteByte('\n')

	tabW := svgutil.TextWidth(f.Type, o.FontSize) + 12
	fmt.Fprintf(b, `    <rect x="%s" y="%s" width="%s" height="%s" fill="%s" stroke="%s"/>`,
		svgutil.Num(x), svgutil.Num(top), svgutil.Num(tabW), svgutil.Num(o.FontSize+6), pal.NodeFill, pal.NodeStroke)
	b.WriteByte('\n')
	fmt.Fprintf(b, `    <text x="%s" y="%s" fill="%s" font-weight="bold">%s</text>`,
		svgutil.Num(x+6), svgutil.Num(top+o.FontSize), pal.Text, svgutil.Esc(f.Type))
	b.WriteByte('\n')
	if f.Label != "" {
		fmt.Fprintf(b, `    <text x="%s" y="%s" fill="%s">%s</text>`,
			svgutil.Num(x+tabW+6), svgutil.Num(top+o.FontSize), pal.Text, svgutil.Esc("["+f.Label+"]"))
		b.WriteByte('\n')
	}

	for _, s := range f.Sections {
		sy := rowY(lay, s.Row) - msgGap/2
		fmt.Fprintf(b, `    <line x1="%s" y1="%s" x2="%s" y2="%s" stroke="%s" stroke-dasharray="2,2"/>`,
			svgutil.Num(x), svgutil.Num(sy), svgutil.Num(x+w), svgutil.Num(sy), pal.NodeStroke)
		b.WriteByte('\n')
		if s.Label != "" {
			fmt.Fprintf(b, `    <text x="%s" y="%s" fill="%s">%s</text>`,
				svgutil.Num(x+6), svgutil.Num(sy+o.FontSize), pal.Text, svgutil.Esc("["+s.Label+"]"))
			b.WriteByte('\n')
		}
	}
}

// writeBar draws an activation bar on a participant's lifeline.
func writeBar(b *strings.Builder, bar *Bar, lay *Layout, pal theme.Palette) {
	p := lay.Diagram.participant(bar.Participant)
	if p == nil {
		return
	}
	y1 := rowY(lay, bar.StartRow)
	y2 := rowY(lay, bar.EndRow)
	if y2-y1 < msgGap/2 {
		y2 = y1 + msgGap/2
	}
	fmt.Fprintf(b, `    <rect x="%s" y="%s" width="8" height="%s" fill="%s" stroke="%s"/>`,
		svgutil.Num(p.X-4), svgutil.Num(y1), svgutil.Num(y2-y1), pal.NodeFill, pal.NodeStroke)
	b.WriteByte('\n')
}

// writeNote draws a note box and its text.
func writeNote(b *strings.Builder, n *Note, lay *Layout, pal theme.Palette, o RenderOptions) {
	x, w := noteBox(lay.Diagram, n, o.FontSize)
	h := o.FontSize + 12
	y := n.Y - h/2
	fmt.Fprintf(b, `    <rect x="%s" y="%s" width="%s" height="%s" fill="%s" stroke="%s"/>`,
		svgutil.Num(x), svgutil.Num(y), svgutil.Num(w), svgutil.Num(h), "#FFF5AD", pal.NodeStroke)
	b.WriteByte('\n')
	fmt.Fprintf(b, `    <text x="%s" y="%s" fill="%s" text-anchor="middle">%s</text>`,
		svgutil.Num(x+w/2), svgutil.Num(y+h/2+o.FontSize*0.35), pal.Text, svgutil.Esc(n.Text))
	b.WriteByte('\n')
}

// writeCross draws an X at the receiving end of a -x / --x message.
func writeCross(b *strings.Builder, x, y float64, pointingRight bool, pal theme.Palette) {
	const s = 5
	cx := x + s
	if pointingRight {
		cx = x - s
	}
	fmt.Fprintf(b, `    <path d="M%s,%s L%s,%s M%s,%s L%s,%s" stroke="%s"/>`,
		svgutil.Num(cx-s), svgutil.Num(y-s), svgutil.Num(cx+s), svgutil.Num(y+s),
		svgutil.Num(cx-s), svgutil.Num(y+s), svgutil.Num(cx+s), svgutil.Num(y-s), pal.Edge)
	b.WriteByte('\n')
}
