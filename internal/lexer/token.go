package lexer

// Kind enumerates token categories produced by the lexer.
type Kind int

const (
	// EOF marks the end of input.
	EOF Kind = iota
	// Newline separates statements.
	Newline
	// Keyword is a reserved word (graph, flowchart, subgraph, end).
	Keyword
	// Ident is a node identifier or bare word.
	Ident
	// Text is the contents of a shape or edge label.
	Text
	// Arrow is an edge connector (-->, ---, -.->, ==>).
	Arrow
	// ShapeOpen is an opening shape delimiter ([ ( ([ (( {).
	ShapeOpen
	// ShapeClose is a closing shape delimiter (] ) ]) )) }).
	ShapeClose
	// Pipe is the | bracketing an inline edge label.
	Pipe
	// Amp is the & joining several nodes on one side of a link.
	Amp
)

// Token is a lexical unit with its source position.
type Token struct {
	Kind Kind
	Val  string
	Line int // 1-based
	Col  int // 1-based, byte offset within the line
}
