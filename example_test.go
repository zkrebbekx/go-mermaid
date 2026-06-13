package mermaid_test

import (
	"errors"
	"fmt"
	"strings"

	mermaid "github.com/Zac300/go-mermaid"
)

// Render a simple flowchart to SVG.
func ExampleRender() {
	svg, err := mermaid.Render("graph TD\n  A[Start] --> B[End]")
	if err != nil {
		panic(err)
	}
	fmt.Println(strings.HasPrefix(string(svg), "<svg"))
	// Output: true
}

// Configure rendering with functional options.
func ExampleRender_options() {
	svg, _ := mermaid.Render("graph LR\n  A --> B",
		mermaid.WithTheme(mermaid.Dark),
		mermaid.WithPadding(24),
	)
	fmt.Println(strings.Contains(string(svg), "#1e1e1e")) // dark background
	// Output: true
}

// Render a sequence diagram; the diagram type is detected from the header.
func ExampleRender_sequence() {
	svg, _ := mermaid.Render("sequenceDiagram\n  Alice->>Bob: Hello")
	fmt.Println(strings.Contains(string(svg), ">Alice<"))
	// Output: true
}

// Recover the source position of a parse error.
func ExampleRender_errorHandling() {
	_, err := mermaid.Render("graph TD\n  A[unterminated")
	fmt.Println(errors.Is(err, mermaid.ErrParse))

	var pe *mermaid.ParseError
	if errors.As(err, &pe) {
		fmt.Printf("error on line %d\n", pe.Line)
	}
	// Output:
	// true
	// error on line 2
}
