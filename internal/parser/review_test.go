package parser

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestClickURLSanitization(t *testing.T) {
	Convey("Given click directives with various URLs", t, func() {
		Convey("When the target uses the javascript: scheme", func() {
			_, _, links, _ := Preprocess("graph TD\nA\nclick A \"javascript:alert(1)\"")
			Convey("Then the link is rejected", func() {
				So(links["A"], ShouldEqual, "")
			})
		})

		Convey("When the target is an https URL", func() {
			_, _, links, _ := Preprocess("graph TD\nA\nclick A \"https://example.com\"")
			Convey("Then the link is kept", func() {
				So(links["A"], ShouldEqual, "https://example.com")
			})
		})

		Convey("When the target is a relative path", func() {
			_, _, links, _ := Preprocess("graph TD\nA\nclick A \"/docs/x\"")
			Convey("Then the link is kept", func() {
				So(links["A"], ShouldEqual, "/docs/x")
			})
		})
	})
}

func TestInlineClassInsideLabel(t *testing.T) {
	Convey("Given a node whose label text contains ':::'", t, func() {
		src, _, _, _ := Preprocess("graph LR\nA[\"a:::b\"]")

		Convey("When preprocessing styling directives", func() {
			Convey("Then the label is left intact and no class is stripped", func() {
				So(src, ShouldContainSubstring, "a:::b")
			})
		})
	})
}

func TestSubgraphPredeclaredMembership(t *testing.T) {
	Convey("Given a node declared before a subgraph then used inside it", t, func() {
		g, err := parse("graph TD\nA --> B\nsubgraph S\nA --> C\nend")

		Convey("When parsing", func() {
			Convey("Then the pre-declared node belongs to the subgraph", func() {
				So(err, ShouldBeNil)
				So(len(g.Subgraphs), ShouldEqual, 1)
				So(g.Subgraphs[0].NodeIDs, ShouldContain, "A")
				So(g.Subgraphs[0].NodeIDs, ShouldContain, "C")
			})
		})
	})
}
