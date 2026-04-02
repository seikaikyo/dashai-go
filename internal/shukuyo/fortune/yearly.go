package fortune

import (
	"time"

	"github.com/seikaikyo/dashai-go/internal/shukuyo/engine"
)

// YearlyFortune is the yearly fortune response.
type YearlyFortune struct {
	Year         int              `json:"year"`
	KuyouStar    engine.KuyouStar `json:"kuyou_star"`
	MonthlyTrend []MonthTrend     `json:"monthly_trend"`
	BestMonths   []int            `json:"best_months"`
}

// MonthTrend holds the aggregated trend for one month.
type MonthTrend struct {
	Month     int    `json:"month"`
	Trend     string `json:"trend"`
	TrendName string `json:"trend_name"`
}

// CalculateYearly computes the yearly fortune using kuyou + monthly aggregation.
func CalculateYearly(birthDate time.Time, year int, lang string) YearlyFortune {
	lang = normalizeLang(lang)
	star := engine.GetKuyouStar(birthDate, year)

	birthIdx, _ := engine.MansionIndexFromDate(birthDate)
	trends := make([]MonthTrend, 12)
	bestLevel := 3 // worst ordinal

	for m := 1; m <= 12; m++ {
		firstDay := time.Date(year, time.Month(m), 1, 0, 0, 0, 0, time.UTC)
		var nextMonth time.Time
		if m == 12 {
			nextMonth = time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC)
		} else {
			nextMonth = time.Date(year, time.Month(m+1), 1, 0, 0, 0, 0, time.UTC)
		}
		daysInMonth := int(nextMonth.Sub(firstDay).Hours() / 24)

		levels := make([]string, daysInMonth)
		for d := 0; d < daysInMonth; d++ {
			dt := firstDay.AddDate(0, 0, d)
			dayIdx := engine.DayMansionIndex(dt)
			rel := engine.GetRelation(birthIdx, dayIdx)
			sd := engine.CheckSpecialDay(dt, dayIdx)
			rh := engine.CheckRyouhanPeriod(dt)

			sdType := ""
			if sd != nil {
				sdType = sd.Type
			}
			lv, _ := DetermineLevel(rel.Group, rel.Direction, rh != nil, sdType)
			levels[d] = lv
		}

		trend := AggregateLevel(levels)
		trends[m-1] = MonthTrend{
			Month:     m,
			Trend:     trend,
			TrendName: LevelName(trend, lang),
		}

		if idx := levelIndex[trend]; idx < bestLevel {
			bestLevel = idx
		}
	}

	// Find best months (matching the best level found)
	var best []int
	for _, t := range trends {
		if levelIndex[t.Trend] == bestLevel {
			best = append(best, t.Month)
		}
	}

	return YearlyFortune{
		Year:         year,
		KuyouStar:    star,
		MonthlyTrend: trends,
		BestMonths:   best,
	}
}

// CalculateYearlyRange computes yearly fortune for a range of years.
func CalculateYearlyRange(birthDate time.Time, startYear, endYear int, lang string) []YearlyFortune {
	if endYear-startYear > 10 {
		endYear = startYear + 10
	}
	results := make([]YearlyFortune, 0, endYear-startYear+1)
	for y := startYear; y <= endYear; y++ {
		results = append(results, CalculateYearly(birthDate, y, lang))
	}
	return results
}
