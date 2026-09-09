package gallery

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

// repoPath resolves a path relative to the repository root, which is two
// levels above this package.
func repoPath(parts ...string) string {
	return filepath.Join(append([]string{"..", ".."}, parts...)...)
}

func sourceDir() string { return repoPath("testdata", "golden") }
func targetDir() string { return repoPath("docs", "gallery") }

func TestGalleryIsUpToDate(t *testing.T) {
	Convey("Given the committed README images", t, func() {
		stale, err := Stale(sourceDir(), targetDir())

		Convey("When checking them against the renderer", func() {
			Convey("Then none was rendered from a different diagram", func() {
				So(err, ShouldBeNil)
				So(stale, ShouldBeEmpty)
			})
		})
	})

	Convey("Given the manifest", t, func() {
		recorded, err := ReadManifest(filepath.Join(targetDir(), ManifestName))

		Convey("When reading it", func() {
			Convey("Then it records one digest per gallery image", func() {
				So(err, ShouldBeNil)
				So(len(recorded), ShouldEqual, len(Names))
				for _, name := range Names {
					So(recorded[name], ShouldNotBeEmpty)
				}
			})
		})
	})

	Convey("Given every gallery image file", t, func() {
		Convey("When looking for it on disk", func() {
			Convey("Then it exists and holds PNG data", func() {
				for _, name := range Names {
					data, err := os.ReadFile(filepath.Join(targetDir(), name+".png"))
					So(err, ShouldBeNil)
					So(bytes.HasPrefix(data, []byte("\x89PNG")), ShouldBeTrue)
				}
			})
		})
	})
}

func TestRenderIsDeterministic(t *testing.T) {
	Convey("Given the same fixtures rendered twice", t, func() {
		first, err1 := Render(sourceDir())
		second, err2 := Render(sourceDir())

		Convey("When comparing the SVG digests", func() {
			Convey("Then they match, so drift always means a real change", func() {
				So(err1, ShouldBeNil)
				So(err2, ShouldBeNil)
				So(len(first), ShouldEqual, len(Names))
				for i := range first {
					So(first[i].Digest, ShouldEqual, second[i].Digest)
				}
			})
		})
	})
}

func TestGalleryErrors(t *testing.T) {
	Convey("Given a missing source directory", t, func() {
		_, err := Render(filepath.Join("does", "not", "exist"))

		Convey("Then it reports the failure instead of writing nothing quietly", func() {
			So(err, ShouldNotBeNil)
		})
	})

	Convey("Given a missing manifest", t, func() {
		recorded, err := ReadManifest(filepath.Join(t.TempDir(), "absent.txt"))

		Convey("Then it reads as empty rather than failing", func() {
			So(err, ShouldBeNil)
			So(recorded, ShouldBeEmpty)
		})
	})

	Convey("Given a target directory with no images", t, func() {
		stale, err := Stale(sourceDir(), t.TempDir())

		Convey("Then every image is reported stale", func() {
			So(err, ShouldBeNil)
			So(len(stale), ShouldEqual, len(Names))
		})
	})
}
