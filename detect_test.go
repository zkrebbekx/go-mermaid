package mermaid

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDetectKind(t *testing.T) {
	Convey("Given diagram sources", t, func() {
		cases := []struct {
			name string
			src  string
			want kind
		}{
			{"flowchart graph", "graph TD\nA-->B", kindFlowchart},
			{"flowchart keyword", "flowchart LR\nA-->B", kindFlowchart},
			{"sequence", "sequenceDiagram\nA->>B: hi", kindSequence},
			{"leading comment", "%% title\nsequenceDiagram\nA->>B: hi", kindSequence},
			{"leading blank lines", "\n\n  graph TD\nA-->B", kindFlowchart},
			{"unknown", "classDiagram\nClass01", kindUnknown},
			{"empty", "", kindUnknown},
		}
		for _, c := range cases {
			c := c
			Convey("When detecting the "+c.name+" case", func() {
				got := detectKind(c.src)

				Convey("Then the diagram kind matches", func() {
					So(got, ShouldEqual, c.want)
				})
			})
		}
	})
}
