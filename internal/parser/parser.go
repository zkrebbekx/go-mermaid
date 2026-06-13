// Package parser turns a token stream into a domain.Graph. It targets the
// flowchart grammar subset: a header (graph/flowchart + direction) followed
// by node and edge statements.
package parser

import (
	"strings"

	"github.com/Zac300/go-mermaid/internal/domain"
	"github.com/Zac300/go-mermaid/internal/lexer"
	"github.com/Zac300/go-mermaid/internal/syntax"
)

// Parse builds a Graph from tokens produced by the lexer.
func Parse(toks []lexer.Token) (*domain.Graph, error) {
	p := &parser{toks: toks}
	return p.parse()
}

type parser struct {
	toks  []lexer.Token
	pos   int
	graph *domain.Graph
	seen  map[string]*domain.Node
}

func (p *parser) parse() (*domain.Graph, error) {
	p.graph = &domain.Graph{Direction: domain.TopBottom}
	p.seen = map[string]*domain.Node{}

	p.skipNewlines()
	if err := p.parseHeader(); err != nil {
		return nil, err
	}

	for !p.at(lexer.EOF) {
		p.skipNewlines()
		if p.at(lexer.EOF) {
			break
		}
		if err := p.parseStatement(); err != nil {
			return nil, err
		}
	}
	return p.graph, nil
}

func (p *parser) parseHeader() error {
	t := p.cur()
	if t.Kind != lexer.Keyword || (t.Val != "graph" && t.Val != "flowchart") {
		return p.errAt(t, "expected 'graph' or 'flowchart'")
	}
	p.next()
	if dir := p.cur(); dir.Kind == lexer.Ident {
		d, ok := direction(dir.Val)
		if !ok {
			return p.errAt(dir, "unknown direction %q", dir.Val)
		}
		p.graph.Direction = d
		p.next()
	}
	if !p.at(lexer.Newline) && !p.at(lexer.EOF) {
		return p.errAt(p.cur(), "expected end of header line")
	}
	return nil
}

// parseStatement parses a chain like A[x] --> B -->|label| C.
func (p *parser) parseStatement() error {
	from, err := p.parseNodeRef()
	if err != nil {
		return err
	}
	for p.at(lexer.Arrow) {
		arrowTok := p.cur()
		p.next()

		label := ""
		if p.at(lexer.Pipe) {
			p.next()
			if !p.at(lexer.Text) {
				return p.errAt(p.cur(), "expected edge label text")
			}
			label = p.cur().Val
			p.next()
			if !p.at(lexer.Pipe) {
				return p.errAt(p.cur(), "expected closing '|'")
			}
			p.next()
		} else if !strings.Contains(arrowTok.Val, ">") {
			// Possible middle-form label: A -- text --> B / A == t ==> B.
			// The first connector is open (no arrowhead); words between it and
			// a second connector are the label, and that second connector
			// determines the arrow style.
			saved := p.pos
			var words []string
			for p.at(lexer.Ident) || p.at(lexer.Text) {
				words = append(words, p.cur().Val)
				p.next()
			}
			if len(words) > 0 && p.at(lexer.Arrow) {
				label = strings.Join(words, " ")
				arrowTok = p.cur()
				p.next()
			} else {
				p.pos = saved // not middle-form; treat as a normal link
			}
		}

		to, err := p.parseNodeRef()
		if err != nil {
			return err
		}
		p.graph.Edges = append(p.graph.Edges, &domain.Edge{
			From:  from.ID,
			To:    to.ID,
			Label: label,
			Arrow: arrowKind(arrowTok.Val),
		})
		from = to
	}
	return nil
}

// parseNodeRef parses an identifier with an optional shape+label, registering
// the node (or merging the label into a previously seen node).
func (p *parser) parseNodeRef() (*domain.Node, error) {
	idTok := p.cur()
	if idTok.Kind != lexer.Ident {
		return nil, p.errAt(idTok, "expected node identifier")
	}
	p.next()

	node := p.seen[idTok.Val]
	if node == nil {
		node = &domain.Node{ID: idTok.Val, Label: idTok.Val, Shape: domain.ShapeRect}
		p.seen[idTok.Val] = node
		p.graph.Nodes = append(p.graph.Nodes, node)
	}

	if p.at(lexer.ShapeOpen) {
		open := p.cur().Val
		p.next()
		if !p.at(lexer.Text) {
			return nil, p.errAt(p.cur(), "expected shape label")
		}
		node.Label = p.cur().Val
		node.Shape = shapeKind(open)
		p.next()
		if !p.at(lexer.ShapeClose) {
			return nil, p.errAt(p.cur(), "expected closing shape delimiter")
		}
		p.next()
	}
	return node, nil
}

// --- token cursor helpers ---

func (p *parser) cur() lexer.Token     { return p.toks[p.pos] }
func (p *parser) at(k lexer.Kind) bool { return p.toks[p.pos].Kind == k }

func (p *parser) next() {
	if p.pos < len(p.toks)-1 {
		p.pos++
	}
}

func (p *parser) skipNewlines() {
	for p.at(lexer.Newline) {
		p.next()
	}
}

func (p *parser) errAt(t lexer.Token, format string, args ...any) error {
	return syntax.Errorf(t.Line, t.Col, format, args...)
}

// --- mappings ---

func direction(s string) (domain.Direction, bool) {
	switch strings.ToUpper(s) {
	case "TD", "TB":
		return domain.TopBottom, true
	case "BT":
		return domain.BottomTop, true
	case "LR":
		return domain.LeftRight, true
	case "RL":
		return domain.RightLeft, true
	}
	return "", false
}

func shapeKind(open string) domain.Shape {
	switch open {
	case "(":
		return domain.ShapeRound
	case "([":
		return domain.ShapeStadium
	case "((":
		return domain.ShapeCircle
	case "{":
		return domain.ShapeDiamond
	default:
		return domain.ShapeRect
	}
}

func arrowKind(s string) domain.Arrow {
	switch {
	case strings.Contains(s, "."):
		return domain.ArrowDotted
	case strings.Contains(s, "="):
		return domain.ArrowThick
	case !strings.Contains(s, ">"):
		return domain.ArrowOpen
	default:
		return domain.ArrowNormal
	}
}
