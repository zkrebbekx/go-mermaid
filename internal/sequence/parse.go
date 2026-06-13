package sequence

import (
	"strings"

	"github.com/Zac300/go-mermaid/internal/syntax"
)

// arrowTokens are message operators, longest first so the longest match at a
// position wins (e.g. "-->>" before "->>").
var arrowTokens = []struct {
	tok   string
	arrow Arrow
}{
	{"-->>", Arrow{Dashed: true, Head: HeadArrow}},
	{"-->", Arrow{Dashed: true, Head: HeadNone}},
	{"--x", Arrow{Dashed: true, Head: HeadCross}},
	{"--)", Arrow{Dashed: true, Head: HeadArrow}},
	{"->>", Arrow{Dashed: false, Head: HeadArrow}},
	{"->", Arrow{Dashed: false, Head: HeadNone}},
	{"-x", Arrow{Dashed: false, Head: HeadCross}},
	{"-)", Arrow{Dashed: false, Head: HeadArrow}},
}

// frameKeywords open a grouping frame.
var frameKeywords = map[string]bool{
	"loop": true, "alt": true, "opt": true, "par": true,
	"rect": true, "critical": true, "break": true,
}

// blockKeywords are recognized but not drawn (skipped) keywords.
var blockKeywords = map[string]bool{
	"box": true,
}

type parser struct {
	d       *Diagram
	act     map[string][]int // activation start-row stack per participant
	frames  []*Frame         // open frame stack
	autonum bool             // autonumber active
	msgNum  int              // running autonumber counter
}

// Parse builds a Diagram from sequence diagram source.
func Parse(src string) (*Diagram, error) {
	p := &parser{d: &Diagram{}, act: map[string][]int{}}
	lines := strings.Split(src, "\n")

	headerSeen := false
	for i, raw := range lines {
		lineNo := i + 1
		line := strings.TrimSpace(stripComment(raw))
		if line == "" {
			continue
		}
		if !headerSeen {
			if firstWord(line) != "sequenceDiagram" {
				return nil, syntax.Errorf(lineNo, 1, "expected 'sequenceDiagram' header")
			}
			headerSeen = true
			continue
		}

		kw := strings.ToLower(firstWord(line))
		switch {
		case kw == "participant" || kw == "actor":
			if err := p.parseParticipant(line, lineNo); err != nil {
				return nil, err
			}
		case kw == "note":
			if err := p.parseNote(line, lineNo); err != nil {
				return nil, err
			}
		case kw == "autonumber":
			p.autonum = true
		case kw == "activate":
			p.activate(strings.TrimSpace(line[len("activate"):]), p.d.rows)
		case kw == "deactivate":
			p.deactivate(strings.TrimSpace(line[len("deactivate"):]), p.d.rows)
		case frameKeywords[kw]:
			f := &Frame{Type: kw, Label: strings.TrimSpace(line[len(kw):]), StartRow: p.d.rows}
			p.d.Frames = append(p.d.Frames, f)
			p.frames = append(p.frames, f)
		case kw == "else" || kw == "and":
			if n := len(p.frames); n > 0 {
				top := p.frames[n-1]
				top.Sections = append(top.Sections, &Section{Row: p.d.rows, Label: strings.TrimSpace(line[len(kw):])})
			}
		case kw == "end":
			if n := len(p.frames); n > 0 {
				top := p.frames[n-1]
				top.EndRow = p.d.rows - 1
				p.frames = p.frames[:n-1]
			}
		case blockKeywords[kw]:
			continue
		default:
			if err := p.parseMessage(line, lineNo); err != nil {
				return nil, err
			}
		}
	}

	if !headerSeen {
		return nil, syntax.Errorf(1, 1, "expected 'sequenceDiagram' header")
	}
	p.closeOpenBars()
	return p.d, nil
}

func (p *parser) nextRow() int {
	r := p.d.rows
	p.d.rows++
	return r
}

func (p *parser) parseParticipant(line string, lineNo int) error {
	rest := strings.TrimSpace(line[len(firstWord(line)):])
	if rest == "" {
		return syntax.Errorf(lineNo, 1, "participant requires a name")
	}
	id, label := rest, rest
	if idx := strings.Index(rest, " as "); idx >= 0 {
		id = strings.TrimSpace(rest[:idx])
		label = strings.TrimSpace(rest[idx+4:])
	}
	p.ensureParticipant(id, label)
	return nil
}

