package gantt

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestParse(t *testing.T) {
	Convey("Given a gantt with dates and an after dependency", t, func() {
		src := "gantt\ntitle P\ndateFormat YYYY-MM-DD\nsection Phase\nDesign : d1, 2024-01-01, 10d\nBuild : after d1, 2w"

		Convey("When parsing", func() {
			d, err := Parse(src)

			Convey("Then tasks, durations, and dependencies resolve", func() {
				So(err, ShouldBeNil)
				So(d.Title, ShouldEqual, "P")
				So(len(d.Tasks), ShouldEqual, 2)
				So(d.Tasks[0].Days, ShouldEqual, 10)
				So(d.Tasks[1].Days, ShouldEqual, 14) // 2w
			})

			Convey("Then 'after' chains the start to the prior task's end", func() {
				So(d.Tasks[1].Start.Equal(d.Tasks[0].End()), ShouldBeTrue)
			})

			Convey("Then bounds span all tasks", func() {
				min, max := d.Bounds()
				So(min.Format("2006-01-02"), ShouldEqual, "2024-01-01")
				So(max.After(min), ShouldBeTrue)
			})
		})
	})

	Convey("Given a task without a duration", t, func() {
		Convey("When parsing", func() {
			_, err := Parse("gantt\nsection S\nBad : 2024-01-01")

			Convey("Then it returns an error", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Given no header", t, func() {
		Convey("When parsing", func() {
			_, err := Parse("title X")

			Convey("Then it returns an error", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestRender(t *testing.T) {
	Convey("Given a gantt, when rendering", t, func() {
		out, err := Render("gantt\ntitle P\ndateFormat YYYY-MM-DD\nsection S\nA : 2024-01-01, 5d\nB : after a, 3d",
			RenderOptions{Theme: "default", FontSize: 14, Padding: 16})
		svg := string(out)

		Convey("Then it draws task bars and labels", func() {
			So(err, ShouldBeNil)
			So(svg, ShouldStartWith, "<svg")
			So(svg, ShouldContainSubstring, ">A<")
			So(svg, ShouldContainSubstring, "<rect")
			So(svg, ShouldContainSubstring, "2024-01-01")
		})
	})
}
