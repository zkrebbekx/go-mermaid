// Command mermaid renders Mermaid diagrams to SVG.
//
// Usage:
//
//	mermaid [flags] [input ...]
//
// With no input file (or "-"), source is read from stdin and SVG is written
// to stdout (or -o). With multiple input files, each FILE.mmd is rendered to
// FILE.svg and -o is not allowed. Pass -png (or an .png -o path) for PNG.
//
//	mermaid diagram.mmd > diagram.svg
//	echo "graph TD; A-->B" | mermaid -theme dark -o out.svg
//	mermaid a.mmd b.mmd c.mmd      # writes a.svg, b.svg, c.svg
//	mermaid -png -scale 2 -o d.png d.mmd
//	mermaid serve -addr :8080      # GET/POST; ?format=png&scale=N for PNG
package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	mermaid "github.com/zkrebbekx/go-mermaid"
	"github.com/zkrebbekx/go-mermaid/raster"
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
	pngOut := flag.Bool("png", false, "output PNG instead of SVG")
	scale := flag.Float64("scale", 1, "PNG scale factor")
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
	asPNG := *pngOut || strings.HasSuffix(*out, ".png")

	args := flag.Args()
	if len(args) > 1 {
		if *out != "" {
			return fmt.Errorf("-o cannot be used with multiple input files")
		}
		return renderBatch(args, opts, asPNG, *scale)
	}

	src, err := readInput(firstArg(args))
	if err != nil {
		return err
	}
	data, err := renderBytes(string(src), opts, asPNG, *scale)
	if err != nil {
		return err
	}
	if *out == "" || *out == "-" {
		_, err = os.Stdout.Write(data)
		return err
	}
	return os.WriteFile(*out, data, 0o644)
}

// renderBytes produces SVG or PNG bytes for the given source.
func renderBytes(src string, opts []mermaid.Option, asPNG bool, scale float64) ([]byte, error) {
	if asPNG {
		return raster.PNG(src, scale, opts...)
	}
	return mermaid.Render(src, opts...)
}

// renderBatch renders each input file to a sibling .svg or .png file.
func renderBatch(files []string, opts []mermaid.Option, asPNG bool, scale float64) error {
	ext := ".svg"
	if asPNG {
		ext = ".png"
	}
	for _, f := range files {
		src, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		data, err := renderBytes(string(src), opts, asPNG, scale)
		if err != nil {
			return fmt.Errorf("%s: %w", f, err)
		}
		dst := strings.TrimSuffix(f, filepath.Ext(f)) + ext
		if err := os.WriteFile(dst, data, 0o644); err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "wrote", dst)
	}
	return nil
}

// renderHandler renders the request body (POST) or ?src= (GET) to SVG, or to
// PNG when ?format=png. Optional queries: ?theme=, ?scale= (PNG).
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

		if r.URL.Query().Get("format") == "png" {
			scale := 1.0
			if s, err := strconv.ParseFloat(r.URL.Query().Get("scale"), 64); err == nil && s > 0 {
				scale = s
			}
			data, err := raster.PNG(src, scale, opts...)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write(data)
			return
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
