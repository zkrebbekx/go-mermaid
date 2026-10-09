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

func TestLexLinkHeads(t *testing.T) {
	Convey("Given links with cross, circle and arrow heads", t, func() {
		cases := []struct {
			in   string
			want []string // Ident and Arrow token values, in order
		}{
			{"A --x B", []string{"A", "--x", "B"}},
			{"A --o B", []string{"A", "--o", "B"}},
			{"A x--x B", []string{"A", "x--x", "B"}},
			{"A o--o B", []string{"A", "o--o", "B"}},
			{"A -.-x B", []string{"A", "-.-x", "B"}},
			{"A ==o B", []string{"A", "==o", "B"}},
			{"A[a] x==x B", []string{"A", "x==x", "B"}},
			{"A <--> B", []string{"A", "<-->", "B"}},
			// x and o stay node IDs when they can't be heads.
			{"x --> y", []string{"x", "-->", "y"}},
			{"o --> x", []string{"o", "-->", "x"}},
			{"A --oops", []string{"A", "--", "oops"}},
			{"A --- ops", []string{"A", "---", "ops"}},
			{"A & x --> B", []string{"A", "x", "-->", "B"}},
			{"A x-1 --> B", []string{"A", "x-1", "-->", "B"}},
			{"A --> o", []string{"A", "-->", "o"}},
			// A head needs a full -- or == line, as in Mermaid.
			{"A -x B", []string{"A", "-", "x", "B"}},
			// A head at the end of a line is still a head, as in Mermaid.
			{"A --x", []string{"A", "--x"}},
			// Words in a middle-form label stay words.
			{"A -- x --> B", []string{"A", "--", "x", "-->", "B"}},
			{"A x-- x --x B", []string{"A", "x--", "x", "--x", "B"}},
		}
		for _, c := range cases {
			Convey("When lexing "+c.in, func() {
				toks, err := Lex(c.in)

				Convey("Then heads join the link and node IDs stay whole", func() {
					So(err, ShouldBeNil)
					var got []string
					for _, tk := range toks {
						if tk.Kind == Ident || tk.Kind == Arrow {
							got = append(got, tk.Val)
						}
					}
					So(got, ShouldResemble, c.want)
				})
			})
		}
	})
}
