package sequence

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFrames(t *testing.T) {
	Convey("Given a loop frame", t, func() {
		src := "sequenceDiagram\nA->>B: start\nloop every minute\nB->>A: ping\nend"

		Convey("When parsing", func() {
			d, err := Parse(src)

			Convey("Then the frame spans the contained rows", func() {
				So(err, ShouldBeNil)
				So(len(d.Frames), ShouldEqual, 1)
				So(d.Frames[0].Type, ShouldEqual, "loop")
				So(d.Frames[0].Label, ShouldEqual, "every minute")
				So(d.Frames[0].StartRow, ShouldEqual, 1)
				So(d.Frames[0].EndRow, ShouldEqual, 1)
			})
		})
	})

	Convey("Given an alt/else frame", t, func() {
		src := "sequenceDiagram\nalt ok\nA->>B: yes\nelse not ok\nA->>B: no\nend"

		Convey("When parsing", func() {
			d, err := Parse(src)

			Convey("Then the frame records the else section", func() {
				So(err, ShouldBeNil)
				So(d.Frames[0].Type, ShouldEqual, "alt")
				So(len(d.Frames[0].Sections), ShouldEqual, 1)
				So(d.Frames[0].Sections[0].Label, ShouldEqual, "not ok")
			})
		})
	})

	Convey("Given nested frames", t, func() {
		src := "sequenceDiagram\nloop outer\nA->>B: x\nopt maybe\nB->>A: y\nend\nend"

		Convey("When parsing", func() {
			d, err := Parse(src)

			Convey("Then both frames are recorded", func() {
				So(err, ShouldBeNil)
				So(len(d.Frames), ShouldEqual, 2)
			})
		})
	})

	Convey("Given autonumber", t, func() {
		Convey("When parsing", func() {
			d, err := Parse("sequenceDiagram\nautonumber\nA->>B: x\nB->>A: y")

			Convey("Then messages are numbered in order", func() {
				So(err, ShouldBeNil)
				So(d.Messages[0].Num, ShouldEqual, 1)
				So(d.Messages[1].Num, ShouldEqual, 2)
			})
		})

		Convey("When rendering", func() {
			out, _ := Render("sequenceDiagram\nautonumber\nA->>B: hello",
				RenderOptions{Theme: "default", FontSize: 14, Padding: 16})

			Convey("Then the number prefixes the label", func() {
				So(string(out), ShouldContainSubstring, ">1. hello<")
			})
		})
	})

	Convey("Given a loop, when rendering", t, func() {
		Convey("Then a frame box and label tab appear", func() {
			out, err := Render("sequenceDiagram\nA->>B: x\nloop retry\nB->>A: y\nend",
				RenderOptions{Theme: "default", FontSize: 14, Padding: 16})
			So(err, ShouldBeNil)
			So(string(out), ShouldContainSubstring, ">loop<")
			So(string(out), ShouldContainSubstring, "[retry]")
		})
	})
}
