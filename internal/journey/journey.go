// Package journey parses and renders Mermaid user-journey diagrams to SVG.
//
// Syntax:
//
//	journey
//	    title My day
//	    section Work
//	      Make tea: 5: Me
//	      Do work: 1: Me, Cat
package journey

import (
	"strconv"
	"strings"

	"github.com/zkrebbekx/go-mermaid/internal/syntax"
)

// Task is a single journey step with a satisfaction score (1-5) and actors.
type Task struct {
	Name   string
	Score  int
	Actors []string
}

// Section groups consecutive tasks.
type Section struct {
	Name  string
	Tasks []*Task
}

// Diagram is a parsed user-journey diagram.
type Diagram struct {
	Title    string
	Sections []*Section
}

// Tasks returns all tasks across sections in order.
func (d *Diagram) Tasks() []*Task {
	var all []*Task
	for _, s := range d.Sections {
		all = append(all, s.Tasks...)
	}
	return all
}

// Parse builds a Diagram from journey source.
func Parse(src string) (*Diagram, error) {
	d := &Diagram{}
	var cur *Section

	headerSeen := false
	for i, raw := range strings.Split(src, "\n") {
		lineNo := i + 1
		line := strings.TrimSpace(stripComment(raw))
		if line == "" {
			continue
		}
		if !headerSeen {
			if firstWord(line) != "journey" {
				return nil, syntax.Errorf(lineNo, 1, "expected 'journey' header")
			}
			headerSeen = true
			continue
		}
		switch {
		case strings.HasPrefix(line, "title "):
			d.Title = strings.TrimSpace(line[len("title "):])
		case strings.HasPrefix(line, "section "):
			cur = &Section{Name: strings.TrimSpace(line[len("section "):])}
			d.Sections = append(d.Sections, cur)
		default:
			t, err := parseTask(line, lineNo)
			if err != nil {
				return nil, err
			}
			if cur == nil {
				cur = &Section{}
				d.Sections = append(d.Sections, cur)
			}
			cur.Tasks = append(cur.Tasks, t)
		}
	}
	if !headerSeen {
		return nil, syntax.Errorf(1, 1, "expected 'journey' header")
	}
	return d, nil
}

func parseTask(line string, lineNo int) (*Task, error) {
	parts := strings.SplitN(line, ":", 3)
	if len(parts) < 2 {
		return nil, syntax.Errorf(lineNo, 1, "expected 'name: score[: actors]'")
	}
	score, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return nil, syntax.Errorf(lineNo, 1, "invalid task score")
	}
	t := &Task{Name: strings.TrimSpace(parts[0]), Score: score}
	if len(parts) == 3 {
		for _, a := range strings.Split(parts[2], ",") {
			if a = strings.TrimSpace(a); a != "" {
				t.Actors = append(t.Actors, a)
			}
		}
	}
	return t, nil
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
