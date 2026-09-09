package gantt

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestTaskTagsAndEndDates(t *testing.T) {
	Convey("Given a milestone with a zero duration", t, func() {
		d, err := Parse("gantt\ndateFormat YYYY-MM-DD\nsection s\nm : milestone, 2024-01-01, 0d")

		Convey("When parsing", func() {
			Convey("Then it is accepted instead of failing for want of a duration", func() {
				So(err, ShouldBeNil)
				So(d.Tasks[0].Milestone, ShouldBeTrue)
				So(d.Tasks[0].Days, ShouldEqual, 0)
			})
		})
	})

	Convey("Given a task written as start and end dates", t, func() {
		d, err := Parse("gantt\ndateFormat YYYY-MM-DD\nsection s\nTask : 2024-01-01, 2024-01-05")

		Convey("When parsing", func() {
			Convey("Then the length is the span between them", func() {
				So(err, ShouldBeNil)
				So(d.Tasks[0].Days, ShouldEqual, 4)
			})
		})
	})

	Convey("Given an end date before the start", t, func() {
		_, err := Parse("gantt\ndateFormat YYYY-MM-DD\nsection s\nTask : 2024-01-05, 2024-01-01")

		Convey("When parsing", func() {
			Convey("Then it reports the error rather than drawing a negative bar", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Given a status tag", t, func() {
		d, err := Parse("gantt\ndateFormat YYYY-MM-DD\nsection s\nTask : crit, 2024-01-01, 5d")

		Convey("When parsing", func() {
			Convey("Then the tag is recorded and not mistaken for the task id", func() {
				So(err, ShouldBeNil)
				So(d.Tasks[0].Status, ShouldEqual, "crit")
				So(d.Tasks[0].ID, ShouldEqual, "")
				So(d.Tasks[0].Days, ShouldEqual, 5)
			})
		})
	})

	Convey("Given a task with an explicit id and a duration", t, func() {
		d, err := Parse("gantt\ndateFormat YYYY-MM-DD\nsection s\nTask : a1, 2024-01-01, 5d")

		Convey("When parsing", func() {
			Convey("Then the id still works", func() {
				So(err, ShouldBeNil)
				So(d.Tasks[0].ID, ShouldEqual, "a1")
			})
		})
	})
}
