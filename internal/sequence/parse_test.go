package sequence

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestParse(t *testing.T) {
	Convey("Given a sequence diagram with aliases and messages", t, func() {
		src := "sequenceDiagram\n" +
			"  participant A as Alice\n" +
			"  participant B as Bob\n" +
			"  A->>B: Hello\n" +
			"  B-->>A: Hi"

		Convey("When parsing", func() {
			d, err := Parse(src)

			Convey("Then participants keep declared order and aliases", func() {
				So(err, ShouldBeNil)
				So(len(d.Participants), ShouldEqual, 2)
				So(d.Participants[0].ID, ShouldEqual, "A")
				So(d.Participants[0].Label, ShouldEqual, "Alice")
			})

			Convey("Then messages capture sender, receiver, and text", func() {
				So(len(d.Messages), ShouldEqual, 2)
				So(d.Messages[0].From, ShouldEqual, "A")
				So(d.Messages[0].To, ShouldEqual, "B")
				So(d.Messages[0].Text, ShouldEqual, "Hello")
			})

			Convey("Then arrow styles are distinguished", func() {
				So(d.Messages[0].Arrow.Dashed, ShouldBeFalse)
				So(d.Messages[0].Arrow.Head, ShouldEqual, HeadArrow)
				So(d.Messages[1].Arrow.Dashed, ShouldBeTrue)
			})
		})
	})

	Convey("Given messages with undeclared participants", t, func() {
		Convey("When parsing", func() {
			d, err := Parse("sequenceDiagram\nA->>B: hi\nB->>C: yo")

			Convey("Then participants are created in first-seen order", func() {
				So(err, ShouldBeNil)
				ids := []string{d.Participants[0].ID, d.Participants[1].ID, d.Participants[2].ID}
				So(ids, ShouldResemble, []string{"A", "B", "C"})
			})
		})
	})

	Convey("Given each arrow operator", t, func() {
		cases := []struct {
			src    string
			dashed bool
			head   Head
		}{
			{"sequenceDiagram\nA->B: x", false, HeadNone},
			{"sequenceDiagram\nA->>B: x", false, HeadArrow},
			{"sequenceDiagram\nA-->B: x", true, HeadNone},
			{"sequenceDiagram\nA-->>B: x", true, HeadArrow},
			{"sequenceDiagram\nA-xB: x", false, HeadCross},
			{"sequenceDiagram\nA--xB: x", true, HeadCross},
		}
		for _, c := range cases {
			c := c
			Convey("When parsing "+c.src, func() {
				d, err := Parse(c.src)

				Convey("Then the arrow style matches", func() {
					So(err, ShouldBeNil)
					So(d.Messages[0].Arrow.Dashed, ShouldEqual, c.dashed)
					So(d.Messages[0].Arrow.Head, ShouldEqual, c.head)
				})
			})
		}
	})

	Convey("Given a self-message", t, func() {
		Convey("When parsing", func() {
			d, err := Parse("sequenceDiagram\nA->>A: think")

			Convey("Then sender and receiver are the same participant", func() {
				So(err, ShouldBeNil)
				So(d.Messages[0].From, ShouldEqual, "A")
				So(d.Messages[0].To, ShouldEqual, "A")
				So(len(d.Participants), ShouldEqual, 1)
			})
		})
	})

	Convey("Given unsupported block keywords", t, func() {
		Convey("When parsing a diagram that uses loop/note", func() {
			d, err := Parse("sequenceDiagram\nA->>B: hi\nloop every minute\nNote over A: waiting\nend")

			Convey("Then those lines are skipped without error", func() {
				So(err, ShouldBeNil)
				So(len(d.Messages), ShouldEqual, 1)
			})
		})
	})

	Convey("Given source without the header", t, func() {
		Convey("When parsing", func() {
			_, err := Parse("A->>B: hi")

			Convey("Then it returns an error", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}
