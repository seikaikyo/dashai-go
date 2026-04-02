package fortune

import "time"

// MonthlyFortune is the monthly fortune response.
type MonthlyFortune struct {
	Year  int            `json:"year"`
	Month int            `json:"month"`
	Days  []DailyFortune `json:"days"`
	Trend string         `json:"trend"`
	TrendName string     `json:"trend_name"`
}

// CalculateMonthly computes daily fortune for every day of a month.
func CalculateMonthly(birthDate time.Time, year, month int, lang string) MonthlyFortune {
	lang = normalizeLang(lang)
	firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)

	// Days in this month
	var nextMonth time.Time
	if month == 12 {
		nextMonth = time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC)
	} else {
		nextMonth = time.Date(year, time.Month(month+1), 1, 0, 0, 0, 0, time.UTC)
	}
	daysInMonth := int(nextMonth.Sub(firstDay).Hours() / 24)

	days := make([]DailyFortune, daysInMonth)
	levels := make([]string, daysInMonth)
	for i := 0; i < daysInMonth; i++ {
		d := firstDay.AddDate(0, 0, i)
		days[i] = CalculateDaily(birthDate, d, lang)
		levels[i] = days[i].Fortune.Level
	}

	trend := AggregateLevel(levels)

	return MonthlyFortune{
		Year:      year,
		Month:     month,
		Days:      days,
		Trend:     trend,
		TrendName: LevelName(trend, lang),
	}
}
