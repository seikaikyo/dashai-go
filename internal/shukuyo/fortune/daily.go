package fortune

import (
	"math/rand"
	"time"

	"github.com/seikaikyo/dashai-go/internal/shukuyo/engine"
)

// DailyFortune is the complete daily fortune response.
type DailyFortune struct {
	Date        string              `json:"date"`
	Weekday     WeekdayInfo         `json:"weekday"`
	JPWeekday   int                 `json:"jp_weekday"`
	DayMansion  DayMansionInfo      `json:"day_mansion"`
	YourMansion MansionBrief        `json:"your_mansion"`
	Relation    engine.Relation     `json:"mansion_relation"`
	Fortune     FortuneDetail       `json:"fortune"`
	Advice      string              `json:"advice"`
	Lucky       *LuckyInfo          `json:"lucky,omitempty"`
	SpecialDay  *engine.SpecialDayType `json:"special_day"`
	Ryouhan     *engine.RyouhanPeriod  `json:"ryouhan"`
	Sanki       engine.SankiPosition   `json:"sanki"`
}

// WeekdayInfo holds rich weekday info including planet.
type WeekdayInfo struct {
	Name    string `json:"name"`
	Reading string `json:"reading,omitempty"`
	Yosei   string `json:"yosei,omitempty"`
	Planet  string `json:"planet,omitempty"`
}

// DayMansionInfo extends MansionBrief with day fortune.
type DayMansionInfo struct {
	Index      int             `json:"index"`
	NameJP     string          `json:"name_jp"`
	Reading    string          `json:"reading"`
	Yosei      string          `json:"yosei"`
	DayFortune *DayFortuneText `json:"day_fortune,omitempty"`
}

// DayFortuneText holds the sutra-based day fortune descriptions.
type DayFortuneText struct {
	Auspicious       []string `json:"auspicious,omitempty"`
	Inauspicious     []string `json:"inauspicious,omitempty"`
	Summary          string   `json:"summary,omitempty"`
	IsMostAuspicious bool     `json:"is_most_auspicious"`
}

// MansionBrief is a compact mansion reference.
type MansionBrief struct {
	Index   int    `json:"index"`
	NameJP  string `json:"name_jp"`
	Reading string `json:"reading"`
	Yosei   string `json:"yosei"`
}

// FortuneDetail holds the level and category descriptions.
type FortuneDetail struct {
	Level      string `json:"level"`
	LevelName  string `json:"level_name"`
	BaseLevel  string `json:"base_level"`
	CareerDesc string `json:"career_desc,omitempty"`
	LoveDesc   string `json:"love_desc,omitempty"`
	HealthDesc string `json:"health_desc,omitempty"`
	WealthDesc string `json:"wealth_desc,omitempty"`
}

// LuckyInfo holds lucky direction, color, and numbers.
type LuckyInfo struct {
	Direction        string `json:"direction"`
	DirectionReading string `json:"direction_reading,omitempty"`
	Color            string `json:"color"`
	ColorReading     string `json:"color_reading,omitempty"`
	ColorHex         string `json:"color_hex,omitempty"`
	Numbers          []int  `json:"numbers"`
}

// CalculateDaily computes the daily fortune for a person.
func CalculateDaily(birthDate, targetDate time.Time, lang string) DailyFortune {
	lang = normalizeLang(lang)

	// Engine calculations (all T21n1299 original)
	birthIdx, _ := engine.MansionIndexFromDate(birthDate)
	dayIdx := engine.DayMansionIndex(targetDate)
	rel := engine.GetRelation(birthIdx, dayIdx)
	sanki := engine.GetSankiPosition(birthIdx, dayIdx)
	jpwd := engine.JPWeekday(targetDate)
	specialDay := engine.CheckSpecialDay(targetDate, dayIdx)
	ryouhan := engine.CheckRyouhanPeriod(targetDate)

	// Modern interpretation layer
	sdType := ""
	if specialDay != nil {
		sdType = specialDay.Type
	}
	ryouhanActive := ryouhan != nil
	level, baseLevel := DetermineLevel(rel.Group, rel.Direction, ryouhanActive, sdType)

	birthM := engine.Mansions27[birthIdx]
	dayM := engine.Mansions27[dayIdx]

	// Load i18n texts
	fortunes := loadI18n(lang)
	fortuneData := loadFortuneI18n(lang)

	// Weekday info
	weekday := buildWeekdayInfo(jpwd, lang, fortuneData)

	// Category descriptions by level
	levelKey := levelToDescKey(level)
	fortune := FortuneDetail{
		Level:      level,
		LevelName:  LevelName(level, lang),
		BaseLevel:  baseLevel,
		CareerDesc: getCategoryDesc(fortunes, "career", levelKey),
		LoveDesc:   getCategoryDesc(fortunes, "love", levelKey),
		HealthDesc: getCategoryDesc(fortunes, "health", levelKey),
		WealthDesc: getCategoryDesc(fortunes, "wealth", levelKey),
	}

	// Advice text
	advice := getDailyAdvice(fortuneData, levelKey)

	// Lucky items based on yosei
	lucky := buildLuckyInfo(birthM.Yosei, fortuneData, targetDate)

	// Day fortune descriptions
	dayFortune := getDayFortune(fortunes, rel.Group)

	return DailyFortune{
		Date:    targetDate.Format("2006-01-02"),
		Weekday: weekday,
		JPWeekday: jpwd,
		DayMansion: DayMansionInfo{
			Index:      dayIdx,
			NameJP:     dayM.NameJP,
			Reading:    dayM.Reading,
			Yosei:      dayM.Yosei,
			DayFortune: dayFortune,
		},
		YourMansion: MansionBrief{
			Index: birthIdx, NameJP: birthM.NameJP,
			Reading: birthM.Reading, Yosei: birthM.Yosei,
		},
		Relation:   rel,
		Fortune:    fortune,
		Advice:     advice,
		Lucky:      lucky,
		SpecialDay: specialDay,
		Ryouhan:    ryouhan,
		Sanki:      sanki,
	}
}

