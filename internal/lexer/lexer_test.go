package lexer

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func kinds(toks []Token) []Kind {
	ks := make([]Kind, 0, len(toks))
	for _, t := range toks {
		ks = append(ks, t.Kind)
	}
	return ks
}

func TestLex(t *testing.T) {
	Convey("Given Mermaid flowchart source", t, func() {

		Convey("When lexing a header and a simple edge", func() {
			toks, err := Lex("graph TD\nA --> B")

			Convey("Then it produces the expected token kinds", func() {
				So(err, ShouldBeNil)
				So(kinds(toks), ShouldResemble, []Kind{
					Keyword, Ident, // graph TD
					Newline,
					Ident, Arrow, Ident, // A --> B
					EOF,
				})
			})
		})

		Convey("When lexing shapes and an inline edge label", func() {
			toks, err := Lex("graph LR\nA[Start] -->|go| B((End))")

			Convey("Then shape and pipe text are captured as Text tokens", func() {
				So(err, ShouldBeNil)
				var texts []string
				for _, tk := range toks {
					if tk.Kind == Text {
						texts = append(texts, tk.Val)
					}
				}
				So(texts, ShouldResemble, []string{"Start", "go", "End"})
			})
		})

		Convey("When a shape is left unterminated", func() {
			_, err := Lex("graph TD\nA[Start --> B")

			Convey("Then it returns an error", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestLexTable(t *testing.T) {
	Convey("Given assorted flowchart snippets", t, func() {
		cases := []struct {
			name string
			in   string
			want []Kind
		}{
			{"bare nodes", "graph TD\nA", []Kind{Keyword, Ident, Newline, Ident, EOF}},
			{"dotted arrow", "graph TD\nA -.-> B", []Kind{Keyword, Ident, Newline, Ident, Arrow, Ident, EOF}},
			{"comment skipped", "graph TD\n%% note\nA --> B", []Kind{Keyword, Ident, Newline, Newline, Ident, Arrow, Ident, EOF}},
		}
		for _, c := range cases {
			c := c
			Convey("When lexing the "+c.name+" case", func() {
				toks, err := Lex(c.in)

				Convey("Then the token kinds match", func() {
					So(err, ShouldBeNil)
					So(kinds(toks), ShouldResemble, c.want)
				})
			})
		}
	})
}

func TestLexQuoted(t *testing.T) {
	Convey("Given a quoted middle-form edge label", t, func() {
		toks, err := Lex(`A -- "has (parens)" --> B`)

		Convey("When lexing", func() {
			Convey("Then the quotes become one Text token without them", func() {
				So(err, ShouldBeNil)
				So(kinds(toks), ShouldResemble, []Kind{Ident, Arrow, Text, Arrow, Ident, EOF})
				So(toks[2].Val, ShouldEqual, "has (parens)")
				So(toks[2].Col, ShouldEqual, 6)
			})
		})
	})

	Convey("Given a quote closed only on a later line", t, func() {
		_, err := Lex("A -- \"open\n\" --> B")

		Convey("When lexing", func() {
			Convey("Then it reports an unterminated string at the opening quote", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "line 1 col 6")
				So(err.Error(), ShouldContainSubstring, "unterminated string")
			})
		})
	})

	Convey("Given a quote that is never closed", t, func() {
		_, err := Lex(`A -- "open`)

		Convey("When lexing", func() {
			Convey("Then it reports an unterminated string", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "unterminated string")
			})
		})
	})
}
