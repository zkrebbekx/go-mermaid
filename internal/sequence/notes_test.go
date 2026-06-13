package sequence

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNotesAndActivations(t *testing.T) {
	Convey("Given messages with notes interleaved", t, func() {
		src := "sequenceDiagram\nA->>B: hi\nNote right of B: thinking\nB-->>A: bye"

		Convey("When parsing", func() {
			d, err := Parse(src)

			Convey("Then notes and messages share an ordered row space", func() {
				So(err, ShouldBeNil)
				So(len(d.Messages), ShouldEqual, 2)
				So(len(d.Notes), ShouldEqual, 1)
				So(d.Notes[0].Pos, ShouldEqual, NoteRight)
				So(d.Notes[0].Text, ShouldEqual, "thinking")
				So(d.Messages[0].Row, ShouldEqual, 0)
				So(d.Notes[0].Row, ShouldEqual, 1)
				So(d.Messages[1].Row, ShouldEqual, 2)
			})
		})
	})

	Convey("Given a note over two participants", t, func() {
		Convey("When parsing", func() {
			d, err := Parse("sequenceDiagram\nA->>B: x\nNote over A,B: shared")

			Convey("Then the note spans both", func() {
				So(err, ShouldBeNil)
				So(d.Notes[0].Pos, ShouldEqual, NoteOver)
				So(d.Notes[0].Of, ShouldResemble, []string{"A", "B"})
			})
		})
	})

	Convey("Given activation via +/- suffixes", t, func() {
		Convey("When parsing", func() {
			d, err := Parse("sequenceDiagram\nA->>+B: call\nB-->>-A: return")

			Convey("Then a bar spans from activation to deactivation", func() {
				So(err, ShouldBeNil)
				So(len(d.Bars), ShouldEqual, 1)
				So(d.Bars[0].Participant, ShouldEqual, "B")
				So(d.Bars[0].StartRow, ShouldEqual, 0)
				So(d.Bars[0].EndRow, ShouldEqual, 1)
			})
		})
	})

	Convey("Given explicit activate/deactivate", t, func() {
		Convey("When parsing", func() {
			d, err := Parse("sequenceDiagram\nactivate B\nA->>B: x\ndeactivate B")

			Convey("Then a bar is produced", func() {
				So(err, ShouldBeNil)
				So(len(d.Bars), ShouldEqual, 1)
			})
		})
	})

	Convey("Given an unclosed activation", t, func() {
		Convey("When parsing", func() {
			d, err := Parse("sequenceDiagram\nA->>+B: x\nB->>C: y")

			Convey("Then it is closed at the last row", func() {
				So(err, ShouldBeNil)
				So(len(d.Bars), ShouldEqual, 1)
				So(d.Bars[0].EndRow, ShouldEqual, 1)
			})
		})
	})

	Convey("Given notes and activations", t, func() {
		Convey("When rendering", func() {
			out, err := Render("sequenceDiagram\nA->>+B: hi\nNote over B: working\nB-->>-A: bye",
				RenderOptions{Theme: "default", FontSize: 14, Padding: 16})
			svg := string(out)

			Convey("Then a note box and an activation bar are drawn", func() {
				So(err, ShouldBeNil)
				So(svg, ShouldContainSubstring, "#FFF5AD") // note fill
				So(svg, ShouldContainSubstring, ">working<")
			})
		})
	})
}