// --- Helper functions for enrichment ---

func buildWeekdayInfo(jpwd int, lang string, fortuneData map[string]any) WeekdayInfo {
	lk := langKey(lang)
	name := ""
	if jpwd >= 0 && jpwd <= 6 {
		name = weekdayShort[lk][jpwd]
	}

	// Planet from fortune_data.json
	planet := ""
	if planets, ok := fortuneData["weekday_planets"].(map[string]any); ok {
		key := []string{"0", "1", "2", "3", "4", "5", "6"}
		if jpwd >= 0 && jpwd < 7 {
			if p, ok := planets[key[jpwd]].(string); ok {
				planet = p
			}
		}
	}

	// Yosei for weekdays: 日月火水木金土
	weekdayYosei := [7]string{"日", "月", "火", "水", "木", "金", "土"}
	yosei := ""
	if jpwd >= 0 && jpwd < 7 {
		yosei = weekdayYosei[jpwd]
	}

	return WeekdayInfo{
		Name:   name,
		Yosei:  yosei,
		Planet: planet,
	}
}

func levelToDescKey(level string) string {
	switch level {
	case LevelDaikichi:
		return "excellent"
	case LevelKichi:
		return "good"
	case LevelShokyo:
		return "fair"
	case LevelKyo:
		return "caution"
	default:
		return "fair"
	}
}

func getCategoryDesc(fortunes map[string]any, category, levelKey string) string {
	descs, ok := fortunes["daily_category_descriptions"].(map[string]any)
	if !ok {
		return ""
	}
	catMap, ok := descs[category].(map[string]any)
	if !ok {
		return ""
	}
	entry, ok := catMap[levelKey].(map[string]any)
	if !ok {
		return ""
	}
	// Extract the text (might be in "text" or "classic" field)
	if text, ok := entry["text"].(string); ok {
		return text
	}
	if text, ok := entry["classic"].(string); ok {
		return text
	}
	return ""
}

func getDailyAdvice(fortuneData map[string]any, levelKey string) string {
	adviceMap, ok := fortuneData["daily_advice"].(map[string]any)
	if !ok {
		return ""
	}
	items, ok := adviceMap[levelKey].([]any)
	if !ok || len(items) == 0 {
		return ""
	}
	// Pick one randomly
	idx := rand.Intn(len(items))
	item, ok := items[idx].(map[string]any)
	if !ok {
		return ""
	}
	if text, ok := item["text"].(string); ok {
		return text
	}
	return ""
}

func buildLuckyInfo(yosei string, fortuneData map[string]any, date time.Time) *LuckyInfo {
	luckyItems, ok := fortuneData["lucky_items"].(map[string]any)
	if !ok {
		return nil
	}

	// Direction from yosei
	direction := ""
	if dirs, ok := luckyItems["directions"].(map[string]any); ok {
		if d, ok := dirs[yosei].(string); ok {
			direction = d
		}
	}

	// Color from yosei
	color := ""
	colorHex := ""
	if colors, ok := luckyItems["colors"].(map[string]any); ok {
		if c, ok := colors[yosei].(map[string]any); ok {
			if n, ok := c["name"].(string); ok {
				color = n
			}
			if h, ok := c["hex"].(string); ok {
				colorHex = h
			}
		}
	}

	// Lucky numbers: deterministic from date + yosei
	seed := int64(date.Year()*10000+int(date.Month())*100+date.Day()) + int64(len(yosei))
	rng := rand.New(rand.NewSource(seed))
	numbers := []int{rng.Intn(9) + 1, rng.Intn(9) + 1, rng.Intn(9) + 1}

	if direction == "" && color == "" {
		return nil
	}

	return &LuckyInfo{
		Direction: direction,
		Color:     color,
		ColorHex:  colorHex,
		Numbers:   numbers,
	}
}

func getDayFortune(fortunes map[string]any, group string) *DayFortuneText {
	descs, ok := fortunes["daily_fortune_descriptions"].(map[string]any)
	if !ok {
		return nil
	}
	groupData, ok := descs[group].(map[string]any)
	if !ok {
		return nil
	}

	var auspicious, inauspicious []string
	summary := ""

	if items, ok := groupData["items"].([]any); ok {
		for _, item := range items {
			if s, ok := item.(string); ok {
				auspicious = append(auspicious, s)
			}
		}
	}
	if s, ok := groupData["classic"].(string); ok {
		summary = s
	}

	return &DayFortuneText{
		Auspicious:       auspicious,
		Inauspicious:     inauspicious,
		Summary:          summary,
		IsMostAuspicious: group == "eishin",
	}
}
