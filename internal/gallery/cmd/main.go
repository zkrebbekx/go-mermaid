// Command gallery regenerates the README example images. Run it from the
// repository root after any change that alters rendering:
//
//	go run ./internal/gallery/cmd
//
// With -check it reports any image whose source diagram now renders
// differently, without writing. CI uses that form.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/zkrebbekx/go-mermaid/internal/gallery"
)

func main() {
	check := flag.Bool("check", false, "report drift instead of writing")
	source := flag.String("source", filepath.Join("testdata", "golden"), "directory holding the .mmd fixtures")
	target := flag.String("target", filepath.Join("docs", "gallery"), "directory holding the .png images")
	flag.Parse()

	if err := run(*check, *source, *target); err != nil {
		fmt.Fprintln(os.Stderr, "gallery:", err)
		os.Exit(1)
	}
}

func run(check bool, source, target string) error {
	if !check {
		return gallery.Write(source, target)
	}
	stale, err := gallery.Stale(source, target)
	if err != nil {
		return err
	}
	if len(stale) > 0 {
		return fmt.Errorf("these images are out of date, run `go run ./internal/gallery/cmd`:\n  %s",
			strings.Join(stale, "\n  "))
	}
	return nil
}
