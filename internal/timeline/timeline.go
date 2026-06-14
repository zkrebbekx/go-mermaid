// Package timeline parses and renders Mermaid timeline diagrams to SVG.
//
// Syntax:
//
//	timeline
//	    title History
//	    section Early
//	      2002 : LinkedIn
//	      2004 : Facebook : Google
package timeline

import (
	"strings"

	"github.com/zkrebbekx/go-mermaid/internal/syntax"
)

// Period is a point in time with one or more events.
type Period struct {
	Time   string
	Events []string
}

// Section groups consecutive periods.
type Section struct {
	Name    string
	Periods []*Period
}

// Diagram is a parsed timeline.
type Diagram struct {
	Title    string
	Sections []*Section
}

// Periods returns all periods across sections in order.
func (d *Diagram) Periods() []*Period {
	var all []*Period
	for _, s := range d.Sections {
		all = append(all, s.Periods...)
	}
	return all
}

// Parse builds a Diagram from timeline source.
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
			if firstWord(line) != "timeline" {
				return nil, syntax.Errorf(lineNo, 1, "expected 'timeline' header")
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
			parts := strings.Split(line, ":")
			if len(parts) < 2 {
				return nil, syntax.Errorf(lineNo, 1, "expected 'time : event'")
			}
			p := &Period{Time: strings.TrimSpace(parts[0])}
			for _, ev := range parts[1:] {
				if ev = strings.TrimSpace(ev); ev != "" {
					p.Events = append(p.Events, ev)
				}
			}
			if cur == nil {
				cur = &Section{}
				d.Sections = append(d.Sections, cur)
			}
			cur.Periods = append(cur.Periods, p)
		}
	}
	if !headerSeen {
		return nil, syntax.Errorf(1, 1, "expected 'timeline' header")
	}
	return d, nil
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
