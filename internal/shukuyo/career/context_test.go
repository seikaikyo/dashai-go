package career

import (
	"testing"
)

// TestContextDifferentiation verifies that different contexts return
// different descriptions for the same birth_date + company_date pair.
func TestContextDifferentiation(t *testing.T) {
	contexts := []ContextType{
		ContextEmployment, ContextConsulting, ContextOutsourcing,
		ContextB2B, ContextHR, ContextHeadhunter,
	}

	descs := make(map[ContextType]string)
	for _, ctx := range contexts {
		resp, err := Analyze(AnalyzeRequest{
			BirthDate:   "1990-05-15",
			CompanyDate: "1986-09-18",
			Context:     ctx,
			Lang:        "zh-TW",
		})
		if err != nil {
			t.Fatalf("context %s: %v", ctx, err)
		}
		if resp.Relation.Description == "" {
			t.Errorf("context %s: empty description", ctx)
		}
		descs[ctx] = resp.Relation.Description

		// Relation group and drain should be identical across contexts
		if resp.Relation.Group != "kisei" {
			t.Errorf("context %s: group=%s, want kisei", ctx, resp.Relation.Group)
		}
		if resp.Drain.Severity != DrainModerateHigh {
			t.Errorf("context %s: drain=%s, want moderate_high", ctx, resp.Drain.Severity)
		}
	}

	// Employment and consulting should have different descriptions
	if descs[ContextEmployment] == descs[ContextConsulting] {
		t.Error("employment and consulting have identical descriptions")
	}
	if descs[ContextEmployment] == descs[ContextB2B] {
		t.Error("employment and b2b have identical descriptions")
	}
	if descs[ContextConsulting] == descs[ContextOutsourcing] {
		t.Error("consulting and outsourcing have identical descriptions")
	}
	if descs[ContextHR] == descs[ContextHeadhunter] {
		t.Error("hr and headhunter have identical descriptions")
	}
}

// TestAllContextsLoadTexts verifies all 6 contexts can load their JSON data.
func TestAllContextsLoadTexts(t *testing.T) {
	contexts := []ContextType{
		ContextEmployment, ContextConsulting, ContextOutsourcing,
		ContextB2B, ContextHR, ContextHeadhunter,
	}
	langs := []string{"zh-TW", "ja", "en"}

	for _, ctx := range contexts {
		for _, lang := range langs {
			rels := LoadRelations(ctx, lang)
			if len(rels) != 6 {
				t.Errorf("%s/%s relations: got %d groups, want 6", ctx, lang, len(rels))
			}
			careers := LoadCareer(ctx, lang)
			if len(careers) != 11 {
				t.Errorf("%s/%s career: got %d directions, want 11", ctx, lang, len(careers))
			}
			drains := LoadDrainTexts(ctx, lang)
			if len(drains) != 11 {
				t.Errorf("%s/%s drain: got %d directions, want 11", ctx, lang, len(drains))
			}
			flags := LoadRedFlags(ctx, lang)
			if len(flags) == 0 {
				t.Errorf("%s/%s red_flags: empty", ctx, lang)
			}
		}
	}
}
