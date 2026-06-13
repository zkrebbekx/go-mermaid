# go-mermaid

[![CI](https://github.com/Zac300/go-mermaid/actions/workflows/ci.yml/badge.svg)](https://github.com/Zac300/go-mermaid/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/Zac300/go-mermaid/branch/main/graph/badge.svg)](https://codecov.io/gh/Zac300/go-mermaid)
[![Go Reference](https://pkg.go.dev/badge/github.com/Zac300/go-mermaid.svg)](https://pkg.go.dev/github.com/Zac300/go-mermaid)
[![Go Report Card](https://goreportcard.com/badge/github.com/Zac300/go-mermaid)](https://goreportcard.com/report/github.com/Zac300/go-mermaid)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

Render [Mermaid](https://mermaid.js.org) diagrams to SVG in **pure Go** — no
headless browser, no Node.js, no JavaScript runtime. Just a library and a
single static binary.

> **Status:** pre-1.0, actively developed. 13 diagram types supported (see
> the [table below](#diagram-types)). Not affiliated with the Mermaid project;
> this is an independent, compatible renderer.

## Why

Every existing Go path to Mermaid SVG shells out to headless Chrome or a Node
sidecar. That's heavy, slow, and hard to deploy. `go-mermaid` does the parse →
layout → SVG pipeline natively, so you can render diagrams inside any Go service
or CLI with zero external dependencies.

## Install

Library:

```bash
go get github.com/Zac300/go-mermaid
```

CLI:

```bash
go install github.com/Zac300/go-mermaid/cmd/mermaid@latest
```

## Usage

### Library

```go
package main

import (
	"os"

	mermaid "github.com/Zac300/go-mermaid"
)

func main() {
	svg, err := mermaid.Render("graph TD\n  A[Start] --> B{OK?}\n  B -->|yes| C([Done])")
	if err != nil {
		panic(err)
	}
	os.WriteFile("diagram.svg", svg, 0o644)
}
```

With options:

```go
svg, err := mermaid.Render(src,
	mermaid.WithTheme(mermaid.Dark),
	mermaid.WithFont("Inter", 14),
	mermaid.WithPadding(24),
	mermaid.WithSpacing(50, 60),
)
```

### CLI

```bash
mermaid diagram.mmd > diagram.svg
echo "graph LR; A-->B-->C" | mermaid -theme dark -o out.svg
mermaid a.mmd b.mmd c.mmd      # batch: writes a.svg, b.svg, c.svg
mermaid -list-types            # list supported diagram types
mermaid -list-themes           # list themes
mermaid serve -addr :8080      # HTTP render endpoint (POST source, or GET ?src=)
```

## Error handling

`Render` wraps stage-specific sentinels so you can branch on the failure, and
parse errors carry source position:

```go
svg, err := mermaid.Render(src)
if errors.Is(err, mermaid.ErrParse) {
	var pe *mermaid.ParseError
	if errors.As(err, &pe) {
		log.Printf("syntax error at line %d col %d: %s", pe.Line, pe.Col, pe.Msg)
	}
}
```

Sentinels: `ErrParse`, `ErrLayout`, `ErrRender`, `ErrUnsupported`.

## Diagram types

The renderer dispatches on the diagram header. Status vs. Mermaid:

| Type | Header | Notes |
| --- | --- | --- |
| Flowchart | `graph` / `flowchart` | 12 shapes, subgraphs, styling, orthogonal edges |
| Sequence | `sequenceDiagram` | notes, activations, loop/alt/opt frames, autonumber |
| Class | `classDiagram` | members, 6 relationship types |
| State | `stateDiagram-v2` | start/end, composite states |
| Entity-relationship | `erDiagram` | attributes, crow's-foot cardinality |
| Pie | `pie` | legend with percentages |
| User journey | `journey` | scored tasks, section bands |
| Quadrant | `quadrantChart` | axes, 4 quadrants, points |
| Git graph | `gitGraph` | branches, merges, tags |
| Timeline | `timeline` | sections, periods, events |
| Mindmap | `mindmap` | indentation hierarchy |
| Gantt | `gantt` | dates, durations, `after` deps |
| C4 | `C4Context` / `C4Container` | people, systems, relationships |

Flowchart extras: curved edges (`WithCurvedEdges`), clickable nodes
(`click ID href`).

Unsupported types return `ErrUnsupported`. A `---`/`title:` front-matter block
and `accTitle:`/`accDescr:` (SVG `<title>`/`<desc>`) are honored for all types.
Themes: `default`, `dark`, `neutral`, `forest`, `base`.

Rendering is fast — roughly 10–50µs per diagram with no external processes.

### Flowchart syntax

| Feature | Example |
| --- | --- |
| Directions | `graph TD`, `TB`, `BT`, `LR`, `RL` |
| Rectangle | `A[Label]` |
| Rounded | `A(Label)` |
| Stadium | `A([Label])` |
| Circle | `A((Label))` |
| Diamond | `A{Label}` |
| Hexagon | `A{{Label}}` |
| Subroutine | `A[[Label]]` |
| Cylinder | `A[(Label)]` |
| Parallelogram | `A[/Label/]`, `A[\Label\]` |
| Trapezoid | `A[/Label\]`, `A[\Label/]` |
| Arrow / open / dotted / thick | `A --> B`, `A --- B`, `A -.-> B`, `A ==> B` |
| Edge label | `A -->\|text\| B`, `A -- text --> B` |
| Subgraph | `subgraph T` … `end` |
| Styling | `classDef`, `class`, `style`, `A:::cls` |
| Link | `click A href "url"` |
| Separators / comments | `A-->B; B-->C`, `%% comment` |

### Sequence syntax

| Feature | Example |
| --- | --- |
| Participant | `participant A` / `actor A` |
| Alias | `participant A as Alice` |
| Message + arrowhead | `A->>B: text` |
| Reply (dashed) | `B-->>A: text` |
| Plain line | `A->B` / `A-->B` |
| Cross end | `A-xB` / `A--xB` |
| Self-message | `A->>A: text` |

Notes, loops, alt/opt and activations are parsed but not yet drawn (skipped).

## Roadmap

- [x] Crossing minimization (median heuristic, dummy nodes for long edges)
- [x] Barycenter x-positioning (parents centered over children)
- [x] Multi-rank edge bends and self-loops
- [x] Sequence diagrams
- [x] Orthogonal edge routing, subgraphs, styling, 12 node shapes
- [x] Sequence notes, activations, loop/alt/opt frames, autonumber
- [x] 12 diagram types (see table above)
- [x] Front-matter titles, accessibility (title/desc), 5 themes
- [x] CLI batch render, `serve` HTTP mode, fuzz-tested parsers
- [ ] Network-simplex ranking (tighter flowchart layouts)
- [ ] Spline edge routing
- [ ] C4, requirement, sankey diagrams
- [ ] PNG output

## Architecture

Each diagram type has a self-contained pipeline; the public `Render` dispatches
on the header. The flowchart pipeline is:

```
source → lexer → parser → domain.Graph → layout → render → SVG
```

`internal/domain` holds the flowchart model; `internal/sequence` is the
self-contained sequence pipeline; `internal/theme` and `internal/svgutil` are
shared by all renderers. See [CONTRIBUTING.md](CONTRIBUTING.md).

## Releasing

Releases are automated with release-please + GoReleaser, driven by Conventional
Commits. See [RELEASING.md](RELEASING.md) for the versioning policy and the
optional AI code-review setup.

## License

[MIT](LICENSE) © Zac Krebbekx
