package fortune

import (
	"fmt"
	"time"

	"github.com/seikaikyo/dashai-go/internal/shukuyo/engine"
)

// YearlyFortune is the yearly fortune response.
type YearlyFortune struct {
	Year                 int                 `json:"year"`
	YourMansion          MansionBrief        `json:"your_mansion"`
	KuyouStar            KuyouStarEnriched   `json:"kuyou_star"`
	Fortune              FortuneDetail       `json:"fortune"`
	Theme                *ThemeBlock         `json:"theme,omitempty"`
	CategoryDescriptions *CategoryDescs      `json:"category_descriptions,omitempty"`
	MonthlyTrend         []MonthTrend        `json:"monthly_trend"`
	BestMonths           []int               `json:"best_months"`
	Opportunities        []string            `json:"opportunities,omitempty"`
	Warnings             []string            `json:"warnings,omitempty"`
	Advice               string              `json:"advice,omitempty"`
}

// KuyouStarEnriched extends engine.KuyouStar with descriptive text.
type KuyouStarEnriched struct {
	Index       int    `json:"index"`
	Name        string `json:"name"`
	Reading     string `json:"reading"`
	Yosei       string `json:"yosei,omitempty"`
	Buddha      string `json:"buddha"`
	KazoeAge    int    `json:"kazoe_age"`
	Level       string `json:"level"`
	FortuneName string `json:"fortune_name"`
	Description string `json:"description,omitempty"`
}

// CategoryDescs holds four-category yearly descriptions.
type CategoryDescs struct {
	Career string `json:"career,omitempty"`
	Love   string `json:"love,omitempty"`
	Health string `json:"health,omitempty"`
	Wealth string `json:"wealth,omitempty"`
}

// MonthTrend holds the aggregated trend for one month.
type MonthTrend struct {
	Month        int     `json:"month"`
	Trend        string  `json:"trend"`
	TrendName    string  `json:"trend_name"`
	RelationType string  `json:"relation_type,omitempty"`
	RyouhanRatio float64 `json:"ryouhan_ratio,omitempty"`
}

