# CLAUDE.md

Guidance for AI assistants and contributors working in this repository.

## What this is

`go-mermaid` renders Mermaid diagrams to SVG in pure Go — no headless browser,
no Node, no JavaScript runtime. Supports 13 diagram types (flowchart,
sequence, class, state, ER, pie, journey, quadrant, gitGraph, timeline,
mindmap, gantt, C4).

## Architecture

A linear, testable pipeline. Each stage produces or consumes the pure model in
`internal/domain`.

```
source → lexer → parser → domain.Graph → layout → render → SVG
```

| Package             | Responsibility                                             |
| ------------------- | ---------------------------------------------------------- |
| `internal/domain`   | Pure flowchart model (Graph, Node, Edge, geometry). No I/O. |
| `internal/lexer`    | Flowchart source text → tokens.                            |
| `internal/parser`   | Tokens → `domain.Graph`.                                   |
| `internal/layout`   | Sugiyama layout: acyclic → rank → order → position.        |
| `internal/render`   | Laid-out flowchart graph → SVG bytes.                      |
| `internal/sequence` | Self-contained sequence pipeline: parse → layout → render. |
| `internal/theme`    | Shared color palettes (all diagram types).                 |
| `internal/svgutil`  | Shared SVG helpers (number formatting, XML escaping).      |
| `internal/syntax`   | Shared positional error type.                              |
| `.` (root)          | Public API: `Render`, diagram-type dispatch, options, errors. |
| `cmd/mermaid`       | CLI wrapper over the library.                              |

Diagram-type packages (each self-contained: parse → layout/place → render,
wired into the root `detectKind` dispatch): `internal/sequence`, `class`,
`state`, `er`, `pie`, `journey`, `quadrant`, `git`, `timeline`, `mindmap`,
`gantt`. Flowchart is the exception, split across `lexer`/`parser`/`layout`/
`render`. New types should follow the self-contained package shape.

Keep stages decoupled. A rendering change must not require touching the parser,
and vice versa. New dependencies between `internal/*` packages should follow the
pipeline direction only (no cycles; `domain` and `syntax` are leaves).

## Commands

```bash
make test     # go test -race ./...
make cover    # HTML coverage report
make lint     # golangci-lint
make golden   # regenerate golden SVGs after an intentional output change
make build    # build the CLI to bin/mermaid
```

Always run `make lint test` before committing. After any change that alters SVG
output, run `make golden` and review the diff — golden files are the contract.

## Coding standards

Follow the Go standard library style ([Effective Go](https://go.dev/doc/effective_go),
[Google Go Style Guide](https://google.github.io/styleguide/go/)). Specifics
that matter here:

### Comments

- **Document every exported identifier** with a doc comment starting with the
  identifier's name. This is the stdlib convention and `revive` enforces it.
- **Do not over-comment internal logic.** Code should read on its own. Add an
  inline comment only when the *why* is non-obvious: a chosen algorithm, a
  non-trivial invariant, a workaround, or a subtle edge case.
- Never narrate the obvious. Delete comments that restate the code.

```go
// Good — explains a non-obvious decision.
// Reverse back edges so ranking sees a DAG; restored after positioning.
e.From, e.To = e.To, e.From

// Bad — restates the code.
// increment the counter
count++
```

### Naming

- Short, lower-case package names; no `util`/`common`/`helpers` grab-bags.
- Receivers: short and consistent (`p *parser`, `l *lexer`).
- Exported names carry the weight of their package: prefer `mermaid.Render`,
  not `mermaid.RenderMermaid`.

### Errors

- Wrap with `%w` to preserve the chain. The public `Render` wraps each stage in
  a sentinel (`ErrParse`, `ErrLayout`, `ErrRender`) so callers can `errors.Is`.
- Parse/lex failures use the positional `syntax.Error` (exposed publicly as
  `mermaid.ParseError`) so callers can `errors.As` for line/column.
- No `panic` in library code for input errors — return an error. Untrusted
  diagram source must never crash the renderer.

### API design

- Prefer one obvious entry point (`Render`) plus functional options
  (`WithTheme`, `WithFont`, …). Keep the surface small.
- Zero-value config must be sensible; options only override defaults.
- Keep `internal/` truly internal. Anything users need (e.g. the error type)
  must be re-exported from the root package, typically via a type alias.

### Formatting

- `gofmt -s` and `goimports` clean. Tabs for indentation (Go default).
- No commented-out code. Delete it; git remembers.

## Testing

- Table-driven tests for input/output matrices.
- BDD `Convey` blocks ([goconvey](https://github.com/smartystreets/goconvey))
  for behavior, using Given/When/Then nesting.
- Golden-file tests (`testdata/golden/*.mmd` + `*.svg`) for full render output.
  Regenerate with `make golden`; review diffs before committing.
- New behavior needs tests. Don't let coverage regress.

```go
func TestParse(t *testing.T) {
    Convey("Given a flowchart source", t, func() {
        Convey("When parsing one edge", func() {
            g, err := parse("graph LR\nA --> B")
            Convey("Then the edge is captured", func() {
                So(err, ShouldBeNil)
                So(len(g.Edges), ShouldEqual, 1)
            })
        })
    })
}
```

## Commits

- [Conventional Commits](https://www.conventionalcommits.org): `feat:`, `fix:`,
  `docs:`, `refactor:`, `test:`, `chore:`.
- Subject in imperative mood, ≤ ~72 chars. Body explains *why* when non-obvious.

## Roadmap context

The flowchart layout is Sugiyama-style: cycle removal, longest-path ranking,
median crossing minimization with dummy nodes for long edges, and barycenter
x-positioning. Edges are orthogonal by default with an optional curved mode.
Ranking is still longest-path (network-simplex would tighten it). Remaining
ideas: network-simplex ranking, spline routing, and more diagram types
(requirement, sankey) plus PNG export. Keep new code behind the existing stage
interfaces so these land without API churn.
