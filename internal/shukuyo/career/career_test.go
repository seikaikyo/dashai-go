package career

import (
	"testing"
)

// Known test pair: 1990-05-15 (birth) vs 1986-09-18 (company founding)
// Python API returns eishin for this pair.
const (
	testBirth   = "1990-05-15"
	testCompany = "1986-09-18"
)

func TestAnalyze_Employment(t *testing.T) {
	resp, err := Analyze(AnalyzeRequest{
		BirthDate:   testBirth,
		CompanyDate: testCompany,
		Context:     ContextEmployment,
		Lang:        "zh-TW",
	})
	if err != nil {
		t.Fatal(err)
	}

	// Verify basic structure
	if resp.Person1.Index < 0 || resp.Person1.Index > 26 {
		t.Errorf("person1 index out of range: %d", resp.Person1.Index)
	}
	if resp.Person2.Index < 0 || resp.Person2.Index > 26 {
		t.Errorf("person2 index out of range: %d", resp.Person2.Index)
	}
	if resp.Relation.Group == "" {
		t.Error("relation group is empty")
	}
	if resp.Level.Tier < 1 || resp.Level.Tier > 4 {
		t.Errorf("tier out of range: %d", resp.Level.Tier)
	}
	if resp.Drain.Severity == "" {
		t.Error("drain severity is empty")
	}
	if resp.Relation.Description == "" {
		t.Error("relation description is empty (i18n loading failed?)")
	}
	if resp.Career == nil {
		t.Error("career advice is nil")
	}
}

func TestAnalyze_LanguageSwitching(t *testing.T) {
	langs := []string{"zh-TW", "ja", "en"}
	descs := make([]string, 3)

	for i, lang := range langs {
		resp, err := Analyze(AnalyzeRequest{
			BirthDate:   testBirth,
			CompanyDate: testCompany,
			Context:     ContextEmployment,
			Lang:        lang,
		})
		if err != nil {
			t.Fatal(err)
		}
		descs[i] = resp.Relation.Description
		if descs[i] == "" {
			t.Errorf("empty description for lang=%s", lang)
		}
	}

	// zh-TW and ja should differ
	if descs[0] == descs[1] {
		t.Error("zh-TW and ja descriptions are identical")
	}
	// zh-TW and en should differ
	if descs[0] == descs[2] {
		t.Error("zh-TW and en descriptions are identical")
	}
}

func TestDrainLevel_MatchesPython(t *testing.T) {
	cases := []struct {
		group, direction string
		want             string
	}{
		{"mei", "命", DrainNone},
		{"gyotai", "業", DrainLow},
		{"gyotai", "胎", DrainMinimal},
		{"eishin", "栄", DrainLow},
		{"eishin", "親", DrainMinimal},
		{"yusui", "友", DrainLow},
		{"yusui", "衰", DrainHigh},
		{"ankai", "安", DrainModerate},
		{"ankai", "壊", DrainVeryHigh},
		{"kisei", "危", DrainModerateHigh},
		{"kisei", "成", DrainLow},
	}

	for _, tc := range cases {
		got := DrainLevel(tc.group, tc.direction)
		if got != tc.want {
			t.Errorf("DrainLevel(%s, %s) = %s, want %s", tc.group, tc.direction, got, tc.want)
		}
	}
}

func TestCareerLevel_MatchesPython(t *testing.T) {
	cases := []struct {
		group string
		want  string
	}{
		{"eishin", LevelDaikichi},
		{"gyotai", LevelKichi},
		{"mei", LevelKichi},
		{"yusui", LevelShokyo},
		{"kisei", LevelShokyo},
		{"ankai", LevelKyo},
	}

	for _, tc := range cases {
		got := CareerLevel(tc.group)
		if got != tc.want {
			t.Errorf("CareerLevel(%s) = %s, want %s", tc.group, got, tc.want)
		}
	}
}

