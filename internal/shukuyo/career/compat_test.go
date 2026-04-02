package career

import (
	"fmt"
	"testing"
)

// TestBackwardCompatibility_Python verifies Go career results match Python API.
// Python API results (captured 2026-04-02):
//   User: birth=1990-05-15, mansion_index=10 (危宿)
//   Company A (1986-09-18): type=kisei, dir=危, tier=3, drain=moderate_high, p2_idx=14
//   Company B (2000-01-01): type=ankai, dir=安, tier=4, drain=moderate,      p2_idx=4
//   Company C (1975-06-15): type=kisei, dir=成, tier=3, drain=low,           p2_idx=24
func TestBackwardCompatibility_Python(t *testing.T) {
	cases := []struct {
		companyID      string
		companyName    string
		foundingDate   string
		wantGroup      string
		wantDirection  string
		wantTier       int
		wantDrain      string
		wantP2Index    int
	}{
		{"A", "Company A", "1986-09-18", "kisei", "危", 3, DrainModerateHigh, 14},
		{"B", "Company B", "2000-01-01", "ankai", "安", 4, DrainModerate, 4},
		{"C", "Company C", "1975-06-15", "kisei", "成", 3, DrainLow, 24},
	}

	for _, tc := range cases {
		t.Run(tc.companyID, func(t *testing.T) {
			resp, err := Analyze(AnalyzeRequest{
				BirthDate:   "1990-05-15",
				CompanyDate: tc.foundingDate,
				Context:     ContextEmployment,
				Lang:        "zh-TW",
			})
			if err != nil {
				t.Fatal(err)
			}

			// Person1 (user) mansion index
			if resp.Person1.Index != 10 {
				t.Errorf("user mansion index: got %d, want 10", resp.Person1.Index)
			}

			// Person2 (company) mansion index
			if resp.Person2.Index != tc.wantP2Index {
				t.Errorf("p2 index: got %d, want %d", resp.Person2.Index, tc.wantP2Index)
			}

			// Relation group
			if resp.Relation.Group != tc.wantGroup {
				t.Errorf("group: got %s, want %s", resp.Relation.Group, tc.wantGroup)
			}

			// Direction
			if resp.Relation.Direction != tc.wantDirection {
				t.Errorf("direction: got %s, want %s", resp.Relation.Direction, tc.wantDirection)
			}

			// Tier
			if resp.Level.Tier != tc.wantTier {
				t.Errorf("tier: got %d, want %d", resp.Level.Tier, tc.wantTier)
			}

			// Drain severity
			if resp.Drain.Severity != tc.wantDrain {
				t.Errorf("drain: got %s, want %s", resp.Drain.Severity, tc.wantDrain)
			}

			fmt.Printf("  %s: group=%s dir=%s tier=%d drain=%s p1=%d p2=%d  OK\n",
				tc.companyName, resp.Relation.Group, resp.Relation.Direction,
				resp.Level.Tier, resp.Drain.Severity, resp.Person1.Index, resp.Person2.Index)
		})
	}
}
