package mermaid

import (
	"fmt"
	"strings"
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
			{"unknown", "architecture-beta\n  x", kindUnknown},
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

func TestDiagramTypesMatchDetection(t *testing.T) {
	Convey("Given the advertised diagram types", t, func() {
		var headers []string
		for _, entry := range DiagramTypes() {
			headers = append(headers, strings.Split(entry, " / ")...)
		}

		Convey("When detecting each header", func() {
			seen := map[kind]bool{}
			for _, h := range headers {
				seen[detectKind(h)] = true
			}

			Convey("Then every header is recognised", func() {
				for _, h := range headers {
					SoMsg(h, detectKind(h), ShouldNotEqual, kindUnknown)
				}
			})

			Convey("Then every kind the dispatcher knows is advertised", func() {
				for k := kindUnknown + 1; k < kindEnd; k++ {
					SoMsg(fmt.Sprintf("kind %d", k), seen[k], ShouldBeTrue)
				}
			})
		})
	})
}