func TestBatchAnalyze(t *testing.T) {
	companies := []CompanyEntry{
		{ID: "1", Name: "Company A", FoundingDate: "1986-09-18"},
		{ID: "2", Name: "Company B", FoundingDate: "2000-01-01"},
		{ID: "3", Name: "Company C", FoundingDate: "1975-06-15"},
	}

	resp, err := BatchAnalyze(BatchRequest{
		BirthDate: testBirth,
		Year:      2026,
		Companies: companies,
		Context:   ContextEmployment,
		Lang:      "zh-TW",
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(resp.Companies) != 3 {
		t.Errorf("expected 3 companies, got %d", len(resp.Companies))
	}
	if resp.User.Kuyou == nil {
		t.Error("user kuyou is nil")
	}
	if len(resp.TierSummary) == 0 {
		t.Error("tier summary is empty")
	}
}

func TestInterviewDates(t *testing.T) {
	resp, err := InterviewDates(InterviewDatesRequest{
		BirthDate:   testBirth,
		CompanyDate: testCompany,
		Days:        30,
		Lang:        "zh-TW",
	})
	if err != nil {
		t.Fatal(err)
	}

	if resp.Target.Date == "" {
		t.Error("target date is empty")
	}
	if len(resp.Alternatives) == 0 {
		t.Error("no alternatives found in 30 days")
	}
	// Alternatives should be sorted by verdict
	for _, alt := range resp.Alternatives {
		if alt.Verdict == "warning" {
			t.Error("warning verdict should be filtered out")
		}
	}
}

func TestTeamMatrix(t *testing.T) {
	candidates := []CandidateEntry{
		{ID: "1", Name: "Alice", BirthDate: "1990-05-15"},
		{ID: "2", Name: "Bob", BirthDate: "1988-03-22"},
		{ID: "3", Name: "Carol", BirthDate: "1992-11-08"},
	}

	resp, err := TeamMatrix(TeamMatrixRequest{
		CompanyDate: testCompany,
		Candidates:  candidates,
		Lang:        "zh-TW",
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(resp.Matrix) != 3 {
		t.Errorf("expected 3x3 matrix, got %dx?", len(resp.Matrix))
	}
	if len(resp.Matrix[0]) != 3 {
		t.Errorf("expected 3x3 matrix, got ?x%d", len(resp.Matrix[0]))
	}
	// Diagonal should be "self"
	for i := 0; i < 3; i++ {
		if resp.Matrix[i][i].Group != "self" {
			t.Errorf("diagonal [%d][%d] should be 'self', got %s", i, i, resp.Matrix[i][i].Group)
		}
	}
	if len(resp.Rankings) != 3 {
		t.Errorf("expected 3 rankings, got %d", len(resp.Rankings))
	}
}

func BenchmarkBatch20(b *testing.B) {
	companies := make([]CompanyEntry, 20)
	for i := range companies {
		companies[i] = CompanyEntry{
			ID:           string(rune('A' + i)),
			Name:         "Company " + string(rune('A'+i)),
			FoundingDate: "1990-01-01",
		}
	}
	req := BatchRequest{
		BirthDate: testBirth,
		Year:      2026,
		Companies: companies,
		Context:   ContextEmployment,
		Lang:      "zh-TW",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BatchAnalyze(req)
	}
}

func BenchmarkTeamMatrix50(b *testing.B) {
	candidates := make([]CandidateEntry, 50)
	for i := range candidates {
		// Spread birth dates across a range
		day := 1 + (i % 28)
		month := 1 + (i % 12)
		candidates[i] = CandidateEntry{
			ID:        string(rune('A' + i%26)),
			Name:      "Candidate " + string(rune('A'+i%26)),
			BirthDate: "1990-" + pad(month) + "-" + pad(day),
		}
	}
	req := TeamMatrixRequest{
		CompanyDate: testCompany,
		Candidates:  candidates,
		Lang:        "zh-TW",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TeamMatrix(req)
	}
}

func pad(n int) string {
	if n < 10 {
		return "0" + string(rune('0'+n))
	}
	return string(rune('0'+n/10)) + string(rune('0'+n%10))
}
