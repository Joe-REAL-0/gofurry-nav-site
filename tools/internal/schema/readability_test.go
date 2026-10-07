package schema

import (
	"reflect"
	"slices"
	"testing"
)

func readableSnapshot() Snapshot {
	return Snapshot{
		FormatVersion: 1,
		Tables: []Table{{Name: "sample", Comment: "中文业务表",
			Columns:  []Column{{Name: "appid", Comment: "Steam AppID，用于关联采集对象"}},
			Triggers: []Trigger{{Name: "audit", Comment: "更新业务记录时写入审计"}},
			Indexes:  []Index{{Name: "without_comment"}}, Constraints: []Constraint{{Name: "without_comment"}},
		}},
		Functions: []Function{{Identity: "project(p_day date)", Comment: "按 UTC 日期投影事实并写入结果"}},
		Sequences: []Sequence{{Name: "without_comment"}},
	}
}

func TestReadabilityMandatoryObjects(t *testing.T) {
	for _, kind := range []string{"table", "column", "function", "trigger"} {
		for _, comment := range []string{"", " \t\n\u3000", "Steam AppID", "UTC timestamp"} {
			t.Run(kind+"/"+comment, func(t *testing.T) {
				snapshot := readableSnapshot()
				var path string
				switch kind {
				case "table":
					snapshot.Tables[0].Comment, path = comment, "public.sample"
				case "column":
					snapshot.Tables[0].Columns[0].Comment, path = comment, "public.sample.appid"
				case "function":
					snapshot.Functions[0].Comment, path = comment, "public.project(p_day date)"
				case "trigger":
					snapshot.Tables[0].Triggers[0].Comment, path = comment, "public.sample.audit"
				}
				findings := ValidateReadability(snapshot)
				if len(findings) != 1 || findings[0].ObjectKind != kind || findings[0].ObjectName != path || findings[0].Reason == "" {
					t.Fatalf("unactionable finding: %#v", findings)
				}
			})
		}
	}
}

func TestReadabilityAllowsChineseTechnicalTermsAndOptionalObjects(t *testing.T) {
	snapshot := readableSnapshot()
	if findings := ValidateReadability(snapshot); len(findings) != 0 {
		t.Fatal(findings)
	}
	// Han includes supplementary ideographs, not only the basic CJK block.
	snapshot.Tables[0].Comment = "\U00020000 JSON"
	if findings := ValidateReadability(snapshot); len(findings) != 0 {
		t.Fatal(findings)
	}
}

func TestReadabilityFindingsAreDeterministic(t *testing.T) {
	snapshot := Snapshot{Tables: []Table{
		{Name: "z", Columns: []Column{{Name: "b"}, {Name: "a"}}, Triggers: []Trigger{{Name: "z"}, {Name: "a"}}},
		{Name: "a"},
	}, Functions: []Function{{Identity: "run(x text)"}, {Identity: "run(x integer)"}}}
	first := ValidateReadability(snapshot)
	slices.Reverse(snapshot.Tables[0].Columns)
	slices.Reverse(snapshot.Tables[0].Triggers)
	slices.Reverse(snapshot.Tables)
	slices.Reverse(snapshot.Functions)
	second := ValidateReadability(snapshot)
	if !reflect.DeepEqual(first, second) || len(first) != 8 {
		t.Fatalf("unstable findings: %#v / %#v", first, second)
	}
	if first[0].ObjectName != "public.z.a" || first[2].ObjectName != "public.run(x integer)" {
		t.Fatalf("unexpected object-path order: %#v", first)
	}
}

func TestSchemaDifferenceIncludesReadabilityComments(t *testing.T) {
	for _, kind := range []string{"table", "column", "function", "trigger"} {
		t.Run(kind, func(t *testing.T) {
			expected, actual := readableSnapshot(), readableSnapshot()
			switch kind {
			case "table":
				actual.Tables[0].Comment = "另一业务语义"
			case "column":
				actual.Tables[0].Columns[0].Comment = "另一字段语义"
			case "function":
				actual.Functions[0].Comment = "另一函数语义"
			case "trigger":
				actual.Tables[0].Triggers[0].Comment = "另一触发语义"
			}
			if Difference(expected, actual) == "" {
				t.Fatal("comment drift was ignored")
			}
		})
	}
}
