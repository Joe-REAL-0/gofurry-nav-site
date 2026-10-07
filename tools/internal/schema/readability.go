package schema

import (
	"sort"
	"strings"
	"unicode"
)

// ReadabilityFinding identifies a current application object that needs a comment.
// ObjectName includes public schema, table context, or the full function signature.
type ReadabilityFinding struct {
	ObjectKind string
	ObjectName string
	Reason     string
}

// ValidateReadability enforces the mechanical part of contracts/database-schema.md.
// Inspect owns the application boundary; this check neither expands nor filters it.
// Semantic accuracy remains a review obligation, not a word-count heuristic.
func ValidateReadability(snapshot Snapshot) []ReadabilityFinding {
	var findings []ReadabilityFinding
	check := func(kind, name, comment string) {
		reason := ""
		if strings.TrimSpace(comment) == "" {
			reason = "缺少非空数据库注释"
		} else if !strings.ContainsFunc(comment, func(r rune) bool { return unicode.Is(unicode.Han, r) }) {
			reason = "数据库注释不包含中文语义（Unicode Han）"
		}
		if reason != "" {
			findings = append(findings, ReadabilityFinding{ObjectKind: kind, ObjectName: name, Reason: reason})
		}
	}
	for _, table := range snapshot.Tables {
		path := "public." + table.Name
		check("table", path, table.Comment)
		for _, column := range table.Columns {
			check("column", path+"."+column.Name, column.Comment)
		}
		for _, trigger := range table.Triggers {
			check("trigger", path+"."+trigger.Name, trigger.Comment)
		}
	}
	for _, function := range snapshot.Functions {
		check("function", "public."+function.Identity, function.Comment)
	}
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].ObjectKind != findings[j].ObjectKind {
			return findings[i].ObjectKind < findings[j].ObjectKind
		}
		if findings[i].ObjectName != findings[j].ObjectName {
			return findings[i].ObjectName < findings[j].ObjectName
		}
		return findings[i].Reason < findings[j].Reason
	})
	return findings
}
