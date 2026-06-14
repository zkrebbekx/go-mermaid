package parser

import "github.com/zkrebbekx/go-mermaid/internal/syntax"

// ParseError reports a syntax error with its source position. It is an
// alias for syntax.Error so the lexer, parser, and public package all
// share one positional error type.
type ParseError = syntax.Error
