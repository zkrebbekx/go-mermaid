package sequence

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/zkrebbekx/go-mermaid/internal/svgutil"
)

func layoutOf(src string) *Layout {
	d, err := Parse(src)
	if err != nil {
		panic(err)
	}
	return Compute(d, Options{FontSize: 14, Padding: 16})
}

func TestCanvasCoversOverhang(t *testing.T) {
	Convey("Given a note placed left of the first participant", t, func() {
		lay := layoutOf("sequenceDiagram\nparticipant A\nparticipant B\nA->>B: hi\nNote left of A: a long note on the far left side")

		Convey("When computing the layout", func() {
			Convey("Then the drawing is shifted right so the note stays on the canvas", func() {
				So(lay.OffsetX, ShouldBeGreaterThan, 0)
				n := lay.Diagram.Notes[0]
				x, _ := noteBox(lay.Diagram, n, svgutil.FaceSans, 14)
				So(x+lay.OffsetX, ShouldBeGreaterThanOrEqualTo, 0)
			})

			Convey("Then the width covers the whole note", func() {
				n := lay.Diagram.Notes[0]
				x, w := noteBox(lay.Diagram, n, svgutil.FaceSans, 14)
				So(x+w+lay.OffsetX, ShouldBeLessThanOrEqualTo, lay.Width)
			})
		})
	})

	Convey("Given a note placed right of the last participant", t, func() {
		lay := layoutOf("sequenceDiagram\nparticipant A\nA->>A: x\nNote right of A: a long note on the right side")

		Convey("When computing the layout", func() {
			Convey("Then the width still covers it", func() {
				n := lay.Diagram.Notes[0]
				x, w := noteBox(lay.Diagram, n, svgutil.FaceSans, 14)
				So(x+w+lay.OffsetX, ShouldBeLessThanOrEqualTo, lay.Width)
			})
		})
	})

	Convey("Given a self-message with a long label", t, func() {
		lay := layoutOf("sequenceDiagram\nparticipant A\nA->>A: a really quite long self message label")

		Convey("When computing the layout", func() {
			Convey("Then the width covers the label drawn beside the loop", func() {
				m := lay.Diagram.Messages[0]
				p := lay.Diagram.participant(m.From)
				labelEnd := p.X + selfLoopW + selfLabelGap +
					svgutil.TextWidth(MessageLabel(m), 14)
				So(labelEnd+lay.OffsetX, ShouldBeLessThanOrEqualTo, lay.Width)
			})
		})
	})

	Convey("Given a diagram with a frame", t, func() {
		lay := layoutOf("sequenceDiagram\nA->>B: hi\nloop every day\nB-->>A: yo\nend")

		Convey("When computing the layout", func() {
			Convey("Then the frame box inset is inside the canvas on both sides", func() {
				lo, hi := participantSpan(lay.Diagram)
				So(lo-frameInset+lay.OffsetX, ShouldBeGreaterThanOrEqualTo, 0)
				So(hi+frameInset+lay.OffsetX, ShouldBeLessThanOrEqualTo, lay.Width)
			})
		})
	})

	Convey("Given a plain diagram with no overhang", t, func() {
		lay := layoutOf("sequenceDiagram\nA->>B: hi")

		Convey("When computing the layout", func() {
			Convey("Then no shift is applied", func() {
				So(lay.OffsetX, ShouldEqual, 0)
			})
		})
	})
}
