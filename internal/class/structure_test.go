package class

import (
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

func TestCardinality(t *testing.T) {
	Convey("Given a relationship with multiplicity on both ends", t, func() {
		src := `classDiagram
Customer "1" o-- "0..*" Order : places`

		Convey("When parsing", func() {
			d := parsed(src)

			Convey("Then the class names are clean, not merged with the quotes", func() {
				So(d.class("Customer"), ShouldNotBeNil)
				So(d.class("Order"), ShouldNotBeNil)
				So(d.class(`Customer "1"`), ShouldBeNil)
				So(d.class(`"0..*" Order`), ShouldBeNil)
				So(len(d.Classes), ShouldEqual, 2)
			})

			Convey("Then each multiplicity is kept on the relation", func() {
				So(d.Relations[0].LeftCard, ShouldEqual, "1")
				So(d.Relations[0].RightCard, ShouldEqual, "0..*")
				So(d.Relations[0].Label, ShouldEqual, "places")
			})
		})

		Convey("When rendering", func() {
			svg := rendered(src)

			Convey("Then both multiplicities are drawn", func() {
				So(svg, ShouldContainSubstring, ">1<")
				So(svg, ShouldContainSubstring, ">0..*<")
			})
		})
	})

	Convey("Given a relationship with no multiplicity", t, func() {
		d := parsed("classDiagram\nAnimal <|-- Dog")

		Convey("When parsing", func() {
			Convey("Then both ends are empty", func() {
				So(d.Relations[0].LeftCard, ShouldEqual, "")
				So(d.Relations[0].RightCard, ShouldEqual, "")
				So(d.class("Animal"), ShouldNotBeNil)
				So(d.class("Dog"), ShouldNotBeNil)
			})
		})
	})
}

func TestGenericsAndAnnotations(t *testing.T) {
	Convey("Given a generic class", t, func() {
		d := parsed("classDiagram\nclass Box~T~ {\n+items List~T~\n}")

		Convey("When parsing", func() {
			Convey("Then the bare name identifies it and the generic form is kept for display", func() {
				So(d.class("Box"), ShouldNotBeNil)
				So(d.class("Box").Display, ShouldEqual, "Box~T~")
				So(d.class("Box").Label(), ShouldEqual, "Box~T~")
			})
		})

		Convey("When rendering", func() {
			Convey("Then the type parameter is visible", func() {
				So(rendered("classDiagram\nclass Box~T~ {\n+items List~T~\n}"), ShouldContainSubstring, "Box~T~")
			})
		})
	})

	Convey("Given a class with a stereotype annotation", t, func() {
		src := "classDiagram\nclass A {\n<<interface>>\n+run()\n}"
		d := parsed(src)

		Convey("When parsing", func() {
			Convey("Then it is an annotation, not an ordinary attribute", func() {
				So(d.class("A").Annotation, ShouldEqual, "interface")
				So(d.class("A").Attributes, ShouldBeEmpty)
				So(d.class("A").Methods, ShouldResemble, []string{"+run()"})
			})
		})

		Convey("When rendering", func() {
			Convey("Then it is drawn in guillemets above the name", func() {
				So(rendered(src), ShouldContainSubstring, "«interface»")
			})
		})
	})
}

func TestNamespacesAndDirection(t *testing.T) {
	Convey("Given classes inside a namespace", t, func() {
		src := "classDiagram\nnamespace app {\nclass A\nclass B\n}\nA --> B"

		Convey("When parsing", func() {
			d := parsed(src)

			Convey("Then it no longer fails as an unrecognized statement", func() {
				So(len(d.Classes), ShouldEqual, 2)
			})

			Convey("Then the namespace records its members", func() {
				ns := d.namespace("app")
				So(ns, ShouldNotBeNil)
				So(ns.Members, ShouldResemble, []string{"A", "B"})
				So(d.class("A").Namespace, ShouldEqual, "app")
			})
		})

		Convey("When rendering", func() {
			svg := rendered(src)

			Convey("Then a dashed box is drawn with the namespace name", func() {
				So(svg, ShouldContainSubstring, `stroke-dasharray="4,3"`)
				So(svg, ShouldContainSubstring, ">app<")
			})
		})
	})

	Convey("Given a direction line", t, func() {
		d := parsed("classDiagram\ndirection LR\nA --> B")

		Convey("When parsing", func() {
			Convey("Then it sets the direction and creates no class", func() {
				So(d.Direction, ShouldEqual, "LR")
				So(len(d.Classes), ShouldEqual, 2)
			})
		})
	})
}
