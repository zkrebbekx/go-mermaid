// Package git parses and renders Mermaid gitGraph diagrams to SVG.
//
// Syntax:
//
//	gitGraph
//	   commit
//	   branch develop
//	   commit id: "x"
//	   checkout main
//	   merge develop tag: "v1"
package git

import (
	"strings"

	"github.com/zkrebbekx/go-mermaid/internal/syntax"
)

// Commit is a point on a branch lane. Order is its global x position.
type Commit struct {
	Order  int
	Branch string
	ID     string
	Tag    string
	Merge  bool
}

// Branch is a lane in the graph.
type Branch struct {
	Name string
	Lane int
}

// Merge connects a source branch's tip to a merge commit.
type Merge struct {
	FromOrder int
	ToOrder   int
}

// Diagram is a parsed gitGraph.
type Diagram struct {
	Branches []*Branch
	Commits  []*Commit
	Merges   []*Merge
}

// Branch returns the lane index for a branch name (−1 if absent).
func (d *Diagram) lane(name string) int {
	for _, b := range d.Branches {
		if b.Name == name {
			return b.Lane
		}
	}
	return -1
}

// Parse builds a Diagram from gitGraph source.
func Parse(src string) (*Diagram, error) {
	d := &Diagram{}
	d.Branches = append(d.Branches, &Branch{Name: "main", Lane: 0})
	current := "main"
	order := 0
	lastOrder := map[string]int{} // branch -> tip commit order

	headerSeen := false
	for i, raw := range strings.Split(src, "\n") {
		lineNo := i + 1
		line := strings.TrimSpace(stripComment(raw))
		if line == "" {
			continue
		}
		if !headerSeen {
			if w := strings.TrimRight(strings.ToLower(firstWord(line)), ":"); w != "gitgraph" {
				return nil, syntax.Errorf(lineNo, 1, "expected 'gitGraph' header")
			}
			headerSeen = true
			continue
		}

		switch kw := firstWord(line); kw {
		case "commit":
			c := &Commit{Order: order, Branch: current}
			c.ID, c.Tag = commitMeta(line)
			d.Commits = append(d.Commits, c)
			lastOrder[current] = order
			order++
		case "branch":
			name := strings.TrimSpace(line[len("branch"):])
			if name != "" && d.lane(name) < 0 {
				d.Branches = append(d.Branches, &Branch{Name: name, Lane: len(d.Branches)})
				current = name
			}
		case "checkout", "switch":
			name := strings.TrimSpace(line[len(kw):])
			if d.lane(name) >= 0 {
				current = name
			}
		case "merge":
			rest := strings.TrimSpace(line[len("merge"):])
			from := firstWord(rest)
			c := &Commit{Order: order, Branch: current, Merge: true}
			_, c.Tag = commitMeta(line)
			d.Commits = append(d.Commits, c)
			if src, ok := lastOrder[from]; ok {
				d.Merges = append(d.Merges, &Merge{FromOrder: src, ToOrder: order})
			}
			lastOrder[current] = order
			order++
		case "cherry-pick":
			c := &Commit{Order: order, Branch: current}
			d.Commits = append(d.Commits, c)
			lastOrder[current] = order
			order++
		default:
			return nil, syntax.Errorf(lineNo, 1, "unrecognized statement %q", line)
		}
	}
	if !headerSeen {
		return nil, syntax.Errorf(1, 1, "expected 'gitGraph' header")
	}
	return d, nil
}

// commitMeta extracts id: "..." and tag: "..." from a commit/merge line.
func commitMeta(line string) (id, tag string) {
	return quotedAfter(line, "id:"), quotedAfter(line, "tag:")
}

func quotedAfter(line, key string) string {
	i := strings.Index(line, key)
	if i < 0 {
		return ""
	}
	rest := line[i+len(key):]
	if q := strings.IndexByte(rest, '"'); q >= 0 {
		rest = rest[q+1:]
		if e := strings.IndexByte(rest, '"'); e >= 0 {
			return rest[:e]
		}
	}
	return ""
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
