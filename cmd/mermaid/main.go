// Command mermaid renders Mermaid diagrams to SVG.
//
// Usage:
//
//	mermaid [flags] [input ...]
//
// With no input file (or "-"), source is read from stdin and SVG is written
// to stdout (or -o). With multiple input files, each FILE.mmd is rendered to
// FILE.svg and -o is not allowed.
//
//	mermaid diagram.mmd > diagram.svg
//	echo "graph TD; A-->B" | mermaid -theme dark -o out.svg
//	mermaid a.mmd b.mmd c.mmd      # writes a.svg, b.svg, c.svg
package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	mermaid "github.com/Zac300/go-mermaid"
)

// version is set at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "mermaid:", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) >= 2 && os.Args[1] == "serve" {
		fs := flag.NewFlagSet("serve", flag.ContinueOnError)
		addr := fs.String("addr", ":8080", "listen address")
		if err := fs.Parse(os.Args[2:]); err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "mermaid serving on", *addr)
		return http.ListenAndServe(*addr, renderHandler())
	}

	theme := flag.String("theme", "default", "color theme: default, dark, neutral")
	out := flag.String("o", "", "output file (single-input mode only; default stdout)")
	padding := flag.Float64("padding", 16, "outer padding in pixels")
	showVersion := flag.Bool("version", false, "print version and exit")
	listThemes := flag.Bool("list-themes", false, "list available themes and exit")
	listTypes := flag.Bool("list-types", false, "list supported diagram types and exit")
	flag.Usage = usage
	flag.Parse()

	switch {
	case *showVersion:
		fmt.Println("mermaid", version)
		return nil
	case *listThemes:
		for _, t := range mermaid.Themes() {
			fmt.Println(t)
		}
		return nil
	case *listTypes:
		for _, t := range mermaid.DiagramTypes() {
			fmt.Println(t)
		}
		return nil
	}

	opts := []mermaid.Option{
		mermaid.WithTheme(mermaid.Theme(*theme)),
		mermaid.WithPadding(*padding),
	}

	args := flag.Args()
	if len(args) > 1 {
		if *out != "" {
			return fmt.Errorf("-o cannot be used with multiple input files")
		}
		return renderBatch(args, opts)
	}

	src, err := readInput(firstArg(args))
	if err != nil {
		return err
	}
	svg, err := mermaid.Render(string(src), opts...)
	if err != nil {
		return err
	}
	if *out == "" || *out == "-" {
		_, err = os.Stdout.Write(svg)
		return err
	}
	return os.WriteFile(*out, svg, 0o644)
}

// renderBatch renders each input file to a sibling .svg file.
func renderBatch(files []string, opts []mermaid.Option) error {
	for _, f := range files {
		src, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		svg, err := mermaid.Render(string(src), opts...)
		if err != nil {
			return fmt.Errorf("%s: %w", f, err)
		}
		dst := strings.TrimSuffix(f, filepath.Ext(f)) + ".svg"
		if err := os.WriteFile(dst, svg, 0o644); err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "wrote", dst)
	}
	return nil
}

// renderHandler renders the request body (POST) or ?src= (GET) to SVG.
// The optional ?theme= query selects a theme.
func renderHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var src string
		if r.Method == http.MethodGet {
			src = r.URL.Query().Get("src")
		} else {
			body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			src = string(body)
		}
		opts := []mermaid.Option{}
		if t := r.URL.Query().Get("theme"); t != "" {
			opts = append(opts, mermaid.WithTheme(mermaid.Theme(t)))
		}
		svg, err := mermaid.Render(src, opts...)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "image/svg+xml")
		_, _ = w.Write(svg)
	}
}

func readInput(path string) ([]byte, error) {
	if path == "" || path == "-" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(path)
}

func firstArg(args []string) string {
	if len(args) == 0 {
		return ""
	}
	return args[0]
}

func usage() {
	fmt.Fprintf(os.Stderr, `mermaid %s - render Mermaid diagrams to SVG (pure Go)

Usage:
  mermaid [flags] [input ...]

Supported diagrams: flowchart, sequence, class, state, ER, pie.

Flags:
`, version)
	flag.PrintDefaults()
}
