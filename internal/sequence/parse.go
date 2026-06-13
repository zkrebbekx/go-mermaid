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

// blockKeywords are recognized but not yet rendered; their lines are skipped
// so a diagram using them still parses (features arrive incrementally).
var blockKeywords = map[string]bool{
	"note": true, "loop": true, "alt": true, "opt": true, "else": true,
	"end": true, "activate": true, "deactivate": true, "autonumber": true,
	"par": true, "and": true, "rect": true, "critical": true, "break": true,
	"box": true,
}

// Parse builds a Diagram from sequence diagram source.
func Parse(src string) (*Diagram, error) {
	d := &Diagram{}
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

		switch kw := firstWord(line); {
		case kw == "participant" || kw == "actor":
			if err := d.parseParticipant(line, lineNo); err != nil {
				return nil, err
			}
		case blockKeywords[strings.ToLower(kw)]:
			// Recognized control/notation keyword: skip until supported.
			continue
		default:
			if err := d.parseMessage(line, lineNo); err != nil {
				return nil, err
			}
		}
	}

	if !headerSeen {
		return nil, syntax.Errorf(1, 1, "expected 'sequenceDiagram' header")
	}
	return d, nil
}

// parseParticipant handles "participant A" and "participant A as Alice".
func (d *Diagram) parseParticipant(line string, lineNo int) error {
	rest := strings.TrimSpace(line[len(firstWord(line)):])
	if rest == "" {
		return syntax.Errorf(lineNo, 1, "participant requires a name")
	}
	id, label := rest, rest
	if idx := strings.Index(rest, " as "); idx >= 0 {
		id = strings.TrimSpace(rest[:idx])
		label = strings.TrimSpace(rest[idx+4:])
	}
	d.ensureParticipant(id, label)
	return nil
}

// parseMessage handles "A->>B: text" and its arrow variants.
func (d *Diagram) parseMessage(line string, lineNo int) error {
	idx, tok, arrow := findArrow(line)
	if idx < 0 {
		return syntax.Errorf(lineNo, 1, "unrecognized statement %q", line)
	}
	from := strings.TrimSpace(line[:idx])
	rest := line[idx+len(tok):]

	to, text := rest, ""
	if c := strings.IndexByte(rest, ':'); c >= 0 {
		to = rest[:c]
		text = strings.TrimSpace(rest[c+1:])
	}
	to = strings.TrimSpace(to)
	// Activation suffixes (+/-) are accepted but not yet rendered.
	to = strings.TrimRight(strings.TrimSpace(strings.TrimLeft(to, "+-")), "+-")
	from = strings.TrimRight(from, "+-")

	if from == "" || to == "" {
		return syntax.Errorf(lineNo, 1, "message needs a sender and receiver")
	}
	d.ensureParticipant(from, from)
	d.ensureParticipant(to, to)
	d.Messages = append(d.Messages, &Message{From: from, To: to, Text: text, Arrow: arrow})
	return nil
}

// ensureParticipant adds a participant in first-seen order if absent.
func (d *Diagram) ensureParticipant(id, label string) {
	if d.participant(id) != nil {
		return
	}
	d.Participants = append(d.Participants, &Participant{ID: id, Label: label})
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
