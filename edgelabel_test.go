package mermaid_test

import (
	"regexp"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	mermaid "github.com/zkrebbekx/go-mermaid"
)

func TestEdgeLabelsHaveBackground(t *testing.T) {
	cases := []struct {
		name, src, label string
	}{
		{"flowchart", "flowchart LR\nA -->|go| B", "go"},
		{"state", "stateDiagram-v2\ndirection LR\nIdle --> Busy : start work", "start work"},
		{"class", "classDiagram\ndirection LR\nAnimal <|-- Duck : inherits", "inherits"},
		{"er", "erDiagram\nCUSTOMER ||--o{ ORDER : places", "places"},
		{"c4", "C4Context\nSystem(a, \"A\")\nSystem(b, \"B\")\nRel(a, b, \"Uses\", \"HTTPS\")", "Uses"},
		{"requirement", "requirementDiagram\nrequirement r {\nid: 1\n}\nelement e {\ntype: x\n}\ne - satisfies -> r", "«satisfies»"},
	}
	for _, tc := range cases {
		Convey("Given a "+tc.name+" diagram with an edge label", t, func() {
			Convey("When rendering it", func() {
				svg, err := mermaid.Render(tc.src)

				Convey("Then the label is drawn over a background box, so the edge does not run through it", func() {
					So(err, ShouldBeNil)
					// The box, then the label's text lines (C4 draws its tech line first).
					re := regexp.MustCompile(`<rect [^>]*fill="#ffffff"/>\n(\s*<text[^\n]*\n)?\s*<text [^>]*>` +
						regexp.QuoteMeta(tc.label) + `</text>`)
					So(re.Match(svg), ShouldBeTrue)
				})
			})
		})
	}
}