func (p *parser) parseMessage(line string, lineNo int) error {
	idx, tok, arrow := findArrow(line)
	if idx < 0 {
		return syntax.Errorf(lineNo, 1, "unrecognized statement %q", line)
	}
	from := strings.TrimSpace(line[:idx])
	rest := line[idx+len(tok):]

	toRaw, text := rest, ""
	if c := strings.IndexByte(rest, ':'); c >= 0 {
		toRaw = rest[:c]
		text = strings.TrimSpace(rest[c+1:])
	}
	toRaw = strings.TrimSpace(toRaw)

	activateTarget, deactivateSender := false, false
	switch {
	case strings.HasPrefix(toRaw, "+"):
		activateTarget, toRaw = true, strings.TrimSpace(toRaw[1:])
	case strings.HasPrefix(toRaw, "-"):
		deactivateSender, toRaw = true, strings.TrimSpace(toRaw[1:])
	}
	to := toRaw

	if from == "" || to == "" {
		return syntax.Errorf(lineNo, 1, "message needs a sender and receiver")
	}
	p.ensureParticipant(from, from)
	p.ensureParticipant(to, to)

	row := p.nextRow()
	num := 0
	if p.autonum {
		p.msgNum++
		num = p.msgNum
	}
	p.d.Messages = append(p.d.Messages, &Message{From: from, To: to, Text: text, Arrow: arrow, Row: row, Num: num})
	if activateTarget {
		p.activate(to, row)
	}
	if deactivateSender {
		p.deactivate(from, row)
	}
	return nil
}

func (p *parser) parseNote(line string, lineNo int) error {
	rest := strings.TrimSpace(line[len("note"):])
	var pos NotePos
	switch {
	case strings.HasPrefix(rest, "right of "):
		pos, rest = NoteRight, rest[len("right of "):]
	case strings.HasPrefix(rest, "left of "):
		pos, rest = NoteLeft, rest[len("left of "):]
	case strings.HasPrefix(rest, "over "):
		pos, rest = NoteOver, rest[len("over "):]
	default:
		return syntax.Errorf(lineNo, 1, "note needs 'right of', 'left of', or 'over'")
	}
	who, text := rest, ""
	if c := strings.IndexByte(rest, ':'); c >= 0 {
		who = rest[:c]
		text = strings.TrimSpace(rest[c+1:])
	}
	var of []string
	for _, w := range strings.Split(who, ",") {
		if w = strings.TrimSpace(w); w != "" {
			of = append(of, w)
			p.ensureParticipant(w, w)
		}
	}
	if len(of) == 0 {
		return syntax.Errorf(lineNo, 1, "note needs a participant")
	}
	p.d.Notes = append(p.d.Notes, &Note{Pos: pos, Of: of, Text: text, Row: p.nextRow()})
	return nil
}

func (p *parser) activate(name string, row int) {
	name = strings.TrimSpace(name)
	if name == "" {
		return
	}
	p.act[name] = append(p.act[name], row)
}

func (p *parser) deactivate(name string, row int) {
	name = strings.TrimSpace(name)
	st := p.act[name]
	if len(st) == 0 {
		return
	}
	start := st[len(st)-1]
	p.act[name] = st[:len(st)-1]
	p.d.Bars = append(p.d.Bars, &Bar{Participant: name, StartRow: start, EndRow: row})
}

// closeOpenBars ends any activations left open at the last row.
func (p *parser) closeOpenBars() {
	end := p.d.rows - 1
	if end < 0 {
		end = 0
	}
	for name, st := range p.act {
		for _, start := range st {
			p.d.Bars = append(p.d.Bars, &Bar{Participant: name, StartRow: start, EndRow: end})
		}
	}
}

func (p *parser) ensureParticipant(id, label string) {
	if p.d.participant(id) != nil {
		return
	}
	p.d.Participants = append(p.d.Participants, &Participant{ID: id, Label: label})
}

// findArrow returns the earliest arrow operator in s (longest match wins).
func findArrow(s string) (idx int, tok string, arrow Arrow) {
	for i := 0; i < len(s); i++ {
		for _, at := range arrowTokens {
			if strings.HasPrefix(s[i:], at.tok) {
				return i, at.tok, at.arrow
			}
		}
	}
	return -1, "", Arrow{}
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
