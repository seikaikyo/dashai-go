package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"
)

func mustParseDate(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return t
}

// TestMansionIndex10 verifies 10 hand-picked mansion indices with lunar date check.
func TestMansionIndex10(t *testing.T) {
	tests := []struct {
		solar        string
		lunarMonth   int
		lunarDay     int
		mansionIndex int
	}{
		{"1977-01-15", 11, 26, 5},
		{"1990-05-20", 4, 26, 15},
		{"2000-01-01", 11, 25, 4},
		{"1985-12-25", 11, 14, 20},
		{"1995-07-04", 6, 7, 0},
		{"2003-03-15", 2, 13, 25},
		{"1988-08-08", 6, 26, 19},
		{"1970-06-30", 5, 27, 18},
		{"2010-10-10", 9, 3, 4},
		{"1965-02-14", 1, 13, 23},
	}

	for _, tt := range tests {
		t.Run(tt.solar, func(t *testing.T) {
			bd := mustParseDate(tt.solar)
			ld := SolarToLunar(bd)

			if ld.Month != tt.lunarMonth {
				t.Errorf("lunar month: got %d, want %d", ld.Month, tt.lunarMonth)
			}
			if ld.Day != tt.lunarDay {
				t.Errorf("lunar day: got %d, want %d", ld.Day, tt.lunarDay)
			}

			idx, _ := MansionIndexFromDate(bd)
			if idx != tt.mansionIndex {
				t.Errorf("mansion index: got %d, want %d", idx, tt.mansionIndex)
			}
		})
	}
}

// TestMansionIndex100 verifies 100 random birth dates against Python golden data.
func TestMansionIndex100(t *testing.T) {
	data, err := os.ReadFile("testdata/golden_100_mansions.json")
	if err != nil {
		t.Skipf("testdata not found: %v", err)
	}

	type entry struct {
		Solar      string `json:"s"`
		LunarMonth int    `json:"lm"`
		LunarDay   int    `json:"ld"`
		Index      int    `json:"i"`
	}
	var golden []entry
	if err := json.Unmarshal(data, &golden); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(golden) != 100 {
		t.Fatalf("expected 100, got %d", len(golden))
	}

	for _, g := range golden {
		bd := mustParseDate(g.Solar)
		ld := SolarToLunar(bd)
		idx := MansionIndex(ld.Month, ld.Day)

		if ld.Month != g.LunarMonth || ld.Day != g.LunarDay {
			t.Errorf("%s: lunar got %d/%d, want %d/%d", g.Solar, ld.Month, ld.Day, g.LunarMonth, g.LunarDay)
		}
		if idx != g.Index {
			t.Errorf("%s: mansion got %d, want %d", g.Solar, idx, g.Index)
		}
	}
}

// TestRelation729 verifies all 27x27 = 729 relation combinations.
// Pure original text: distance → position name → group → inverse.
func TestRelation729(t *testing.T) {
	data, err := os.ReadFile("testdata/golden_729.json")
	if err != nil {
		t.Skipf("testdata not found: %v", err)
	}

	type entry struct {
		Idx1      int    `json:"i1"`
		Idx2      int    `json:"i2"`
		Fwd       int    `json:"f"`
		Group     string `json:"g"`
		Direction string `json:"d"`
		Inverse   string `json:"inv"`
	}
	var golden []entry
	if err := json.Unmarshal(data, &golden); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(golden) != 729 {
		t.Fatalf("expected 729, got %d", len(golden))
	}

	for _, g := range golden {
		rel := GetRelation(g.Idx1, g.Idx2)

		if rel.ForwardDistance != g.Fwd {
			t.Errorf("(%d,%d) fwd: got %d, want %d", g.Idx1, g.Idx2, rel.ForwardDistance, g.Fwd)
		}
		if rel.Group != g.Group {
			t.Errorf("(%d,%d) group: got %q, want %q", g.Idx1, g.Idx2, rel.Group, g.Group)
		}
		if rel.Direction != g.Direction {
			t.Errorf("(%d,%d) dir: got %q, want %q", g.Idx1, g.Idx2, rel.Direction, g.Direction)
		}
		if rel.InverseDirection != g.Inverse {
			t.Errorf("(%d,%d) inv: got %q, want %q", g.Idx1, g.Idx2, rel.InverseDirection, g.Inverse)
		}
	}
}

