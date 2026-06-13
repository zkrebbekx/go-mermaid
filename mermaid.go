// Package mermaid renders Mermaid diagrams to SVG in pure Go, with no
// headless browser or JavaScript runtime required.
//
// The v0 surface targets flowcharts (graph TD / graph LR). Other diagram
// types are planned; see the package roadmap in the README.
//
// Basic use:
//
//	svg, err := mermaid.Render("graph TD\n  A --> B")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	os.Stdout.Write(svg)
//
// Rendering is configured with functional options:
//
//	svg, err := mermaid.Render(src,
//	    mermaid.WithTheme(mermaid.Dark),
//	    mermaid.WithPadding(24),
//	)
package mermaid

import (
	"fmt"
	"io"

	"github.com/Zac300/go-mermaid/internal/class"
	"github.com/Zac300/go-mermaid/internal/er"
	"github.com/Zac300/go-mermaid/internal/gantt"
	gitgraph "github.com/Zac300/go-mermaid/internal/git"
	"github.com/Zac300/go-mermaid/internal/journey"
	"github.com/Zac300/go-mermaid/internal/layout"
	"github.com/Zac300/go-mermaid/internal/lexer"
	"github.com/Zac300/go-mermaid/internal/mindmap"
	"github.com/Zac300/go-mermaid/internal/parser"
	"github.com/Zac300/go-mermaid/internal/pie"
	"github.com/Zac300/go-mermaid/internal/quadrant"
	"github.com/Zac300/go-mermaid/internal/render"
	"github.com/Zac300/go-mermaid/internal/sequence"
	"github.com/Zac300/go-mermaid/internal/state"
	"github.com/Zac300/go-mermaid/internal/syntax"
	"github.com/Zac300/go-mermaid/internal/timeline"
)

// ParseError reports a lexing or parsing failure with its source position
// (Line, Col). When Render fails during the parse stage, the returned error
// wraps a *ParseError recoverable with errors.As.
type ParseError = syntax.Error

// Render parses a Mermaid source document and returns the rendered SVG.
//
// Errors are wrapped so callers can match the failing stage with
// errors.Is (ErrParse, ErrLayout, ErrRender, ErrUnsupported) and, for
// parse failures, recover source position with errors.As(&ParseError{}).
func Render(src string, opts ...Option) (out []byte, err error) {
	// Safety net: diagram source is untrusted, so a bug in any parser or
	// renderer must surface as an error, never a panic in the caller.
	defer func() {
		if r := recover(); r != nil {
			out, err = nil, fmt.Errorf("%w: internal panic: %v", ErrRender, r)
		}
	}()

	cfg := defaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	title, body := parseFrontmatter(src)
	accTitle, accDescr, body := extractA11y(body)

	var raw []byte
	switch detectKind(body) {
	case kindFlowchart:
		raw, err = renderFlowchart(body, cfg, title)
	case kindSequence:
		raw, err = sequence.Render(body, sequence.RenderOptions{
			Theme: string(cfg.theme), FontFace: cfg.fontFace, FontSize: cfg.fontSize, Padding: cfg.padding, Title: title,
		})
		err = wrapParse(err)
	case kindPie:
		raw, err = pie.Render(body, pie.RenderOptions{
			Theme: string(cfg.theme), FontFace: cfg.fontFace, FontSize: cfg.fontSize, Padding: cfg.padding, Title: title,
		})
		err = wrapParse(err)
	case kindClass:
		raw, err = class.Render(body, class.RenderOptions{
			Theme: string(cfg.theme), FontFace: cfg.fontFace, FontSize: cfg.fontSize, Padding: cfg.padding, Title: title,
		})
		err = wrapParse(err)
	case kindState:
		raw, err = state.Render(body, state.RenderOptions{
			Theme: string(cfg.theme), FontFace: cfg.fontFace, FontSize: cfg.fontSize, Padding: cfg.padding, Title: title,
		})
		err = wrapParse(err)
	case kindER:
		raw, err = er.Render(body, er.RenderOptions{
			Theme: string(cfg.theme), FontFace: cfg.fontFace, FontSize: cfg.fontSize, Padding: cfg.padding, Title: title,
		})
		err = wrapParse(err)
	case kindJourney:
		raw, err = journey.Render(body, journey.RenderOptions{
			Theme: string(cfg.theme), FontFace: cfg.fontFace, FontSize: cfg.fontSize, Padding: cfg.padding, Title: title,
		})
		err = wrapParse(err)
	case kindQuadrant:
		raw, err = quadrant.Render(body, quadrant.RenderOptions{
			Theme: string(cfg.theme), FontFace: cfg.fontFace, FontSize: cfg.fontSize, Padding: cfg.padding, Title: title,
		})
		err = wrapParse(err)
	case kindGit:
		raw, err = gitgraph.Render(body, gitgraph.RenderOptions{
			Theme: string(cfg.theme), FontFace: cfg.fontFace, FontSize: cfg.fontSize, Padding: cfg.padding, Title: title,
		})
		err = wrapParse(err)
	case kindTimeline:
		raw, err = timeline.Render(body, timeline.RenderOptions{
			Theme: string(cfg.theme), FontFace: cfg.fontFace, FontSize: cfg.fontSize, Padding: cfg.padding, Title: title,
		})
		err = wrapParse(err)
	case kindMindmap:
		raw, err = mindmap.Render(body, mindmap.RenderOptions{
			Theme: string(cfg.theme), FontFace: cfg.fontFace, FontSize: cfg.fontSize, Padding: cfg.padding, Title: title,
		})
		err = wrapParse(err)
	case kindGantt:
		raw, err = gantt.Render(body, gantt.RenderOptions{
			Theme: string(cfg.theme), FontFace: cfg.fontFace, FontSize: cfg.fontSize, Padding: cfg.padding, Title: title,
		})
		err = wrapParse(err)
	default:
		return nil, fmt.Errorf("%w: unrecognized diagram type", ErrUnsupported)
	}
	if err != nil {
		return nil, err
	}

	a11yTitle := accTitle
	if a11yTitle == "" {
		a11yTitle = title
	}
	return injectA11y(raw, a11yTitle, accDescr), nil
}

// RenderTo renders src and writes the SVG to w. It is a convenience over
// Render for HTTP handlers and file pipelines.
func RenderTo(w io.Writer, src string, opts ...Option) error {
	svg, err := Render(src, opts...)
	if err != nil {
		return err
	}
	_, err = w.Write(svg)
	return err
}

// wrapParse tags a sub-renderer error as a parse-stage failure.
func wrapParse(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%w: %w", ErrParse, err)
}

func renderFlowchart(src string, cfg config, title string) ([]byte, error) {
	src, styles, links := parser.Preprocess(src)

	tokens, err := lexer.Lex(src)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrParse, err)
	}

	graph, err := parser.Parse(tokens)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrParse, err)
	}

	for id, st := range styles {
		if n := graph.NodeByID(id); n != nil {
			n.Style = st
		}
	}
	for id, url := range links {
		if n := graph.NodeByID(id); n != nil {
			n.Link = url
		}
	}

	laid, err := layout.Compute(graph, cfg.layout())
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrLayout, err)
	}

	ro := cfg.render()
	ro.Title = title
	svg, err := render.SVG(laid, ro)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrRender, err)
	}

	return svg, nil
}
