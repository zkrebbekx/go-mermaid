// Package gantt parses and renders Mermaid gantt charts to SVG.
//
// Syntax (subset):
//
//	gantt
//	    title Project
//	    dateFormat YYYY-MM-DD
//	    section Phase 1
//	      Design : d1, 2024-01-01, 10d
//	      Build  : after d1, 2w
package gantt

import (
	"strconv"
	"strings"
	"time"

	"github.com/zkrebbekx/go-mermaid/internal/syntax"
)

// Task is a single bar with a resolved start date and duration in days.
type Task struct {
	ID      string
	Name    string
	Section string
	AfterID string
	StartIn string // raw start date token (empty if "after")
	Days    int

	Start time.Time
}

// End returns the day after the task's last day.
func (t *Task) End() time.Time { return t.Start.AddDate(0, 0, t.Days) }

// Diagram is a parsed gantt chart.
type Diagram struct {
	Title    string
	Layout   string // Go time layout derived from dateFormat
	Sections []string
	Tasks    []*Task
}

// Parse builds a Diagram from gantt source and resolves task start dates.
func Parse(src string) (*Diagram, error) {
	d := &Diagram{Layout: "2006-01-02"}
	section := ""

	headerSeen := false
	for i, raw := range strings.Split(src, "\n") {
		lineNo := i + 1
		line := strings.TrimSpace(stripComment(raw))
		if line == "" {
			continue
		}
		if !headerSeen {
			if firstWord(line) != "gantt" {
				return nil, syntax.Errorf(lineNo, 1, "expected 'gantt' header")
			}
			headerSeen = true
			continue
		}
		switch {
		case strings.HasPrefix(line, "title "):
			d.Title = strings.TrimSpace(line[len("title "):])
		case strings.HasPrefix(line, "dateFormat "):
			d.Layout = toGoLayout(strings.TrimSpace(line[len("dateFormat "):]))
		case strings.HasPrefix(line, "excludes "), strings.HasPrefix(line, "axisFormat "):
			// recognized but ignored
		case strings.HasPrefix(line, "section "):
			section = strings.TrimSpace(line[len("section "):])
			d.Sections = append(d.Sections, section)
		default:
			t, err := parseTask(line, section, lineNo)
			if err != nil {
				return nil, err
			}
			d.Tasks = append(d.Tasks, t)
		}
	}
	if !headerSeen {
		return nil, syntax.Errorf(1, 1, "expected 'gantt' header")
	}
	if err := d.resolve(); err != nil {
		return nil, err
	}
	return d, nil
}

func parseTask(line, section string, lineNo int) (*Task, error) {
	name, rest, ok := strings.Cut(line, ":")
	if !ok {
		return nil, syntax.Errorf(lineNo, 1, "expected 'name: spec'")
	}
	t := &Task{Name: strings.TrimSpace(name), Section: section}
	for _, f := range strings.Split(rest, ",") {
		f = strings.TrimSpace(f)
		switch {
		case f == "":
		case strings.HasPrefix(f, "after "):
			t.AfterID = strings.TrimSpace(f[len("after "):])
		case isDuration(f):
			t.Days = parseDuration(f)
		case looksLikeDate(f):
			t.StartIn = f
		default:
			t.ID = f // bare token is the task id
		}
	}
	if t.Days == 0 {
		return nil, syntax.Errorf(lineNo, 1, "task %q has no duration", t.Name)
	}
	return t, nil
}

// resolve computes each task's Start from its date or "after" dependency.
func (d *Diagram) resolve() error {
	byID := map[string]*Task{}
	for _, t := range d.Tasks {
		switch {
		case t.AfterID != "":
			if dep, ok := byID[t.AfterID]; ok {
				t.Start = dep.End()
			}
		case t.StartIn != "":
			if v, err := time.Parse(d.Layout, t.StartIn); err == nil {
				t.Start = v
			}
		}
		if t.ID != "" {
			byID[t.ID] = t
		}
	}
	return nil
}

// Bounds returns the earliest start and latest end across tasks.
func (d *Diagram) Bounds() (lo, hi time.Time) {
	first := true
	for _, t := range d.Tasks {
		if t.Start.IsZero() {
			continue
		}
		if first {
			lo, hi = t.Start, t.End()
			first = false
			continue
		}
		if t.Start.Before(lo) {
			lo = t.Start
		}
		if t.End().After(hi) {
			hi = t.End()
		}
	}
	return lo, hi
}

func isDuration(f string) bool {
	if len(f) < 2 {
		return false
	}
	unit := f[len(f)-1]
	if unit != 'd' && unit != 'w' && unit != 'h' {
		return false
	}
	_, err := strconv.Atoi(f[:len(f)-1])
	return err == nil
}

func parseDuration(f string) int {
	n, _ := strconv.Atoi(f[:len(f)-1])
	switch f[len(f)-1] {
	case 'w':
		return n * 7
	case 'h':
		if n < 24 {
			return 1
		}
		return n / 24
	default:
		return n
	}
}

func looksLikeDate(f string) bool {
	return strings.Count(f, "-") >= 2 || strings.Count(f, "/") >= 2
}

// toGoLayout converts a Mermaid dateFormat to a Go time layout.
func toGoLayout(format string) string {
	r := strings.NewReplacer("YYYY", "2006", "YY", "06", "MM", "01", "DD", "02", "HH", "15", "mm", "04", "ss", "05")
	return r.Replace(format)
}

func firstWord(s string) string {
	if i := strings.IndexAny(s, " \t"); i >= 0 {
		return s[:i]
	}
	return s
}

func stripComment(s string) string {
	if i := strings.Index(s, "%%"); i >= 0 {
		return s[:i]
	}
	return s
}
