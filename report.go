package dnslint

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Severity classifies how serious a lint Issue is.
type Severity string

const (
	SeverityInfo    Severity = "info"
	SeverityWarning Severity = "warning"
	SeverityError   Severity = "error"
)

// Issue is a single finding produced by a Rule.
type Issue struct {
	Record   Record   `json:"-"`
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	Severity Severity `json:"severity"`
	Message  string   `json:"message"`
}

// Report is the result of running one or more Rules over a record set.
type Report struct {
	Issues []Issue `json:"issues"`
}

// Text renders the report the way a human reads a terminal: one line per
// issue, or a short confirmation if nothing was found.
func (r Report) Text() string {
	if len(r.Issues) == 0 {
		return "no issues found"
	}
	var b strings.Builder
	for _, issue := range r.Issues {
		fmt.Fprintf(&b, "[%s] %s %s: %s\n", strings.ToUpper(string(issue.Severity)), issue.Name, issue.Type, issue.Message)
	}
	return strings.TrimRight(b.String(), "\n")
}

// JSON renders the report as indented JSON — the machine-readable
// counterpart to Text, meant for scripts and CI checks.
func (r Report) JSON() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}

// Format is a convenience wrapper for callers (typically a CLI with a
// --json flag) that pick between the two output modes at runtime.
func (r Report) Format(asJSON bool) (string, error) {
	if !asJSON {
		return r.Text(), nil
	}
	data, err := r.JSON()
	if err != nil {
		return "", err
	}
	return string(data), nil
}
