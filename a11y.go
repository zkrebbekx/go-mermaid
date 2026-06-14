package mermaid

import (
	"strings"

	"github.com/zkrebbekx/go-mermaid/internal/svgutil"
)

// extractA11y pulls single-line accTitle:/accDescr: directives out of the body
// and returns them with the cleaned body. These feed the SVG <title>/<desc>.
func extractA11y(body string) (accTitle, accDescr, cleaned string) {
	var kept []string
	for _, line := range strings.Split(body, "\n") {
		t := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(t, "accTitle:"):
			accTitle = strings.TrimSpace(t[len("accTitle:"):])
		case strings.HasPrefix(t, "accDescr:"):
			accDescr = strings.TrimSpace(t[len("accDescr:"):])
		default:
			kept = append(kept, line)
		}
	}
	return accTitle, accDescr, strings.Join(kept, "\n")
}

// injectA11y adds role="img" to the root <svg> tag and inserts <title>/<desc>
// elements for screen readers.
func injectA11y(svg []byte, title, desc string) []byte {
	s := string(svg)
	s = strings.Replace(s, "<svg ", `<svg role="img" `, 1)
	if title == "" && desc == "" {
		return []byte(s)
	}
	i := strings.IndexByte(s, '>')
	if i < 0 {
		return []byte(s)
	}
	var ins strings.Builder
	ins.WriteByte('\n')
	if title != "" {
		ins.WriteString("  <title>" + svgutil.Esc(title) + "</title>")
	}
	if desc != "" {
		if title != "" {
			ins.WriteByte('\n')
		}
		ins.WriteString("  <desc>" + svgutil.Esc(desc) + "</desc>")
	}
	return []byte(s[:i+1] + ins.String() + s[i+1:])
}
