// Package gallery regenerates the README example images from the golden
// diagram sources. The images are committed, so a rendering change must be
// followed by a regeneration or the README shows output the code no longer
// produces.
//
// Drift is detected through the SVG each image is rendered from, not through
// the image bytes. SVG output is byte-identical on every platform, because
// every coordinate is rounded to two decimals before it is written.
// Rasterizing to PNG is not: it differs in the last bit between amd64 and
// arm64, so comparing image bytes would fail on whichever architecture did
// not generate them.
package gallery

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	mermaid "github.com/zkrebbekx/go-mermaid"
	"github.com/zkrebbekx/go-mermaid/raster"
)

// Scale is the pixel density of the generated images.
const Scale = 2

// ManifestName is the file recording the SVG digest behind each image.
const ManifestName = "manifest.txt"

// Names are the diagram fixtures shown in the README, without extensions.
// Each must exist as <sourceDir>/<name>.mmd.
var Names = []string{
	"class", "er", "gantt", "gitgraph", "mindmap",
	"pie", "radar", "sequence", "simple", "state",
}

// Image is one rendered gallery entry.
type Image struct {
	Name string
	PNG  []byte
	// Digest is the SHA-256 of the SVG the image was rendered from.
	Digest string
}

// Render reads every fixture under sourceDir and renders it. The results come
// back sorted by name so the output is stable.
func Render(sourceDir string) ([]Image, error) {
	names := append([]string(nil), Names...)
	sort.Strings(names)

	out := make([]Image, 0, len(names))
	for _, name := range names {
		path := filepath.Join(sourceDir, name+".mmd")
		src, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}
		svg, err := mermaid.Render(string(src))
		if err != nil {
			return nil, fmt.Errorf("render %s: %w", name, err)
		}
		png, err := raster.RasterizeSVG(svg, Scale)
		if err != nil {
			return nil, fmt.Errorf("rasterize %s: %w", name, err)
		}
		sum := sha256.Sum256(svg)
		out = append(out, Image{Name: name, PNG: png, Digest: hex.EncodeToString(sum[:])})
	}
	return out, nil
}

// Write renders the fixtures, writes each image into targetDir, and records
// the SVG digests in the manifest.
func Write(sourceDir, targetDir string) error {
	images, err := Render(sourceDir)
	if err != nil {
		return err
	}
	for _, img := range images {
		path := filepath.Join(targetDir, img.Name+".png")
		if err := os.WriteFile(path, img.PNG, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
	}
	return writeManifest(filepath.Join(targetDir, ManifestName), images)
}

// Stale returns the names whose committed image was rendered from a different
// SVG than the code produces now, plus any image file that is missing.
func Stale(sourceDir, targetDir string) ([]string, error) {
	images, err := Render(sourceDir)
	if err != nil {
		return nil, err
	}
	recorded, err := ReadManifest(filepath.Join(targetDir, ManifestName))
	if err != nil {
		return nil, err
	}
	var stale []string
	for _, img := range images {
		if recorded[img.Name] != img.Digest {
			stale = append(stale, img.Name)
			continue
		}
		if _, statErr := os.Stat(filepath.Join(targetDir, img.Name+".png")); statErr != nil {
			stale = append(stale, img.Name)
		}
	}
	return stale, nil
}

// ReadManifest loads the recorded digests. A missing file reads as empty, so
// the first run reports every image as stale rather than failing.
func ReadManifest(path string) (map[string]string, error) {
	out := map[string]string{}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return out, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		name, digest, ok := strings.Cut(line, " ")
		if !ok {
			continue
		}
		out[strings.TrimSpace(name)] = strings.TrimSpace(digest)
	}
	return out, nil
}

// writeManifest records one "name digest" line per image, sorted by name.
func writeManifest(path string, images []Image) error {
	var b strings.Builder
	b.WriteString("# SHA-256 of the SVG each README image was rendered from.\n")
	b.WriteString("# Regenerate with: go run ./internal/gallery/cmd\n")
	for _, img := range images {
		fmt.Fprintf(&b, "%s %s\n", img.Name, img.Digest)
	}
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}
