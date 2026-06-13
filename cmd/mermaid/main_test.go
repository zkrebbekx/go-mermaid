package main

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func resetFlags(args ...string) {
	flag.CommandLine = flag.NewFlagSet("mermaid", flag.ContinueOnError)
	os.Args = append([]string{"mermaid"}, args...)
}

func TestRunFileToFile(t *testing.T) {
	Convey("Given a diagram file and an output path", t, func() {
		dir := t.TempDir()
		in := filepath.Join(dir, "d.mmd")
		So(os.WriteFile(in, []byte("graph TD\nA --> B"), 0o644), ShouldBeNil)
		out := filepath.Join(dir, "d.svg")

		Convey("When running with -o", func() {
			resetFlags("-theme", "dark", "-o", out, in)
			err := run()

			Convey("Then it writes SVG to the output file", func() {
				So(err, ShouldBeNil)
				got, readErr := os.ReadFile(out)
				So(readErr, ShouldBeNil)
				So(string(got), ShouldStartWith, "<svg")
			})
		})
	})
}

func TestRunStdinToStdout(t *testing.T) {
	Convey("Given source on stdin", t, func() {
		dir := t.TempDir()
		stdin := filepath.Join(dir, "in")
		So(os.WriteFile(stdin, []byte("graph LR\nA --> B"), 0o644), ShouldBeNil)
		f, err := os.Open(stdin)
		So(err, ShouldBeNil)
		defer func() { _ = f.Close() }()

		oldIn, oldOut := os.Stdin, os.Stdout
		defer func() { os.Stdin, os.Stdout = oldIn, oldOut }()
		os.Stdin = f

		outFile := filepath.Join(dir, "captured")
		w, err := os.Create(outFile)
		So(err, ShouldBeNil)
		os.Stdout = w

		Convey("When running with '-'", func() {
			resetFlags("-")
			runErr := run()
			_ = w.Close()
			os.Stdout = oldOut

			Convey("Then it writes SVG to stdout", func() {
				So(runErr, ShouldBeNil)
				got, _ := os.ReadFile(outFile)
				So(string(got), ShouldContainSubstring, "<svg")
			})
		})
	})
}

func TestRunBadSource(t *testing.T) {
	Convey("Given a file with invalid diagram syntax", t, func() {
		dir := t.TempDir()
		in := filepath.Join(dir, "bad.mmd")
		So(os.WriteFile(in, []byte("graph TD\nA[oops"), 0o644), ShouldBeNil)

		Convey("When running", func() {
			resetFlags(in)
			err := run()

			Convey("Then it returns an error", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestRunBatch(t *testing.T) {
	Convey("Given multiple input files", t, func() {
		dir := t.TempDir()
		a := filepath.Join(dir, "a.mmd")
		bf := filepath.Join(dir, "b.mmd")
		So(os.WriteFile(a, []byte("graph TD\nA-->B"), 0o644), ShouldBeNil)
		So(os.WriteFile(bf, []byte("pie\n\"X\" : 1"), 0o644), ShouldBeNil)

		Convey("When running in batch mode", func() {
			resetFlags(a, bf)
			err := run()

			Convey("Then each input is rendered to a sibling .svg", func() {
				So(err, ShouldBeNil)
				_, ea := os.Stat(filepath.Join(dir, "a.svg"))
				_, eb := os.Stat(filepath.Join(dir, "b.svg"))
				So(ea, ShouldBeNil)
				So(eb, ShouldBeNil)
			})
		})

		Convey("When -o is combined with multiple files", func() {
			resetFlags("-o", filepath.Join(dir, "x.svg"), a, bf)
			err := run()

			Convey("Then it returns an error", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestRunVersion(t *testing.T) {
	Convey("Given the -version flag", t, func() {
		Convey("When running", func() {
			resetFlags("-version")
			err := run()

			Convey("Then it succeeds without needing input", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestRunMissingFile(t *testing.T) {
	Convey("Given a path that does not exist", t, func() {
		Convey("When running", func() {
			resetFlags(filepath.Join(t.TempDir(), "nope.mmd"))
			err := run()

			Convey("Then it returns an error", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}
