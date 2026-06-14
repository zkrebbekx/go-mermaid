package render

import "github.com/zkrebbekx/go-mermaid/internal/svgutil"

// num and esc delegate to the shared svgutil helpers so all renderers format
// numbers and escape text identically.
func num(f float64) string { return svgutil.Num(f) }
func esc(s string) string  { return svgutil.Esc(s) }
