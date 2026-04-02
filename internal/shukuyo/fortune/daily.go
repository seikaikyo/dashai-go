package fortune

import (
	"time"

	"github.com/seikaikyo/dashai-go/internal/shukuyo/engine"
)

// DailyFortune is the complete daily fortune response.
type DailyFortune struct {
	Date        string            `json:"date"`
	Weekday     string            `json:"weekday"`
	JPWeekday   int               `json:"jp_weekday"`
	DayMansion  MansionBrief      `json:"day_mansion"`
	YourMansion MansionBrief      `json:"your_mansion"`
	Relation    engine.Relation   `json:"relation"`
	Fortune     FortuneDetail     `json:"fortune"`
	SpecialDay  *engine.SpecialDayType `json:"special_day"`
	Ryouhan     *engine.RyouhanPeriod  `json:"ryouhan"`
	Sanki       engine.SankiPosition   `json:"sanki"`
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
	Level     string `json:"level"`
	LevelName string `json:"level_name"`
	BaseLevel string `json:"base_level"`
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

	return DailyFortune{
		Date:      targetDate.Format("2006-01-02"),
		Weekday:   WeekdayName(jpwd, lang),
		JPWeekday: jpwd,
		DayMansion: MansionBrief{
			Index: dayIdx, NameJP: dayM.NameJP,
			Reading: dayM.Reading, Yosei: dayM.Yosei,
		},
		YourMansion: MansionBrief{
			Index: birthIdx, NameJP: birthM.NameJP,
			Reading: birthM.Reading, Yosei: birthM.Yosei,
		},
		Relation: rel,
		Fortune: FortuneDetail{
			Level:     level,
			LevelName: LevelName(level, lang),
			BaseLevel: baseLevel,
		},
		SpecialDay: specialDay,
		Ryouhan:    ryouhan,
		Sanki:      sanki,
	}
}
