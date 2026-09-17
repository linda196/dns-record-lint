// Package dnslint parses BIND-style zone file records and checks them for
// common mistakes: CNAMEs sharing a name with other records, TTLs that were
// never set explicitly, and so on.
package dnslint

import (
	"fmt"
	"strconv"
	"strings"
)

// knownTypes lists the record types ParseRecord understands well enough to
// validate with type-specific rules. Anything else still parses fine, it
// just won't be checked beyond the generic rules.
var knownTypes = map[string]bool{
	"A":     true,
	"AAAA":  true,
	"CNAME": true,
	"MX":    true,
	"NS":    true,
	"TXT":   true,
	"PTR":   true,
	"SOA":   true,
	"SRV":   true,
	"CAA":   true,
}

var knownClasses = map[string]bool{
	"IN": true,
	"CH": true,
	"HS": true,
}

// Record is a single resource record as it would appear in a zone file
// line: NAME [TTL] [CLASS] TYPE RDATA...
type Record struct {
	Name   string
	TTL    int
	TTLSet bool
	Class  string
	Type   string
	Value  string
}

// ParseRecord parses one zone-file line. It expects the caller to have
// already filtered out blank lines, $DIRECTIVE lines, and multi-line
// records (parenthesized SOA blocks and the like) — those need a stateful
// reader, which this package doesn't provide yet.
func ParseRecord(line string) (Record, error) {
	raw := strings.TrimSpace(stripComment(line))
	if raw == "" {
		return Record{}, fmt.Errorf("dnslint: empty record line")
	}

	fields := strings.Fields(raw)
	if len(fields) < 2 {
		return Record{}, fmt.Errorf("dnslint: too few fields: %q", line)
	}

	rec := Record{Name: fields[0]}
	i := 1

	if ttl, err := strconv.Atoi(fields[i]); err == nil {
		rec.TTL = ttl
		rec.TTLSet = true
		i++
	}

	if i < len(fields) && knownClasses[strings.ToUpper(fields[i])] {
		rec.Class = strings.ToUpper(fields[i])
		i++
	}

	if i >= len(fields) {
		return Record{}, fmt.Errorf("dnslint: missing record type: %q", line)
	}
	rec.Type = strings.ToUpper(fields[i])
	i++

	if i >= len(fields) {
		return Record{}, fmt.Errorf("dnslint: missing record data: %q", line)
	}
	rec.Value = strings.Join(fields[i:], " ")

	return rec, nil
}

// stripComment removes a trailing ';' comment. It doesn't understand quoted
// strings, so a semicolon inside a quoted TXT value gets cut too — a known
// limitation rather than something worth complicating this for right now.
func stripComment(line string) string {
	if idx := strings.IndexByte(line, ';'); idx >= 0 {
		return line[:idx]
	}
	return line
}

// IsKnownType reports whether t (case-insensitive) is a record type this
// package has rules for.
func IsKnownType(t string) bool {
	return knownTypes[strings.ToUpper(t)]
}
