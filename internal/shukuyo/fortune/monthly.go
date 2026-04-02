package fortune

import (
	"time"

	"github.com/seikaikyo/dashai-go/internal/shukuyo/engine"
)

// MonthlyFortune is the monthly fortune response.
type MonthlyFortune struct {
	Year         int                 `json:"year"`
	Month        int                 `json:"month"`
	YourMansion  MansionBrief        `json:"your_mansion"`
	MonthMansion *MansionBrief       `json:"month_mansion,omitempty"`
	Relation     *MonthRelation      `json:"relation,omitempty"`
	Theme        *ThemeBlock         `json:"theme,omitempty"`
	Fortune      FortuneDetail       `json:"fortune"`
	Weekly       []WeeklySummary     `json:"weekly"`
	SpecialDays  []MonthSpecialDay   `json:"special_days,omitempty"`
	RyouhanInfo  *RyouhanStat        `json:"ryouhan_info,omitempty"`
	Advice       string              `json:"advice,omitempty"`
	Days         []DailyFortune      `json:"days"`
	Trend        string              `json:"trend"`
	TrendName    string              `json:"trend_name"`
}

// MonthRelation holds the month-level relation.
type MonthRelation struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Reading     string `json:"reading"`
	Description string `json:"description,omitempty"`
}

// ThemeBlock holds a theme description.
type ThemeBlock struct {
	Description string `json:"description,omitempty"`
}

// WeeklySummary holds a weekly aggregation within a month.
type WeeklySummary struct {
	Week       int             `json:"week"`
	WeekStart  string          `json:"week_start"`
	WeekEnd    string          `json:"week_end"`
	DaysCount  int             `json:"days_count"`
	Level      string          `json:"level"`
	HasDarkWeek bool           `json:"has_dark_week"`
	DailyOverview []DailyOverview `json:"daily_overview"`
}

// MonthSpecialDay is a special day in the month.
type MonthSpecialDay struct {
	Date string `json:"date"`
	Type string `json:"type"`
	Name string `json:"name"`
}

// RyouhanStat holds ryouhan period statistics.
type RyouhanStat struct {
	AffectedDays int     `json:"affected_days"`
	TotalDays    int     `json:"total_days"`
	Ratio        float64 `json:"ratio"`
}

