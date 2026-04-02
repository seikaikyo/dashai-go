package fortune

import (
	"testing"
	"time"
)

func bd(s string) time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return t
}

func TestDetermineLevel(t *testing.T) {
	tests := []struct {
		name      string
		group     string
		direction string
		ryouhan   bool
		special   string
		wantLevel string
	}{
		// Base levels from relation group
		{"eishin base", "eishin", "栄", false, "", "daikichi"},
		{"yusui base", "yusui", "友", false, "", "kichi"},
		{"kisei base", "kisei", "危", false, "", "shokyo"},
		{"ankai base", "ankai", "壊", false, "", "kyo"},

		// Direction override
		{"sei override", "kisei", "成", false, "", "kichi"},
		{"an override", "ankai", "安", false, "", "shokyo"},

		// Ryouhan flip
		{"eishin+ryouhan", "eishin", "栄", true, "", "kyo"},
		{"ankai+ryouhan", "ankai", "壊", true, "", "daikichi"},
		{"yusui+ryouhan", "yusui", "友", true, "", "shokyo"},

		// Special day shift
		{"kanro upgrades", "yusui", "友", false, "kanro", "daikichi"},
		{"rasetsu downgrades", "yusui", "友", false, "rasetsu", "shokyo"},

		// Ryouhan + special day (reversed)
		{"ryouhan+kanro downgrades", "eishin", "栄", true, "kanro", "kyo"},   // kyo(flipped) stays kyo (already worst, can't go lower... wait)
		{"ryouhan+rasetsu capped", "ankai", "壊", true, "rasetsu", "daikichi"}, // kyo→flip→daikichi, rasetsu+ryouhan shift -1 capped at daikichi
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := DetermineLevel(tt.group, tt.direction, tt.ryouhan, tt.special)
			if got != tt.wantLevel {
				t.Errorf("got %q, want %q", got, tt.wantLevel)
			}
		})
	}
}

func TestDetermineLevelDetailed(t *testing.T) {
	// eishin/栄, no ryouhan, no special → daikichi
	lv, base := DetermineLevel("eishin", "栄", false, "")
	if lv != "daikichi" || base != "daikichi" {
		t.Errorf("eishin/栄: got %s/%s", lv, base)
	}

	// eishin/栄, ryouhan → kyo (flipped)
	lv, base = DetermineLevel("eishin", "栄", true, "")
	if lv != "kyo" || base != "daikichi" {
		t.Errorf("eishin/栄+ryouhan: got %s/%s", lv, base)
	}

	// kisei/成, override → kichi
	lv, base = DetermineLevel("kisei", "成", false, "")
	if lv != "kichi" || base != "kichi" {
		t.Errorf("kisei/成: got lv=%s base=%s", lv, base)
	}
}

func TestAggregateLevel(t *testing.T) {
	tests := []struct {
		levels []string
		want   string
	}{
		{[]string{"daikichi", "daikichi", "daikichi"}, "daikichi"},
		{[]string{"kyo", "kyo", "kyo"}, "kyo"},
		{[]string{"daikichi", "kyo"}, "kichi"},   // avg 1.5 → rounds to 2 → kichi
		{[]string{"kichi", "shokyo"}, "kichi"},    // avg 1.5 → rounds to 2 → kichi
		{[]string{}, "shokyo"},                     // empty default
	}

	for _, tt := range tests {
		got := AggregateLevel(tt.levels)
		if got != tt.want {
			t.Errorf("AggregateLevel(%v): got %q, want %q", tt.levels, got, tt.want)
		}
	}
}

func TestCalculateDaily(t *testing.T) {
	birth := bd("1977-01-15")
	target := bd("2026-04-02")

	result := CalculateDaily(birth, target, "zh-TW")

	if result.Date != "2026-04-02" {
		t.Errorf("date: got %q", result.Date)
	}
	if result.YourMansion.Index != 5 {
		t.Errorf("birth mansion: got %d, want 5", result.YourMansion.Index)
	}
	if result.Fortune.Level == "" {
		t.Error("level is empty")
	}
	if result.Fortune.LevelName == "" {
		t.Error("level_name is empty")
	}
	if result.Sanki.DayType == "" {
		t.Error("sanki day_type is empty")
	}
}

func TestCalculateWeekly(t *testing.T) {
	birth := bd("1977-01-15")
	target := bd("2026-04-02")

	result := CalculateWeekly(birth, target, "zh-TW")

	if len(result.Days) != 8 {
		t.Errorf("weekly days: got %d, want 8", len(result.Days))
	}
	// First day should be yesterday
	if result.Days[0].Date != "2026-04-01" {
		t.Errorf("first day: got %q, want 2026-04-01", result.Days[0].Date)
	}
}

func TestCalculateMonthly(t *testing.T) {
	birth := bd("1977-01-15")

	result := CalculateMonthly(birth, 2026, 4, "zh-TW")

	if result.Year != 2026 || result.Month != 4 {
		t.Errorf("year/month: got %d/%d", result.Year, result.Month)
	}
	if len(result.Days) != 30 {
		t.Errorf("april days: got %d, want 30", len(result.Days))
	}
	if result.Trend == "" {
		t.Error("trend is empty")
	}
}

func TestCalculateYearly(t *testing.T) {
	birth := bd("1977-01-15")

	result := CalculateYearly(birth, 2026, "zh-TW")

	if result.Year != 2026 {
		t.Errorf("year: got %d", result.Year)
	}
	if result.KuyouStar.Name == "" {
		t.Error("kuyou star name is empty")
	}
	if len(result.MonthlyTrend) != 12 {
		t.Errorf("monthly trends: got %d, want 12", len(result.MonthlyTrend))
	}
	if len(result.BestMonths) == 0 {
		t.Error("no best months found")
	}
}

func TestLevelName(t *testing.T) {
	if got := LevelName("daikichi", "zh-TW"); got != "大吉" {
		t.Errorf("got %q", got)
	}
	if got := LevelName("kyo", "en"); got != "Caution" {
		t.Errorf("got %q", got)
	}
}
