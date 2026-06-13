package mermaid_test

import (
	"strings"
	"testing"

	mermaid "github.com/Zac300/go-mermaid"
)

// FuzzRender checks that Render never panics on arbitrary input and that a
// nil error always comes with non-empty output.
func FuzzRender(f *testing.F) {
	seeds := []string{
		"graph TD\nA[Start] --> B{ok?}\nB -->|yes| C\nB -->|no| A",
		"flowchart LR\nA --> A",
		"sequenceDiagram\nA->>+B: hi\nNote over A,B: x\nB-->>-A: bye",
		"pie title P\n\"a\" : 1\n\"b\" : 2",
		"classDiagram\nclass A{\n+int x\n}\nA <|-- B",
		"stateDiagram-v2\n[*] --> A\nA --> [*]",
		"erDiagram\nA ||--o{ B : has\nA{\nstring n\n}",
		"---\ntitle: T\n---\ngraph TD\nA-->B",
		"",
		"graph",
		"pie\n: :",
		"classDiagram\n<|--",
	}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, src string) {
		out, err := mermaid.Render(src)
		if err == nil {
			if len(out) == 0 {
				t.Fatalf("nil output without error for %q", src)
			}
			if !strings.HasPrefix(string(out), "<svg") {
				t.Fatalf("output is not SVG for %q", src)
			}
		}
	})
}
