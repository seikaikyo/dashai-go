package career

import (
	"sort"
	"time"

	"github.com/seikaikyo/dashai-go/internal/shukuyo/engine"
)

// InterviewDates evaluates auspicious dates for interviews/applications.
func InterviewDates(req InterviewDatesRequest) (*InterviewDatesResponse, error) {
	birth, err := engine.ParseDate(req.BirthDate)
	if err != nil {
		return nil, err
	}
	compDate, err := engine.ParseDate(req.CompanyDate)
	if err != nil {
		return nil, err
	}

	lang := normalizeLang(req.Lang)
	days := req.Days
	if days <= 0 {
		days = 30
	}
	if days > 90 {
		days = 90
	}

	birthIdx, _ := engine.MansionIndexFromDate(birth)
	compIdx, _ := engine.MansionIndexFromDate(compDate)

	// Base relation between person and company (doesn't change daily)
	baseRel := engine.GetRelation(birthIdx, compIdx)
	_ = baseRel // used for context, daily relation is what varies

	today := time.Now().Truncate(24 * time.Hour)

	// Evaluate target date (today)
	target := evaluateDay(birth, birthIdx, today, lang)

	// Scan alternatives
	alternatives := make([]InterviewDay, 0, days)
	for i := 1; i <= days; i++ {
		d := today.AddDate(0, 0, i)
		day := evaluateDay(birth, birthIdx, d, lang)
		alternatives = append(alternatives, day)
	}

	// Sort: best verdict first, then by date proximity
	sort.SliceStable(alternatives, func(i, j int) bool {
		return verdictOrder(alternatives[i].Verdict) < verdictOrder(alternatives[j].Verdict)
	})

	// Return top 5 good dates
	top := make([]InterviewDay, 0, 5)
	for _, alt := range alternatives {
		if alt.Verdict == "warning" {
			continue
		}
		top = append(top, alt)
		if len(top) >= 5 {
			break
		}
	}

	return &InterviewDatesResponse{
		Target:       target,
		Alternatives: top,
	}, nil
}

func evaluateDay(birth time.Time, birthIdx int, date time.Time, lang string) InterviewDay {
	dayIdx := engine.DayMansionIndex(date)
	rel := engine.GetRelation(birthIdx, dayIdx)

	// Career level for the day's relation
	level := CareerLevel(rel.Group)

	// Check special day and ryouhan
	sd := engine.CheckSpecialDay(date, dayIdx)
	rp := engine.CheckRyouhanPeriod(date)

	specialDay := ""
	if sd != nil {
		specialDay = sd.Name
	}
	ryouhan := rp != nil && rp.Active

	// Determine verdict
	verdict := dayVerdict(rel.Group, rel.Direction, ryouhan, specialDay)

	// Weekday
	jpwd := engine.JPWeekday(date)
	wdNames := map[string][7]string{
		"zh": {"日", "月", "火", "水", "木", "金", "土"},
		"ja": {"日", "月", "火", "水", "木", "金", "土"},
		"en": {"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"},
	}
	lk := langKey(lang)
	weekday := ""
	if jpwd >= 0 && jpwd <= 6 {
		weekday = wdNames[lk][jpwd]
	}

	return InterviewDay{
		Date:       date.Format("2006-01-02"),
		Weekday:    weekday,
		Relation:   GroupName(rel.Group, lang),
		Level:      level,
		LevelName:  CareerLevelName(level, lang),
		Verdict:    verdict,
		SpecialDay: specialDay,
		Ryouhan:    ryouhan,
	}
}

func dayVerdict(group, direction string, ryouhan bool, specialDay string) string {
	// Avoid ankai days entirely
	if group == "ankai" {
		return "warning"
	}

	// Ryouhan period flips fortune
	if ryouhan {
		switch group {
		case "eishin":
			return "fair" // normally excellent, but ryouhan reverses
		case "ankai":
			return "fair" // normally bad, ryouhan improves
		}
	}

	// Special days boost
	if specialDay == "甘露日" || specialDay == "金剛峯日" {
		if group == "eishin" || group == "gyotai" || group == "kisei" {
			return "excellent"
		}
	}

	switch group {
	case "eishin":
		return "excellent"
	case "gyotai", "mei":
		return "good"
	case "yusui", "kisei":
		return "fair"
	default:
		return "warning"
	}
}

func verdictOrder(verdict string) int {
	switch verdict {
	case "excellent":
		return 0
	case "good":
		return 1
	case "fair":
		return 2
	case "warning":
		return 3
	default:
		return 4
	}
}
