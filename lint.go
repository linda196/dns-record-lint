package dnslint

import (
	"sort"
	"strings"
)

// Rule inspects a full record set and returns any issues it finds. Rules
// see the whole set rather than one record at a time because some checks,
// like the CNAME conflict below, are inherently cross-record.
type Rule func(records []Record) []Issue

// Lint runs every rule against records and returns a combined, sorted
// Report. Sorting by name then type keeps output stable across runs, which
// matters once callers start diffing JSON reports over time.
func Lint(records []Record, rules ...Rule) Report {
	var issues []Issue
	for _, rule := range rules {
		issues = append(issues, rule(records)...)
	}
	sort.SliceStable(issues, func(i, j int) bool {
		if issues[i].Name != issues[j].Name {
			return issues[i].Name < issues[j].Name
		}
		return issues[i].Type < issues[j].Type
	})
	return Report{Issues: issues}
}

// DefaultRules is a reasonable starting set of checks. Callers can pass a
// subset, or their own rules entirely, directly to Lint instead.
func DefaultRules() []Rule {
	return []Rule{
		CheckMissingTTL,
		CheckCNAMEConflict,
	}
}

// CheckMissingTTL flags records with no explicit TTL. Inheriting a TTL from
// $TTL or the SOA minimum isn't wrong, but it's easy to lose track of which
// records do, so this is informational rather than a warning.
func CheckMissingTTL(records []Record) []Issue {
	var issues []Issue
	for _, rec := range records {
		if !rec.TTLSet {
			issues = append(issues, Issue{
				Record:   rec,
				Name:     rec.Name,
				Type:     rec.Type,
				Severity: SeverityInfo,
				Message:  "no explicit TTL; value comes from zone or SOA default",
			})
		}
	}
	return issues
}

// CheckCNAMEConflict flags names that have a CNAME alongside any other
// record type. RFC 1034 section 3.6.2 forbids this: a CNAME must be the
// only record at its name, and resolvers disagree on what to do otherwise.
func CheckCNAMEConflict(records []Record) []Issue {
	byName := make(map[string][]Record)
	for _, rec := range records {
		key := normalizeName(rec.Name)
		byName[key] = append(byName[key], rec)
	}

	var issues []Issue
	for _, group := range byName {
		if len(group) < 2 {
			continue
		}
		hasCNAME := false
		for _, rec := range group {
			if rec.Type == "CNAME" {
				hasCNAME = true
				break
			}
		}
		if !hasCNAME {
			continue
		}
		for _, rec := range group {
			issues = append(issues, Issue{
				Record:   rec,
				Name:     rec.Name,
				Type:     rec.Type,
				Severity: SeverityError,
				Message:  "name has a CNAME plus other records, which RFC 1034 forbids",
			})
		}
	}
	return issues
}

func normalizeName(name string) string {
	n := strings.ToLower(name)
	return strings.TrimSuffix(n, ".")
}
