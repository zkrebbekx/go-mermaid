package state

import (
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func parsed(src string) *Diagram {
	d, err := Parse(src)
	if err != nil {
		panic(err)
	}
	return d
}

func rendered(src string) string {
	out, err := Render(src, RenderOptions{Theme: "default", FontFace: "sans-serif", FontSize: 14, Padding: 16})
	if err != nil {
		panic(err)
	}
	return string(out)
}

func TestCompositeStates(t *testing.T) {
	Convey("Given a composite state with a nested machine", t, func() {
		src := "stateDiagram-v2\n[*] --> Outer\nstate Outer {\n[*] --> Loading\nLoading --> Ready\nReady --> [*]\n}\nOuter --> Done"

		Convey("When parsing", func() {
			d := parsed(src)

			Convey("Then the nested states survive instead of being skipped", func() {
				So(d.state("Loading"), ShouldNotBeNil)
				So(d.state("Ready"), ShouldNotBeNil)
			})

			Convey("Then the nested states belong to the composite", func() {
				So(d.state("Loading").Parent, ShouldEqual, "Outer")
				So(d.state("Ready").Parent, ShouldEqual, "Outer")
				So(d.state("Done").Parent, ShouldEqual, "")
			})

			Convey("Then the composite records its members, entry and exit included", func() {
				c := d.composite("Outer")
				So(c, ShouldNotBeNil)
				So(c.Members, ShouldContain, "Loading")
				So(c.Members, ShouldContain, "Ready")
				So(c.Members, ShouldContain, pseudoID("Outer", false))
				So(c.Members, ShouldContain, pseudoID("Outer", true))
			})

			Convey("Then the nested [*] is scoped to the composite, not the diagram", func() {
				So(pseudoID("Outer", false), ShouldNotEqual, startID)
				So(d.state(startID), ShouldNotBeNil)
				So(d.state(pseudoID("Outer", false)), ShouldNotBeNil)
			})

			Convey("Then every inner transition is kept", func() {
				So(len(d.Transitions), ShouldEqual, 5)
			})
		})

		Convey("When rendering", func() {
			svg := rendered(src)

			Convey("Then the nested state labels appear in the output", func() {
				So(svg, ShouldContainSubstring, ">Loading<")
				So(svg, ShouldContainSubstring, ">Ready<")
			})

			Convey("Then a dashed cluster box is drawn with the composite name", func() {
				So(svg, ShouldContainSubstring, `stroke-dasharray="4,3"`)
				So(svg, ShouldContainSubstring, ">Outer<")
			})
		})
	})

	Convey("Given a composite nested inside another composite", t, func() {
		d := parsed("stateDiagram-v2\nstate Outer {\nstate Inner {\nX --> Y\n}\n}")

		Convey("When parsing", func() {
			Convey("Then both composites are recorded and the inner one is a member", func() {
				So(d.composite("Outer"), ShouldNotBeNil)
				So(d.composite("Inner"), ShouldNotBeNil)
				So(d.composite("Outer").Members, ShouldContain, "Inner")
				So(d.state("X").Parent, ShouldEqual, "Inner")
			})
		})
	})
}

func TestStateNotes(t *testing.T) {
	Convey("Given notes on either side of a state", t, func() {
		src := "stateDiagram-v2\nA --> B\nnote right of A : explains A\nnote left of B : explains B"

		Convey("When parsing", func() {
			d := parsed(src)

			Convey("Then they become notes, not states", func() {
				So(len(d.Notes), ShouldEqual, 2)
				So(d.state("explains A"), ShouldBeNil)
				So(d.state("explains B"), ShouldBeNil)
				So(len(d.States), ShouldEqual, 2)
			})

			Convey("Then each records its side and target", func() {
				So(d.Notes[0].Target, ShouldEqual, "A")
				So(d.Notes[0].Side, ShouldEqual, SideRight)
				So(d.Notes[1].Target, ShouldEqual, "B")
				So(d.Notes[1].Side, ShouldEqual, SideLeft)
			})
		})

		Convey("When rendering", func() {
			svg := rendered(src)

			Convey("Then the note text is drawn", func() {
				So(svg, ShouldContainSubstring, ">explains A<")
				So(svg, ShouldContainSubstring, ">explains B<")
			})
		})
	})

	Convey("Given a note that spans several lines", t, func() {
		d := parsed("stateDiagram-v2\nA --> B\nnote right of A\n  first line\n  second line\nend note")

		Convey("When parsing", func() {
			Convey("Then the body is joined and no state is invented", func() {
				So(len(d.Notes), ShouldEqual, 1)
				So(d.Notes[0].Text, ShouldEqual, "first line second line")
				So(len(d.States), ShouldEqual, 2)
			})
		})
	})
}

func TestStatePseudostates(t *testing.T) {
	Convey("Given fork, join and choice pseudostates", t, func() {
		d := parsed("stateDiagram-v2\nstate f <<fork>>\nstate j <<join>>\nstate c <<choice>>\nA --> f")

		Convey("When parsing", func() {
			Convey("Then each carries its kind and drops the stereotype text", func() {
				So(d.state("f").Kind, ShouldEqual, KindFork)
				So(d.state("j").Kind, ShouldEqual, KindJoin)
				So(d.state("c").Kind, ShouldEqual, KindChoice)
				So(d.state("f <<fork>>"), ShouldBeNil)
			})
		})

		Convey("When rendering a fork", func() {
			svg := rendered("stateDiagram-v2\nstate f <<fork>>\n[*] --> f\nf --> A\nf --> B")

			Convey("Then no stereotype text is drawn", func() {
				So(svg, ShouldNotContainSubstring, "fork")
			})
		})

		Convey("When rendering a choice", func() {
			svg := rendered("stateDiagram-v2\nstate c <<choice>>\nA --> c\nc --> B")

			Convey("Then it is drawn as a diamond", func() {
				So(svg, ShouldContainSubstring, "<polygon")
			})
		})
	})
}

func TestStateDirectionAndAlias(t *testing.T) {
	Convey("Given a direction line", t, func() {
		d := parsed("stateDiagram-v2\ndirection LR\nA --> B")

		Convey("When parsing", func() {
			Convey("Then it sets the direction and invents no state", func() {
				So(d.Direction, ShouldEqual, "LR")
				So(d.state("direction LR"), ShouldBeNil)
				So(len(d.States), ShouldEqual, 2)
			})
		})

		Convey("When rendering left to right", func() {
			svg := rendered("stateDiagram-v2\ndirection LR\nA --> B")

			Convey("Then B is placed to the right of A, not below it", func() {
				So(svg, ShouldNotContainSubstring, ">direction<")
			})
		})
	})

	Convey("Given an aliased state declaration", t, func() {
		d := parsed("stateDiagram-v2\nstate \"Long Desc\" as S\nS --> B")

		Convey("When parsing", func() {
			Convey("Then the alias is the ID and the quoted text is the label", func() {
				So(d.state("S"), ShouldNotBeNil)
				So(d.state("S").Label, ShouldEqual, "Long Desc")
				So(d.state(`"Long Desc"`), ShouldBeNil)
			})

			Convey("Then the later reference joins the same state", func() {
				So(len(d.States), ShouldEqual, 2)
			})
		})
	})

	Convey("Given a concurrency separator inside a composite", t, func() {
		d := parsed("stateDiagram-v2\nstate Both {\nA --> B\n--\nC --> D\n}")

		Convey("When parsing", func() {
			Convey("Then no state is invented for it", func() {
				So(d.state("--"), ShouldBeNil)
				So(len(d.States), ShouldEqual, 5) // A B C D and the composite
			})
		})
	})
}

func TestStateRegression(t *testing.T) {
	Convey("Given an ordinary state machine", t, func() {
		src := "stateDiagram-v2\n[*] --> Still\nStill --> Moving : start\nMoving --> Still : stop\nMoving --> Crash\nCrash --> [*]"

		Convey("When rendering", func() {
			svg := rendered(src)

			Convey("Then every state and label still appears", func() {
				for _, want := range []string{">Still<", ">Moving<", ">Crash<", ">start<", ">stop<"} {
					So(svg, ShouldContainSubstring, want)
				}
			})

			Convey("Then no cluster box is drawn", func() {
				So(strings.Contains(svg, `stroke-dasharray="4,3"`), ShouldBeFalse)
			})
		})
	})
}