// TestKuyou270 verifies 270 kuyou star calculations.
func TestKuyou270(t *testing.T) {
	data, err := os.ReadFile("testdata/golden_270_kuyou.json")
	if err != nil {
		t.Skipf("testdata not found: %v", err)
	}

	type entry struct {
		BirthYear int `json:"by"`
		Year      int `json:"y"`
		KazoeAge  int `json:"ka"`
		StarIndex int `json:"si"`
	}
	var golden []entry
	if err := json.Unmarshal(data, &golden); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(golden) != 270 {
		t.Fatalf("expected 270, got %d", len(golden))
	}

	for _, g := range golden {
		bd := time.Date(g.BirthYear, 6, 15, 0, 0, 0, 0, time.UTC)
		star := GetKuyouStar(bd, g.Year)

		if star.KazoeAge != g.KazoeAge {
			t.Errorf("birth=%d year=%d: kazoe got %d, want %d", g.BirthYear, g.Year, star.KazoeAge, g.KazoeAge)
		}
		if star.Index != g.StarIndex {
			t.Errorf("birth=%d year=%d: star got %d, want %d", g.BirthYear, g.Year, star.Index, g.StarIndex)
		}
	}
}

// TestKuyouStarNames verifies one full 9-year cycle.
func TestKuyouStarNames(t *testing.T) {
	bd := mustParseDate("1977-01-15")
	expected := []string{"月曜星", "木曜星", "羅喉星", "土曜星", "水曜星", "金曜星", "日曜星", "火曜星", "計都星"}
	for i, name := range expected {
		star := GetKuyouStar(bd, 2020+i)
		if star.Name != name {
			t.Errorf("year %d: got %q, want %q", 2020+i, star.Name, name)
		}
	}
}

// TestSpecialDay verifies all 21 special day entries from 卷五.
func TestSpecialDay(t *testing.T) {
	tests := []struct {
		weekday, mansion int
		wantType         string
	}{
		// kanro (7 entries)
		{0, 26, "kanro"}, {1, 17, "kanro"}, {2, 5, "kanro"},
		{3, 22, "kanro"}, {4, 21, "kanro"}, {5, 3, "kanro"}, {6, 23, "kanro"},
		// kongou (7 entries)
		{0, 5, "kongou"}, {1, 4, "kongou"}, {2, 12, "kongou"},
		{3, 16, "kongou"}, {4, 20, "kongou"}, {5, 24, "kongou"}, {6, 1, "kongou"},
		// rasetsu (7 entries)
		{0, 15, "rasetsu"}, {1, 21, "rasetsu"}, {2, 25, "rasetsu"},
		{3, 19, "rasetsu"}, {4, 2, "rasetsu"}, {5, 13, "rasetsu"}, {6, 22, "rasetsu"},
		// no special day
		{0, 0, ""}, {3, 3, ""},
	}

	for _, tt := range tests {
		key := specialDayKey{tt.weekday, tt.mansion}
		got, ok := SpecialDayMap[key]
		if tt.wantType == "" {
			if ok {
				t.Errorf("wd=%d m=%d: want none, got %q", tt.weekday, tt.mansion, got)
			}
		} else if !ok || got != tt.wantType {
			t.Errorf("wd=%d m=%d: got %q, want %q", tt.weekday, tt.mansion, got, tt.wantType)
		}
	}
}

