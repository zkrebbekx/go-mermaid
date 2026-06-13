package sequence

import (
	"fmt"
	"strings"

	"github.com/Zac300/go-mermaid/internal/svgutil"
	"github.com/Zac300/go-mermaid/internal/theme"
)

// RenderOptions controls sequence diagram appearance.
type RenderOptions struct {
	Theme    string
	FontFace string
	FontSize float64
	Padding  float64
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
	w := lay.Width + pad*2
	h := lay.Height + pad*2

	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%s" height="%s" viewBox="0 0 %s %s" font-family="%s" font-size="%s">`,
		svgutil.Num(w), svgutil.Num(h), svgutil.Num(w), svgutil.Num(h), svgutil.Esc(o.FontFace), svgutil.Num(o.FontSize))
	b.WriteByte('\n')

	fmt.Fprintf(&b, `  <defs><marker id="seq-arrow" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="7" markerHeight="7" orient="auto-start-reverse"><path d="M0,0 L10,5 L0,10 z" fill="%s"/></marker></defs>`, pal.Edge)
	b.WriteByte('\n')
	fmt.Fprintf(&b, `  <rect width="100%%" height="100%%" fill="%s"/>`, pal.Background)
	b.WriteByte('\n')
	fmt.Fprintf(&b, `  <g transform="translate(%s,%s)">`, svgutil.Num(pad), svgutil.Num(pad))
	b.WriteByte('\n')

	for _, p := range lay.Diagram.Participants {
		writeLifeline(&b, p, lay, pal)
	}
	for _, p := range lay.Diagram.Participants {
		writeHeader(&b, p, lay, pal, o)
	}
	for _, m := range lay.Diagram.Messages {
		writeMessage(&b, m, lay, pal)
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
		writeMsgText(b, m.Text, x+selfLoopW+6, y+6, pal, "start")
		return
	}

	fmt.Fprintf(b, `    <line x1="%s" y1="%s" x2="%s" y2="%s" stroke="%s"%s%s/>`,
		svgutil.Num(from.X), svgutil.Num(m.Y), svgutil.Num(to.X), svgutil.Num(m.Y), pal.Edge, dash, marker)
	b.WriteByte('\n')

	if m.Arrow.Head == HeadCross {
		writeCross(b, to.X, m.Y, from.X < to.X, pal)
	}
	writeMsgText(b, m.Text, (from.X+to.X)/2, m.Y-4, pal, "middle")
}

func writeMsgText(b *strings.Builder, text string, x, y float64, pal theme.Palette, anchor string) {
	if text == "" {
		return
	}
	fmt.Fprintf(b, `    <text x="%s" y="%s" fill="%s" text-anchor="%s">%s</text>`,
		svgutil.Num(x), svgutil.Num(y), pal.Text, anchor, svgutil.Esc(text))
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