// CalculateYearly computes the yearly fortune using kuyou + monthly aggregation.
func CalculateYearly(birthDate time.Time, year int, lang string) YearlyFortune {
	lang = normalizeLang(lang)
	star := engine.GetKuyouStar(birthDate, year)

	birthIdx, _ := engine.MansionIndexFromDate(birthDate)
	birthM := engine.Mansions27[birthIdx]

	trends := make([]MonthTrend, 12)
	allLevels := make([]string, 0, 365)
	bestLevel := 3

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
		ryouhanDays := 0
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
			rhActive := rh != nil
			lv, _ := DetermineLevel(rel.Group, rel.Direction, rhActive, sdType)
			levels[d] = lv
			allLevels = append(allLevels, lv)
			if rhActive {
				ryouhanDays++
			}
		}

		trend := AggregateLevel(levels)

		// Month relation (mid-month mansion)
		midDay := firstDay.AddDate(0, 0, 14)
		midIdx := engine.DayMansionIndex(midDay)
		midRel := engine.GetRelation(birthIdx, midIdx)

		trends[m-1] = MonthTrend{
			Month:        m,
			Trend:        trend,
			TrendName:    LevelName(trend, lang),
			RelationType: midRel.Group,
			RyouhanRatio: float64(ryouhanDays) / float64(daysInMonth),
		}

		if idx := levelIndex[trend]; idx < bestLevel {
			bestLevel = idx
		}
	}

	// Best months
	var best []int
	var kyoMonths []int
	for _, t := range trends {
		if levelIndex[t.Trend] == bestLevel {
			best = append(best, t.Month)
		}
		if t.Trend == LevelKyo {
			kyoMonths = append(kyoMonths, t.Month)
		}
	}

	// Aggregate yearly fortune
	yearLevel := AggregateLevel(allLevels)
	levelKey := levelToDescKey(yearLevel)
	fortunes := loadI18n(lang)
	kuyouI18n := loadKuyouI18n(lang)

	fortune := FortuneDetail{
		Level:      yearLevel,
		LevelName:  LevelName(yearLevel, lang),
		BaseLevel:  yearLevel,
		CareerDesc: getCategoryDesc(fortunes, "career", levelKey),
		LoveDesc:   getCategoryDesc(fortunes, "love", levelKey),
		HealthDesc: getCategoryDesc(fortunes, "health", levelKey),
		WealthDesc: getCategoryDesc(fortunes, "wealth", levelKey),
	}

	// Enrich kuyou star
	enrichedStar := enrichKuyouStar(star, kuyouI18n)

	// Theme
	theme := getYearlyTheme(fortunes, yearLevel)

	// Category descriptions
	catDescs := &CategoryDescs{
		Career: fortune.CareerDesc,
		Love:   fortune.LoveDesc,
		Health: fortune.HealthDesc,
		Wealth: fortune.WealthDesc,
	}

	// Opportunities and warnings
	opportunities := buildOpportunities(best, lang)
	warnings := buildWarnings(kyoMonths, lang)

	// Advice
	advice := getYearlyAdvice(fortunes, yearLevel)

	return YearlyFortune{
		Year: year,
		YourMansion: MansionBrief{
			Index: birthIdx, NameJP: birthM.NameJP,
			Reading: birthM.Reading, Yosei: birthM.Yosei,
		},
		KuyouStar:            enrichedStar,
		Fortune:              fortune,
		Theme:                theme,
		CategoryDescriptions: catDescs,
		MonthlyTrend:         trends,
		BestMonths:           best,
		Opportunities:        opportunities,
		Warnings:             warnings,
		Advice:               advice,
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

func enrichKuyouStar(star engine.KuyouStar, kuyouI18n map[string]any) KuyouStarEnriched {
	enriched := KuyouStarEnriched{
		Index:    star.Index,
		Name:     star.Name,
		Reading:  star.Reading,
		Yosei:    star.Yosei,
		Buddha:   star.Buddha,
		KazoeAge: star.KazoeAge,
	}

	// Level from index
	// 九曜流年等級 — 寺院傳承（放生寺/大聖院/岡寺）
	// Index: 0=羅喉 1=土 2=水 3=金 4=日 5=火 6=計都 7=月 8=木
	kuyouLevels := [9]string{"kyo", "shokyo", "shokyo", "shokyo", "daikichi", "kyo", "kyo", "daikichi", "daikichi"}
	kuyouFortuneNames := [9]string{"大凶", "半吉", "末吉", "半吉", "大吉", "大凶", "大凶", "大吉", "大吉"}
	if star.Index >= 0 && star.Index < 9 {
		enriched.Level = kuyouLevels[star.Index]
		enriched.FortuneName = kuyouFortuneNames[star.Index]
	}

	// Description from i18n
	if stars, ok := kuyouI18n["stars"].(map[string]any); ok {
		key := fmt.Sprintf("%d", star.Index)
		if entry, ok := stars[key].(map[string]any); ok {
			if desc, ok := entry["description"].(string); ok {
				enriched.Description = desc
			}
		}
	}

	return enriched
}

func getYearlyTheme(fortunes map[string]any, level string) *ThemeBlock {
	themes, ok := fortunes["yearly_theme_descriptions"].(map[string]any)
	if !ok {
		return nil
	}
	// Map level to theme key
	themeKey := "neutral"
	switch level {
	case LevelDaikichi:
		themeKey = "generating"
	case LevelKichi:
		themeKey = "generating"
	case LevelShokyo:
		themeKey = "weakening"
	case LevelKyo:
		themeKey = "kyo"
	}
	if desc, ok := themes[themeKey].(string); ok {
		return &ThemeBlock{Description: desc}
	}
	return nil
}

func getYearlyAdvice(fortunes map[string]any, level string) string {
	advices, ok := fortunes["yearly_fortune_advice"].(map[string]any)
	if !ok {
		return ""
	}
	if a, ok := advices[level].(string); ok {
		return a
	}
	return ""
}

func buildOpportunities(bestMonths []int, lang string) []string {
	if len(bestMonths) == 0 {
		return nil
	}
	var ops []string
	for _, m := range bestMonths {
		ops = append(ops, fmt.Sprintf("%d月", m))
	}
	return ops
}

func buildWarnings(kyoMonths []int, lang string) []string {
	if len(kyoMonths) == 0 {
		return nil
	}
	var warns []string
	for _, m := range kyoMonths {
		warns = append(warns, fmt.Sprintf("%d月", m))
	}
	return warns
}