// TestSankiPosition verifies three-nine cycle positions.
func TestSankiPosition(t *testing.T) {
	tests := []struct {
		birthIdx, dayIdx, periodIdx, dayInPeriod int
		dayType                                  string
	}{
		// Full 一九 cycle from mansion 0
		{0, 0, 0, 1, "命の日"}, {0, 1, 0, 2, "栄の日"}, {0, 2, 0, 3, "衰の日"},
		{0, 3, 0, 4, "安の日"}, {0, 4, 0, 5, "危の日"}, {0, 5, 0, 6, "成の日"},
		{0, 6, 0, 7, "壊の日"}, {0, 7, 0, 8, "友の日"}, {0, 8, 0, 9, "親の日"},
		// Period boundaries
		{0, 9, 1, 1, "業の日"}, {0, 10, 1, 2, "栄の日"}, {0, 17, 1, 9, "親の日"},
		{0, 18, 2, 1, "胎の日"}, {0, 19, 2, 2, "栄の日"}, {0, 26, 2, 9, "親の日"},
		// Different birth mansions
		{5, 5, 0, 1, "命の日"}, {5, 14, 1, 1, "業の日"}, {5, 23, 2, 1, "胎の日"},
		{26, 25, 2, 9, "親の日"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("b%d_d%d", tt.birthIdx, tt.dayIdx), func(t *testing.T) {
			s := GetSankiPosition(tt.birthIdx, tt.dayIdx)
			if s.PeriodIndex != tt.periodIdx {
				t.Errorf("period: got %d, want %d", s.PeriodIndex, tt.periodIdx)
			}
			if s.DayInPeriod != tt.dayInPeriod {
				t.Errorf("dayInPeriod: got %d, want %d", s.DayInPeriod, tt.dayInPeriod)
			}
			if s.DayType != tt.dayType {
				t.Errorf("dayType: got %q, want %q", s.DayType, tt.dayType)
			}
		})
	}
}

// TestSankuPositionNamesCompleteness verifies all 27 positions map to a group.
func TestSankuPositionNamesCompleteness(t *testing.T) {
	for d := 0; d < 27; d++ {
		pos := SankuPositionNames[d]
		if pos == "" {
			t.Errorf("distance %d: empty position name", d)
		}
		if _, ok := positionToGroup[pos]; !ok {
			t.Errorf("distance %d: position %q has no group", d, pos)
		}
		if _, ok := DirectionInverse[pos]; !ok {
			t.Errorf("distance %d: position %q has no inverse", d, pos)
		}
	}
}

// TestCompatibility verifies full compatibility with known values.
func TestCompatibility(t *testing.T) {
	d1 := mustParseDate("1977-01-15") // mansion 5
	d2 := mustParseDate("1990-05-20") // mansion 15

	result := Compatibility(d1, d2)

	if result.Person1.Index != 5 {
		t.Errorf("p1 index: got %d, want 5", result.Person1.Index)
	}
	if result.Person2.Index != 15 {
		t.Errorf("p2 index: got %d, want 15", result.Person2.Index)
	}
	// fwd = (15-5)%27 = 10, SANKU_POS[10] = "栄", group = "eishin"
	if result.Relation.ForwardDistance != 10 {
		t.Errorf("fwd: got %d, want 10", result.Relation.ForwardDistance)
	}
	if result.Relation.Direction != "栄" {
		t.Errorf("dir: got %q, want 栄", result.Relation.Direction)
	}
	if result.Relation.Group != "eishin" {
		t.Errorf("group: got %q, want eishin", result.Relation.Group)
	}
	if result.Relation.InverseDirection != "親" {
		t.Errorf("inv: got %q, want 親", result.Relation.InverseDirection)
	}
}

// TestRyouhanMap verifies all 12 months have entries.
func TestRyouhanMap(t *testing.T) {
	covered := make(map[int]bool)
	for k := range RyouhanMap {
		covered[k.month] = true
	}
	for m := 1; m <= 12; m++ {
		if !covered[m] {
			t.Errorf("month %d: no ryouhan entry", m)
		}
	}
}
