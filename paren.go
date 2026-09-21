package dnslint

import (
	"fmt"
	"strings"
)

// JoinParenthesized collapses a zone file's parenthesized rdata groups into
// single logical lines that ParseRecord can handle. SOA records are the
// usual reason this exists: their rdata is long enough that zone files
// wrap it across several lines in parens, e.g.
//
//	@ IN SOA ns1.example.com. hostmaster.example.com. (
//	    2024010100 ; serial
//	    3600       ; refresh
//	    900        ; retry
//	    604800     ; expire
//	    86400 )    ; minimum
//
// Each source line is comment-stripped before joining, because a ';'
// comment only extends to the end of its own line, not to the end of the
// parenthesized group. Lines outside any parens pass through unchanged
// (aside from comment stripping and trimming). Callers run this over the
// raw lines of a zone file before handing each result to ParseRecord.
func JoinParenthesized(lines []string) ([]string, error) {
	var joined []string
	var current strings.Builder
	open := false
	openedAt := 0

	for i, line := range lines {
		stripped := stripComment(line)

		if !open {
			idx := strings.IndexByte(stripped, '(')
			if idx < 0 {
				if trimmed := strings.TrimSpace(stripped); trimmed != "" {
					joined = append(joined, trimmed)
				}
				continue
			}
			open = true
			openedAt = i + 1
			current.Reset()
			current.WriteString(stripped[:idx])
			stripped = stripped[idx+1:]
		}

		if closeIdx := strings.IndexByte(stripped, ')'); closeIdx >= 0 {
			current.WriteByte(' ')
			current.WriteString(stripped[:closeIdx])
			joined = append(joined, strings.Join(strings.Fields(current.String()), " "))
			open = false
			continue
		}

		current.WriteByte(' ')
		current.WriteString(stripped)
	}

	if open {
		return nil, fmt.Errorf("dnslint: unclosed parenthesis starting at line %d", openedAt)
	}

	return joined, nil
}
