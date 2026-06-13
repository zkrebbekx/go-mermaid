package mermaid_test

import (
	"testing"

	mermaid "github.com/Zac300/go-mermaid"
)

var benchSources = map[string]string{
	"flowchart": "graph TD\nA[Start] --> B{ok?}\nB -->|yes| C([Done])\nB -->|no| D((Retry))\nD --> A\nC --> E\nE --> F\nF --> A",
	"sequence":  "sequenceDiagram\nAlice->>+Bob: Request\nNote over Bob: working\nBob-->>-Alice: Response\nAlice->>Alice: log",
	"class":     "classDiagram\nclass Animal {\n+int age\n+run() void\n}\nAnimal <|-- Dog\nAnimal *-- Leg\nDog --> Bone",
	"state":     "stateDiagram-v2\n[*] --> Idle\nIdle --> Running : start\nRunning --> Idle : stop\nRunning --> [*]",
	"er":        "erDiagram\nCUSTOMER ||--o{ ORDER : places\nORDER ||--|{ ITEM : contains\nCUSTOMER {\nstring name\n}",
	"pie":       "pie title Pets\n\"Dogs\" : 386\n\"Cats\" : 85\n\"Rats\" : 15",
}

func BenchmarkRender(b *testing.B) {
	for name, src := range benchSources {
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if _, err := mermaid.Render(src); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
