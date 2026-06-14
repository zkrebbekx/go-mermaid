# Contributing

Thanks for your interest in improving `go-mermaid`.

## Getting started

```bash
git clone https://github.com/Zac300/go-mermaid
cd go-mermaid
make test
```

Useful targets (`make help` lists all):

- `make test` — run tests with the race detector
- `make cover` — HTML coverage report
- `make lint` — run golangci-lint
- `make golden` — regenerate golden SVG files after intentional output changes

## Architecture

The public `Render` detects the diagram type from the header and dispatches.
Flowcharts use the staged pipeline `lexer → parser → domain.Graph → layout →
render`:

- `internal/domain` — the pure flowchart model. No I/O, no third-party deps.
- `internal/lexer` — source text to tokens.
- `internal/parser` — tokens to a `domain.Graph` (plus styling/link preprocess).
- `internal/layout` — Sugiyama-style layered layout (acyclic → rank → order → position).
- `internal/render` — laid-out graph to SVG.
- `internal/syntax` — shared positional error type.
- `internal/theme`, `internal/svgutil` — palettes and SVG helpers shared by all renderers.

Every other diagram type (sequence, class, state, er, pie, journey, quadrant,
git, timeline, mindmap, gantt, c4, requirement, sankey, xychart, block, kanban,
packet, radar) is a self-contained `internal/<type>` package that parses, lays
out, and renders its own SVG, wired into the root dispatch. Several reuse
`internal/layout` for node positioning.

The core library emits SVG and has no third-party dependencies. PNG output
lives in the separate `raster` package (it depends on oksvg/rasterx); the core
never imports it, so SVG-only consumers stay dependency-free.

Keep stages decoupled: a change to rendering shouldn't require touching the
parser, and vice versa.

## Tests

- Table-driven unit tests plus BDD-style `Convey` blocks
  ([goconvey](https://github.com/smartystreets/goconvey)) for behavior.
- Golden-file tests compare full SVG output against committed `.svg` files in
  `testdata/golden`. After an intentional rendering change, run `make golden`
  and review the diff before committing.

## Pull requests

1. Open an issue first for anything non-trivial.
2. Add tests for new behavior; keep coverage from regressing.
3. Run `make lint test` before pushing.
4. Use [Conventional Commits](https://www.conventionalcommits.org) for messages.

## Code of conduct

This project follows the [Contributor Covenant](CODE_OF_CONDUCT.md).
