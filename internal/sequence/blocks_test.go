package sequence

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func parsedSeq(src string) *Diagram {
	d, err := Parse(src)
	if err != nil {
		panic(err)
	}
	return d
}

func TestAutonumberOperands(t *testing.T) {
	Convey("Given autonumber with a start and a step", t, func() {
		d := parsedSeq("sequenceDiagram\nautonumber 10 10\nA->>B: one\nA->>B: two")

		Convey("When parsing", func() {
			Convey("Then numbering honours both operands", func() {
				So(d.Messages[0].Num, ShouldEqual, 10)
				So(d.Messages[1].Num, ShouldEqual, 20)
			})
		})
	})

	Convey("Given autonumber with no operands", t, func() {
		d := parsedSeq("sequenceDiagram\nautonumber\nA->>B: one\nA->>B: two")

		Convey("When parsing", func() {
			Convey("Then it numbers from one in steps of one", func() {
				So(d.Messages[0].Num, ShouldEqual, 1)
				So(d.Messages[1].Num, ShouldEqual, 2)
			})
		})
	})

	Convey("Given autonumber off", t, func() {
		d := parsedSeq("sequenceDiagram\nautonumber off\nA->>B: one")

		Convey("When parsing", func() {
			Convey("Then messages are not numbered", func() {
				So(d.Messages[0].Num, ShouldEqual, 0)
			})
		})
	})
}

func TestBoxDoesNotCloseFrame(t *testing.T) {
	Convey("Given a box nested inside a loop", t, func() {
		d := parsedSeq("sequenceDiagram\nloop every day\nbox Group\nparticipant A\nend\nA->>A: x\nend")

		Convey("When parsing", func() {
			Convey("Then the box's end closes the box, not the loop", func() {
				So(len(d.Frames), ShouldEqual, 1)
				So(d.Frames[0].Type, ShouldEqual, "loop")
				So(d.Frames[0].Label, ShouldEqual, "every day")
				So(d.Frames[0].EndRow, ShouldBeGreaterThanOrEqualTo, d.Frames[0].StartRow)
			})
		})
	})
}

func TestRectColour(t *testing.T) {
	Convey("Given a rect frame with a colour", t, func() {
		d := parsedSeq("sequenceDiagram\nrect rgb(0,0,255)\nA->>B: hi\nend")

		Convey("When parsing", func() {
			Convey("Then the colour is a fill, not a label", func() {
				So(d.Frames[0].Color, ShouldEqual, "rgb(0,0,255)")
				So(d.Frames[0].Label, ShouldEqual, "")
			})
		})

		Convey("When rendering", func() {
			out, err := Render("sequenceDiagram\nrect rgb(0,0,255)\nA->>B: hi\nend",
				RenderOptions{Theme: "default", FontFace: "sans-serif", FontSize: 14, Padding: 16})

			Convey("Then the colour is not drawn as text", func() {
				So(err, ShouldBeNil)
				So(string(out), ShouldNotContainSubstring, ">[rgb(0,0,255)]<")
				So(string(out), ShouldContainSubstring, "rgb(0,0,255)")
			})
		})
	})

	Convey("Given a rect frame with a plain label", t, func() {
		d := parsedSeq("sequenceDiagram\nrect highlight\nA->>B: hi\nend")

		Convey("When parsing", func() {
			Convey("Then it stays a label", func() {
				So(d.Frames[0].Label, ShouldEqual, "highlight")
				So(d.Frames[0].Color, ShouldEqual, "")
			})
		})
	})
}