// CalculateMonthly computes daily fortune for every day of a month.
func CalculateMonthly(birthDate time.Time, year, month int, lang string) MonthlyFortune {
	lang = normalizeLang(lang)

	birthIdx, _ := engine.MansionIndexFromDate(birthDate)
	birthM := engine.Mansions27[birthIdx]

	firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	var nextMonth time.Time
	if month == 12 {
		nextMonth = time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC)
	} else {
		nextMonth = time.Date(year, time.Month(month+1), 1, 0, 0, 0, 0, time.UTC)
	}
	daysInMonth := int(nextMonth.Sub(firstDay).Hours() / 24)

	days := make([]DailyFortune, daysInMonth)
	levels := make([]string, daysInMonth)
	var specialDays []MonthSpecialDay
	ryouhanDays := 0

	for i := 0; i < daysInMonth; i++ {
		d := firstDay.AddDate(0, 0, i)
		days[i] = CalculateDaily(birthDate, d, lang)
		levels[i] = days[i].Fortune.Level

		if days[i].SpecialDay != nil {
			specialDays = append(specialDays, MonthSpecialDay{
				Date: days[i].Date,
				Type: days[i].SpecialDay.Type,
				Name: days[i].SpecialDay.Name,
			})
		}
		if days[i].Ryouhan != nil && days[i].Ryouhan.Active {
			ryouhanDays++
		}
	}

	trend := AggregateLevel(levels)

	// Month mansion (day mansion of the 15th)
	midMonth := firstDay.AddDate(0, 0, 14)
	midIdx := engine.DayMansionIndex(midMonth)
	midM := engine.Mansions27[midIdx]

	// Relation between birth mansion and month mansion
	monthRel := engine.GetRelation(birthIdx, midIdx)
	groupNames := map[string]string{
		"eishin": "栄親", "gyotai": "業胎", "mei": "命",
		"yusui": "友衰", "kisei": "危成", "ankai": "安壊",
	}

	// Theme from i18n
	fortunes := loadI18n(lang)
	fortuneData := loadFortuneI18n(lang)
	theme := getMonthlyTheme(fortunes, monthRel.Group)
	advice := getMonthlyAdvice(fortunes, trend)

	// Category descriptions
	levelKey := levelToDescKey(trend)
	fortune := FortuneDetail{
		Level:      trend,
		LevelName:  LevelName(trend, lang),
		BaseLevel:  trend,
		CareerDesc: getCategoryDesc(fortunes, "career", levelKey),
		LoveDesc:   getCategoryDesc(fortunes, "love", levelKey),
		HealthDesc: getCategoryDesc(fortunes, "health", levelKey),
		WealthDesc: getCategoryDesc(fortunes, "wealth", levelKey),
	}
	_ = fortuneData

	// Weekly summaries (group by 7-day chunks)
	weeklies := buildWeeklySummaries(days, firstDay, lang)

	// Ryouhan info
	var ryouhanInfo *RyouhanStat
	if ryouhanDays > 0 {
		ryouhanInfo = &RyouhanStat{
			AffectedDays: ryouhanDays,
			TotalDays:    daysInMonth,
			Ratio:        float64(ryouhanDays) / float64(daysInMonth),
		}
	}

	return MonthlyFortune{
		Year:  year,
		Month: month,
		YourMansion: MansionBrief{
			Index: birthIdx, NameJP: birthM.NameJP,
			Reading: birthM.Reading, Yosei: birthM.Yosei,
		},
		MonthMansion: &MansionBrief{
			Index: midIdx, NameJP: midM.NameJP,
			Reading: midM.Reading, Yosei: midM.Yosei,
		},
		Relation: &MonthRelation{
			Type: monthRel.Group,
			Name: groupNames[monthRel.Group],
		},
		Theme:       theme,
		Fortune:     fortune,
		Weekly:      weeklies,
		SpecialDays: specialDays,
		RyouhanInfo: ryouhanInfo,
		Advice:      advice,
		Days:        days,
		Trend:       trend,
		TrendName:   LevelName(trend, lang),
	}
}

func buildWeeklySummaries(days []DailyFortune, firstDay time.Time, lang string) []WeeklySummary {
	var weeklies []WeeklySummary
	weekNum := 1

	for i := 0; i < len(days); {
		end := i + 7
		if end > len(days) {
			end = len(days)
		}

		chunk := days[i:end]
		levels := make([]string, len(chunk))
		overviews := make([]DailyOverview, len(chunk))
		hasDark := false

		for j, d := range chunk {
			levels[j] = d.Fortune.Level
			sdName := ""
			if d.SpecialDay != nil {
				sdName = d.SpecialDay.Name
			}
			if d.Sanki.PeriodIndex == 1 {
				hasDark = true
			}
			overviews[j] = DailyOverview{
				Date:       d.Date,
				Weekday:    d.Weekday.Name,
				Level:      d.Fortune.Level,
				SpecialDay: sdName,
				Ryouhan:    d.Ryouhan != nil && d.Ryouhan.Active,
				IsDarkWeek: d.Sanki.PeriodIndex == 1,
			}
		}

		weekStart := firstDay.AddDate(0, 0, i)
		weekEnd := firstDay.AddDate(0, 0, end-1)

		weeklies = append(weeklies, WeeklySummary{
			Week:          weekNum,
			WeekStart:     weekStart.Format("2006-01-02"),
			WeekEnd:       weekEnd.Format("2006-01-02"),
			DaysCount:     len(chunk),
			Level:         AggregateLevel(levels),
			HasDarkWeek:   hasDark,
			DailyOverview: overviews,
		})

		weekNum++
		i = end
	}
	return weeklies
}

func getMonthlyTheme(fortunes map[string]any, group string) *ThemeBlock {
	themes, ok := fortunes["monthly_theme_descriptions"].(map[string]any)
	if !ok {
		return nil
	}
	desc, ok := themes[group].(string)
	if !ok {
		return nil
	}
	return &ThemeBlock{Description: desc}
}

func getMonthlyAdvice(fortunes map[string]any, trend string) string {
	advices, ok := fortunes["monthly_fortune_advice"].(map[string]any)
	if !ok {
		return ""
	}
	if a, ok := advices[trend].(string); ok {
		return a
	}
	return ""
}
