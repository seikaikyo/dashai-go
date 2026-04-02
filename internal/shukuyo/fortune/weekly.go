package fortune

import (
	"time"

	"github.com/seikaikyo/dashai-go/internal/shukuyo/engine"
)

// WeeklyFortune is the weekly fortune response.
type WeeklyFortune struct {
	CenterDate    string            `json:"center_date"`
	WeekStart     string            `json:"week_start"`
	WeekEnd       string            `json:"week_end"`
	YourMansion   MansionBrief      `json:"your_mansion"`
	Fortune       FortuneDetail     `json:"fortune"`
	DailyOverview []DailyOverview   `json:"daily_overview"`
	Advice        string            `json:"advice,omitempty"`
	Focus         string            `json:"focus,omitempty"`
	CategoryTips  *CategoryTips     `json:"category_tips,omitempty"`
	Lucky         *LuckyInfo        `json:"lucky,omitempty"`
	WeekWarnings  []string          `json:"week_warnings,omitempty"`
	Days          []DailyFortune    `json:"days"`
}

// DailyOverview is a compact daily entry for weekly view.
type DailyOverview struct {
	Date       string `json:"date"`
	Weekday    string `json:"weekday"`
	Level      string `json:"level"`
	IsToday    bool   `json:"is_today"`
	SpecialDay string `json:"special_day,omitempty"`
	Ryouhan    bool   `json:"ryouhan_active"`
	IsDarkWeek bool   `json:"is_dark_week"`
}

// CategoryTips holds category-specific weekly tips.
type CategoryTips struct {
	Career string `json:"career,omitempty"`
	Love   string `json:"love,omitempty"`
	Health string `json:"health,omitempty"`
}

// CalculateWeekly computes an 8-day rolling window (yesterday + today + 6 future).
func CalculateWeekly(birthDate, targetDate time.Time, lang string) WeeklyFortune {
	lang = normalizeLang(lang)
	start := targetDate.AddDate(0, 0, -1) // yesterday
	today := targetDate.Format("2006-01-02")

	birthIdx, _ := engine.MansionIndexFromDate(birthDate)
	birthM := engine.Mansions27[birthIdx]

	days := make([]DailyFortune, 8)
	overviews := make([]DailyOverview, 8)
	levels := make([]string, 8)
	var warnings []string

	for i := 0; i < 8; i++ {
		d := start.AddDate(0, 0, i)
		days[i] = CalculateDaily(birthDate, d, lang)
		levels[i] = days[i].Fortune.Level

		isToday := d.Format("2006-01-02") == today
		sdName := ""
		if days[i].SpecialDay != nil {
			sdName = days[i].SpecialDay.Name
		}
		ryouhanActive := days[i].Ryouhan != nil && days[i].Ryouhan.Active
		isDark := days[i].Sanki.PeriodIndex == 1

		overviews[i] = DailyOverview{
			Date:       days[i].Date,
			Weekday:    days[i].Weekday.Name,
			Level:      days[i].Fortune.Level,
			IsToday:    isToday,
			SpecialDay: sdName,
			Ryouhan:    ryouhanActive,
			IsDarkWeek: isDark,
		}

		if days[i].Fortune.Level == LevelKyo && isToday {
			warnings = append(warnings, "today_kyo")
		}
		if ryouhanActive {
			warnings = append(warnings, "ryouhan_active")
		}
	}

	// Aggregate fortune
	aggLevel := AggregateLevel(levels)
	levelKey := levelToDescKey(aggLevel)
	fortunes := loadI18n(lang)
	fortuneData := loadFortuneI18n(lang)

	fortune := FortuneDetail{
		Level:      aggLevel,
		LevelName:  LevelName(aggLevel, lang),
		BaseLevel:  aggLevel,
		CareerDesc: getCategoryDesc(fortunes, "career", levelKey),
		LoveDesc:   getCategoryDesc(fortunes, "love", levelKey),
		HealthDesc: getCategoryDesc(fortunes, "health", levelKey),
		WealthDesc: getCategoryDesc(fortunes, "wealth", levelKey),
	}

	// Weekly focus
	focus := getWeeklyFocus(fortunes)

	// Lucky
	lucky := buildLuckyInfo(birthM.Yosei, fortuneData, targetDate)

	// Category tips
	tips := getWeeklyCategoryTips(fortunes)

	return WeeklyFortune{
		CenterDate: today,
		WeekStart:  start.Format("2006-01-02"),
		WeekEnd:    start.AddDate(0, 0, 7).Format("2006-01-02"),
		YourMansion: MansionBrief{
			Index: birthIdx, NameJP: birthM.NameJP,
			Reading: birthM.Reading, Yosei: birthM.Yosei,
		},
		Fortune:       fortune,
		DailyOverview: overviews,
		Advice:        focus,
		Focus:         focus,
		CategoryTips:  tips,
		Lucky:         lucky,
		WeekWarnings:  warnings,
		Days:          days,
	}
}

func getWeeklyFocus(fortunes map[string]any) string {
	focus, ok := fortunes["weekly_fortune_focus"].(map[string]any)
	if !ok {
		return ""
	}
	if neutral, ok := focus["neutral"].(string); ok {
		return neutral
	}
	return ""
}

func getWeeklyCategoryTips(fortunes map[string]any) *CategoryTips {
	tips, ok := fortunes["weekly_category_tips"].(map[string]any)
	if !ok {
		return nil
	}
	result := &CategoryTips{}
	if c, ok := tips["career"].(string); ok {
		result.Career = c
	}
	if l, ok := tips["love"].(string); ok {
		result.Love = l
	}
	if h, ok := tips["health"].(string); ok {
		result.Health = h
	}
	if result.Career == "" && result.Love == "" && result.Health == "" {
		return nil
	}
	return result
}
