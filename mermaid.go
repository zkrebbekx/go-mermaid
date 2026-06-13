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

	"github.com/Zac300/go-mermaid/internal/class"
	"github.com/Zac300/go-mermaid/internal/er"
	"github.com/Zac300/go-mermaid/internal/layout"
	"github.com/Zac300/go-mermaid/internal/lexer"
	"github.com/Zac300/go-mermaid/internal/parser"
	"github.com/Zac300/go-mermaid/internal/pie"
	"github.com/Zac300/go-mermaid/internal/render"
	"github.com/Zac300/go-mermaid/internal/sequence"
	"github.com/Zac300/go-mermaid/internal/state"
	"github.com/Zac300/go-mermaid/internal/syntax"
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

	switch detectKind(body) {
	case kindFlowchart:
		return renderFlowchart(body, cfg, title)
	case kindSequence:
		svg, err := sequence.Render(body, sequence.RenderOptions{
			Theme:    string(cfg.theme),
			FontFace: cfg.fontFace,
			FontSize: cfg.fontSize,
			Padding:  cfg.padding,
			Title:    title,
		})
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrParse, err)
		}
		return svg, nil
	case kindPie:
		svg, err := pie.Render(body, pie.RenderOptions{
			Theme:    string(cfg.theme),
			FontFace: cfg.fontFace,
			FontSize: cfg.fontSize,
			Padding:  cfg.padding,
			Title:    title,
		})
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrParse, err)
		}
		return svg, nil
	case kindClass:
		svg, err := class.Render(body, class.RenderOptions{
			Theme:    string(cfg.theme),
			FontFace: cfg.fontFace,
			FontSize: cfg.fontSize,
			Padding:  cfg.padding,
			Title:    title,
		})
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrParse, err)
		}
		return svg, nil
	case kindState:
		svg, err := state.Render(body, state.RenderOptions{
			Theme:    string(cfg.theme),
			FontFace: cfg.fontFace,
			FontSize: cfg.fontSize,
			Padding:  cfg.padding,
			Title:    title,
		})
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrParse, err)
		}
		return svg, nil
	case kindER:
		svg, err := er.Render(body, er.RenderOptions{
			Theme:    string(cfg.theme),
			FontFace: cfg.fontFace,
			FontSize: cfg.fontSize,
			Padding:  cfg.padding,
			Title:    title,
		})
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrParse, err)
		}
		return svg, nil
	default:
		return nil, fmt.Errorf("%w: unrecognized diagram type", ErrUnsupported)
	}
}

func renderFlowchart(src string, cfg config, title string) ([]byte, error) {
	tokens, err := lexer.Lex(src)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrParse, err)
	}

	graph, err := parser.Parse(tokens)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrParse, err)
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
