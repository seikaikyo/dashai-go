package fortune

import "time"

// WeeklyFortune is the weekly fortune response.
type WeeklyFortune struct {
	StartDate string         `json:"start_date"`
	Days      []DailyFortune `json:"days"`
}

// CalculateWeekly computes an 8-day rolling window (yesterday + today + 6 future).
func CalculateWeekly(birthDate, targetDate time.Time, lang string) WeeklyFortune {
	start := targetDate.AddDate(0, 0, -1) // yesterday
	days := make([]DailyFortune, 8)
	for i := 0; i < 8; i++ {
		d := start.AddDate(0, 0, i)
		days[i] = CalculateDaily(birthDate, d, lang)
	}
	return WeeklyFortune{
		StartDate: start.Format("2006-01-02"),
		Days:      days,
	}
}
