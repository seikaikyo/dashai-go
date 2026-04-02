package fortune

import (
	"time"

	"github.com/seikaikyo/dashai-go/internal/shukuyo/engine"
)

// CalendarMonthData is the response for the monthly calendar endpoint.
type CalendarMonthData struct {
	Year  int           `json:"year"`
	Month int           `json:"month"`
	Days  []CalendarDay `json:"days"`
}

// CalendarDay is one day in the calendar.
type CalendarDay struct {
	Date       string              `json:"date"`
	Day        int                 `json:"day"`
	Weekday    string              `json:"weekday"`
	DayMansion CalendarDayMansion  `json:"day_mansion"`
	SpecialDay *engine.SpecialDayType `json:"special_day"`
	Ryouhan    *engine.RyouhanPeriod  `json:"ryouhan"`
	Personal   *CalendarPersonal   `json:"personal"`
}

// CalendarDayMansion is the mansion for a calendar day.
type CalendarDayMansion struct {
	NameJP string `json:"name_jp"`
	Index  int    `json:"index"`
	Yosei  string `json:"yosei"`
}

// CalendarPersonal is the personal fortune for a calendar day.
type CalendarPersonal struct {
	RelationType   string `json:"relation_type"`
	RelationName   string `json:"relation_name"`
	Level          string `json:"level"`
	LevelName      string `json:"level_name"`
	SankiPeriod    string `json:"sanki_period"`
	SankiPeriodIdx int    `json:"sanki_period_index"`
	SankiDayType   string `json:"sanki_day_type,omitempty"`
	IsDarkWeek     bool   `json:"is_dark_week"`
}

// CalculateCalendarMonth builds a monthly calendar with fortune data.
func CalculateCalendarMonth(birthDateStr string, year, month int, lang string) (*CalendarMonthData, error) {
	bd, err := engine.ParseDate(birthDateStr)
	if err != nil {
		return nil, err
	}
	lang = normalizeLang(lang)

	birthIdx, _ := engine.MansionIndexFromDate(bd)
	daysInMonth := time.Date(year, time.Month(month+1), 0, 0, 0, 0, 0, time.UTC).Day()

	// Relation group display names
	groupNames := map[string]string{
		"eishin": "栄親", "gyotai": "業胎", "mei": "命",
		"yusui": "友衰", "kisei": "危成", "ankai": "安壊",
	}

	days := make([]CalendarDay, daysInMonth)
	for d := 1; d <= daysInMonth; d++ {
		date := time.Date(year, time.Month(month), d, 0, 0, 0, 0, time.UTC)
		dayIdx := engine.DayMansionIndex(date)
		dayM := engine.Mansions27[dayIdx]
		rel := engine.GetRelation(birthIdx, dayIdx)
		sanki := engine.GetSankiPosition(birthIdx, dayIdx)
		sd := engine.CheckSpecialDay(date, dayIdx)
		rp := engine.CheckRyouhanPeriod(date)

		sdType := ""
		if sd != nil {
			sdType = sd.Type
		}
		ryouhanActive := rp != nil && rp.Active
		level, _ := DetermineLevel(rel.Group, rel.Direction, ryouhanActive, sdType)

		jpwd := engine.JPWeekday(date)

		// Dark week: period 1 (惡期) is considered the dark week
		isDarkWeek := sanki.PeriodIndex == 1

		days[d-1] = CalendarDay{
			Date:    date.Format("2006-01-02"),
			Day:     d,
			Weekday: WeekdayName(jpwd, lang),
			DayMansion: CalendarDayMansion{
				NameJP: dayM.NameJP,
				Index:  dayIdx,
				Yosei:  dayM.Yosei,
			},
			SpecialDay: sd,
			Ryouhan:    rp,
			Personal: &CalendarPersonal{
				RelationType:   rel.Group,
				RelationName:   groupNames[rel.Group],
				Level:          level,
				LevelName:      LevelName(level, lang),
				SankiPeriod:    sanki.PeriodName,
				SankiPeriodIdx: sanki.PeriodIndex,
				SankiDayType:   sanki.DayType,
				IsDarkWeek:     isDarkWeek,
			},
		}
	}

	return &CalendarMonthData{
		Year:  year,
		Month: month,
		Days:  days,
	}, nil
}
