package mermaid_test

import (
	"errors"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	mermaid "github.com/zkrebbekx/go-mermaid"
)

// update regenerates golden files: go test -run TestGolden -update
var update = flag.Bool("update", false, "update golden SVG files")

func TestRender(t *testing.T) {
	Convey("Given the public Render API", t, func() {

		Convey("When rendering a minimal flowchart", func() {
			out, err := mermaid.Render("graph TD\nA --> B")

			Convey("Then it returns valid-looking SVG bytes", func() {
				So(err, ShouldBeNil)
				So(string(out), ShouldStartWith, "<svg")
				So(string(out), ShouldContainSubstring, "</svg>")
			})
		})

		Convey("When options are supplied", func() {
			out, err := mermaid.Render("graph TD\nA --> B", mermaid.WithTheme(mermaid.Dark))

			Convey("Then the dark background is applied", func() {
				So(err, ShouldBeNil)
				So(string(out), ShouldContainSubstring, "#1e1e1e")
			})
		})

		Convey("When the source has a syntax error", func() {
			_, err := mermaid.Render("graph TD\nA[unterminated")

			Convey("Then ErrParse matches and a ParseError is recoverable", func() {
				So(errors.Is(err, mermaid.ErrParse), ShouldBeTrue)
				var pe *mermaid.ParseError
				So(errors.As(err, &pe), ShouldBeTrue)
			})
		})

		Convey("When rendering a sequence diagram", func() {
			out, err := mermaid.Render("sequenceDiagram\nA->>B: hi")

			Convey("Then it dispatches to the sequence renderer", func() {
				So(err, ShouldBeNil)
				So(string(out), ShouldContainSubstring, "seq-arrow")
			})
		})

		Convey("When using RenderTo with a writer", func() {
			var buf strings.Builder
			err := mermaid.RenderTo(&buf, "graph TD\nA-->B")

			Convey("Then the SVG is written to the writer", func() {
				So(err, ShouldBeNil)
				So(buf.String(), ShouldStartWith, "<svg")
			})
		})

		Convey("When RenderTo gets invalid source", func() {
			var buf strings.Builder
			err := mermaid.RenderTo(&buf, "graph TD\nA[oops")

			Convey("Then it returns the error and writes nothing", func() {
				So(err, ShouldNotBeNil)
				So(buf.Len(), ShouldEqual, 0)
			})
		})

		Convey("When the source has a front-matter title", func() {
			out, err := mermaid.Render("---\ntitle: Pipeline\n---\ngraph LR\nA-->B")

			Convey("Then the title is rendered in bold", func() {
				So(err, ShouldBeNil)
				So(string(out), ShouldContainSubstring, ">Pipeline<")
				So(string(out), ShouldContainSubstring, `font-weight="bold"`)
			})
		})

		Convey("When the diagram type is unsupported", func() {
			_, err := mermaid.Render("architecture-beta\n  x")

			Convey("Then ErrUnsupported is returned", func() {
				So(errors.Is(err, mermaid.ErrUnsupported), ShouldBeTrue)
			})
		})
	})
}

func TestGolden(t *testing.T) {
	Convey("Given the golden diagram inputs", t, func() {
		inputs, err := filepath.Glob(filepath.Join("testdata", "golden", "*.mmd"))
		So(err, ShouldBeNil)
		So(len(inputs), ShouldBeGreaterThan, 0)

		for _, in := range inputs {
			in := in
			name := strings.TrimSuffix(filepath.Base(in), ".mmd")

			Convey("When rendering "+name, func() {
				src, readErr := os.ReadFile(in)
				So(readErr, ShouldBeNil)
				got, renderErr := mermaid.Render(string(src))
				So(renderErr, ShouldBeNil)

				goldenPath := strings.TrimSuffix(in, ".mmd") + ".svg"
				if *update {
					So(os.WriteFile(goldenPath, got, 0o644), ShouldBeNil)
					return
				}

				Convey("Then it matches the committed golden SVG", func() {
					want, goldErr := os.ReadFile(goldenPath)
					So(goldErr, ShouldBeNil)
					So(string(got), ShouldEqual, string(want))
				})
			})
		}
	})
}
