package fortune

import (
	"sort"
	"time"

	"github.com/seikaikyo/dashai-go/internal/shukuyo/engine"
)

// LuckyDay is a single date evaluation for the lucky days endpoint.
type LuckyDay struct {
	Date      string `json:"date"`
	Weekday   string `json:"weekday"`
	Level     string `json:"level"`
	LevelName string `json:"level_name"`
	Reason    string `json:"reason"`
	Tip       string `json:"tip,omitempty"`
}

// LuckyDaySummary is the response for the lucky days summary endpoint.
type LuckyDaySummary struct {
	YourMansion MansionBrief         `json:"your_mansion"`
	Categories  []LuckyDayCategory   `json:"categories"`
}

// LuckyDayCategory groups lucky days by activity type.
type LuckyDayCategory struct {
	Key     string            `json:"key"`
	Name    string            `json:"name"`
	Actions []LuckyDayAction  `json:"actions"`
}

// LuckyDayAction is a specific activity with its lucky days.
type LuckyDayAction struct {
	Key       string     `json:"key"`
	Name      string     `json:"name"`
	LuckyDays []LuckyDay `json:"lucky_days"`
}

// LuckyCalendarData is the response for the lucky days calendar view.
type LuckyCalendarData struct {
	Year        int                            `json:"year"`
	Month       int                            `json:"month"`
	YourMansion MansionBrief                   `json:"your_mansion"`
	Days        map[string][]LuckyCalendarDay  `json:"days"`
}

// LuckyCalendarDay is one activity evaluation for a calendar day.
type LuckyCalendarDay struct {
	Category     string `json:"category"`
	CategoryName string `json:"category_name"`
	Action       string `json:"action"`
	ActionName   string `json:"action_name"`
	Level        string `json:"level"`
	Rating       string `json:"rating"`
	Reason       string `json:"reason"`
}

// Action definitions with sutra-based criteria.
type luckyAction struct {
	key      string
	name     string
	category string
	// goodGroups: relation groups that are auspicious for this action
	goodGroups []string
	// badGroups: relation groups to avoid
	badGroups []string
}

var luckyCategories = map[string]string{
	"career":  "career",
	"social":  "social",
	"health":  "health",
	"move":    "move",
}

var luckyActions = []luckyAction{
	// Career
	{key: "interview", name: "面試", category: "career", goodGroups: []string{"eishin", "gyotai", "kisei"}, badGroups: []string{"ankai"}},
	{key: "negotiate", name: "談判", category: "career", goodGroups: []string{"eishin", "kisei"}, badGroups: []string{"ankai", "yusui"}},
	{key: "launch", name: "上線/發布", category: "career", goodGroups: []string{"eishin", "mei"}, badGroups: []string{"ankai"}},
	{key: "founding", name: "開業/創業", category: "career", goodGroups: []string{"eishin", "gyotai"}, badGroups: []string{"ankai", "yusui"}},
	// Social
	{key: "meeting", name: "聚會", category: "social", goodGroups: []string{"eishin", "yusui", "kisei"}, badGroups: []string{"ankai"}},
	{key: "dating", name: "約會", category: "social", goodGroups: []string{"eishin", "gyotai"}, badGroups: []string{"ankai"}},
	// Health
	{key: "medical", name: "就醫", category: "health", goodGroups: []string{"yusui", "mei"}, badGroups: []string{"ankai"}},
	// Move
	{key: "travel", name: "遠行", category: "move", goodGroups: []string{"eishin", "gyotai"}, badGroups: []string{"kisei", "ankai"}},
	{key: "relocate", name: "搬遷", category: "move", goodGroups: []string{"eishin"}, badGroups: []string{"yusui", "ankai"}},
}

// CalculateLuckyDays computes lucky days summary for a birth date.
func CalculateLuckyDays(birthDateStr string, lang string) (*LuckyDaySummary, error) {
	bd, err := engine.ParseDate(birthDateStr)
	if err != nil {
		return nil, err
	}
	lang = normalizeLang(lang)

	birthIdx, _ := engine.MansionIndexFromDate(bd)
	birthM := engine.Mansions27[birthIdx]

	today := time.Now().Truncate(24 * time.Hour)

	// Scan next 30 days
	dayEvals := make([]dayEval, 30)
	for i := 0; i < 30; i++ {
		d := today.AddDate(0, 0, i)
		dayIdx := engine.DayMansionIndex(d)
		rel := engine.GetRelation(birthIdx, dayIdx)
		sd := engine.CheckSpecialDay(d, dayIdx)
		rp := engine.CheckRyouhanPeriod(d)
		sdType := ""
		if sd != nil {
			sdType = sd.Type
		}
		ryouhanActive := rp != nil && rp.Active
		level, _ := DetermineLevel(rel.Group, rel.Direction, ryouhanActive, sdType)

		dayEvals[i] = dayEval{
			date:      d,
			dayIdx:    dayIdx,
			rel:       rel,
			level:     level,
			special:   sd,
			ryouhan:   ryouhanActive,
		}
	}

	// Group by category
	categoryMap := map[string]*LuckyDayCategory{}
	for _, action := range luckyActions {
		cat, ok := categoryMap[action.category]
		if !ok {
			cat = &LuckyDayCategory{Key: action.category, Name: categoryName(action.category, lang)}
			categoryMap[action.category] = cat
		}

		luckyDays := findLuckyDaysForAction(dayEvals, action, lang)
		cat.Actions = append(cat.Actions, LuckyDayAction{
			Key:       action.key,
			Name:      actionName(action, lang),
			LuckyDays: luckyDays,
		})
	}

	categories := make([]LuckyDayCategory, 0, len(categoryMap))
	for _, cat := range []string{"career", "social", "health", "move"} {
		if c, ok := categoryMap[cat]; ok {
			categories = append(categories, *c)
		}
	}

	return &LuckyDaySummary{
		YourMansion: MansionBrief{
			Index: birthIdx, NameJP: birthM.NameJP,
			Reading: birthM.Reading, Yosei: birthM.Yosei,
		},
		Categories: categories,
	}, nil
}

// CalculateLuckyCalendar computes the lucky day calendar for a month.
func CalculateLuckyCalendar(birthDateStr string, year, month int, lang string) (*LuckyCalendarData, error) {
	bd, err := engine.ParseDate(birthDateStr)
	if err != nil {
		return nil, err
	}
	lang = normalizeLang(lang)

	birthIdx, _ := engine.MansionIndexFromDate(bd)
	birthM := engine.Mansions27[birthIdx]

	start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	daysInMonth := time.Date(year, time.Month(month+1), 0, 0, 0, 0, 0, time.UTC).Day()

	days := map[string][]LuckyCalendarDay{}

	for d := 1; d <= daysInMonth; d++ {
		date := time.Date(year, time.Month(month), d, 0, 0, 0, 0, time.UTC)
		dateStr := date.Format("2006-01-02")
		dayIdx := engine.DayMansionIndex(date)
		rel := engine.GetRelation(birthIdx, dayIdx)
		sd := engine.CheckSpecialDay(date, dayIdx)
		rp := engine.CheckRyouhanPeriod(date)
		sdType := ""
		if sd != nil {
			sdType = sd.Type
		}
		ryouhanActive := rp != nil && rp.Active
		level, _ := DetermineLevel(rel.Group, rel.Direction, ryouhanActive, sdType)
		_ = start

		entries := make([]LuckyCalendarDay, 0)
		for _, action := range luckyActions {
			rating := rateAction(rel.Group, level, action)
			entries = append(entries, LuckyCalendarDay{
				Category:     action.category,
				CategoryName: categoryName(action.category, lang),
				Action:       action.key,
				ActionName:   actionName(action, lang),
				Level:        level,
				Rating:       rating,
				Reason:       rel.Group + " / " + rel.Direction,
			})
		}
		days[dateStr] = entries
	}

	return &LuckyCalendarData{
		Year:  year,
		Month: month,
		YourMansion: MansionBrief{
			Index: birthIdx, NameJP: birthM.NameJP,
			Reading: birthM.Reading, Yosei: birthM.Yosei,
		},
		Days: days,
	}, nil
}

// PairLuckyDaysResult is the response for pair lucky days.
type PairLuckyDaysResult struct {
	Person1       MansionBrief       `json:"person1"`
	Person2       MansionBrief       `json:"person2"`
	Compatibility PairCompat         `json:"compatibility"`
	Actions       []PairLuckyAction  `json:"actions"`
}

// PairCompat is the compatibility summary for pair lucky days.
type PairCompat struct {
	Relation    string `json:"relation"`
	Level       string `json:"level"`
	Description string `json:"description"`
}

// PairLuckyAction groups lucky days for a pair activity.
type PairLuckyAction struct {
	Action    string     `json:"action"`
	Name      string     `json:"name"`
	LuckyDays []LuckyDay `json:"lucky_days"`
}

// CalculatePairLuckyDays finds good days for two people to do activities together.
func CalculatePairLuckyDays(date1Str, date2Str, lang string) (*PairLuckyDaysResult, error) {
	d1, err := engine.ParseDate(date1Str)
	if err != nil {
		return nil, err
	}
	d2, err := engine.ParseDate(date2Str)
	if err != nil {
		return nil, err
	}
	lang = normalizeLang(lang)

	idx1, _ := engine.MansionIndexFromDate(d1)
	idx2, _ := engine.MansionIndexFromDate(d2)
	m1 := engine.Mansions27[idx1]
	m2 := engine.Mansions27[idx2]
	pairRel := engine.GetRelation(idx1, idx2)

	today := time.Now().Truncate(24 * time.Hour)

	pairActions := []luckyAction{
		{key: "meeting", name: "聚會", category: "social", goodGroups: []string{"eishin", "yusui"}, badGroups: []string{"ankai"}},
		{key: "dating", name: "約會", category: "social", goodGroups: []string{"eishin", "gyotai"}, badGroups: []string{"ankai"}},
		{key: "collaboration", name: "合作", category: "career", goodGroups: []string{"eishin", "kisei"}, badGroups: []string{"ankai"}},
	}

	actions := make([]PairLuckyAction, 0, len(pairActions))
	for _, action := range pairActions {
		var luckyDays []LuckyDay
		for i := 0; i < 30; i++ {
			d := today.AddDate(0, 0, i)
			dayIdx := engine.DayMansionIndex(d)

			// Both people need a decent day
			rel1 := engine.GetRelation(idx1, dayIdx)
			rel2 := engine.GetRelation(idx2, dayIdx)

			sd := engine.CheckSpecialDay(d, dayIdx)
			rp := engine.CheckRyouhanPeriod(d)
			sdType := ""
			if sd != nil {
				sdType = sd.Type
			}
			ryouhanActive := rp != nil && rp.Active

			lv1, _ := DetermineLevel(rel1.Group, rel1.Direction, ryouhanActive, sdType)
			lv2, _ := DetermineLevel(rel2.Group, rel2.Direction, ryouhanActive, sdType)

			// Both need at least shokyo, and at least one should be good
			if (lv1 == LevelKyo) || (lv2 == LevelKyo) {
				continue
			}

			rating := rateAction(rel1.Group, lv1, action)
			if rating == "avoid" {
				continue
			}

			jpwd := engine.JPWeekday(d)
			luckyDays = append(luckyDays, LuckyDay{
				Date:      d.Format("2006-01-02"),
				Weekday:   WeekdayName(jpwd, lang),
				Level:     betterLevel(lv1, lv2),
				LevelName: LevelName(betterLevel(lv1, lv2), lang),
				Reason:    rel1.Group + " + " + rel2.Group,
			})
		}

		// Sort by level, keep top 5
		sort.SliceStable(luckyDays, func(i, j int) bool {
			return levelRank(luckyDays[i].Level) < levelRank(luckyDays[j].Level)
		})
		if len(luckyDays) > 5 {
			luckyDays = luckyDays[:5]
		}

		actions = append(actions, PairLuckyAction{
			Action:    action.key,
			Name:      actionName(action, lang),
			LuckyDays: luckyDays,
		})
	}

	return &PairLuckyDaysResult{
		Person1: MansionBrief{Index: idx1, NameJP: m1.NameJP, Reading: m1.Reading, Yosei: m1.Yosei},
		Person2: MansionBrief{Index: idx2, NameJP: m2.NameJP, Reading: m2.Reading, Yosei: m2.Yosei},
		Compatibility: PairCompat{
			Relation: pairRel.Group,
			Level:    RelationLevelMap[pairRel.Group],
		},
		Actions: actions,
	}, nil
}

// --- internal helpers ---

type dayEval struct {
	date    time.Time
	dayIdx  int
	rel     engine.Relation
	level   string
	special *engine.SpecialDayType
	ryouhan bool
}

func findLuckyDaysForAction(evals []dayEval, action luckyAction, lang string) []LuckyDay {
	var result []LuckyDay
	for _, ev := range evals {
		rating := rateAction(ev.rel.Group, ev.level, action)
		if rating == "avoid" || rating == "fair" {
			continue
		}
		jpwd := engine.JPWeekday(ev.date)
		result = append(result, LuckyDay{
			Date:      ev.date.Format("2006-01-02"),
			Weekday:   WeekdayName(jpwd, lang),
			Level:     ev.level,
			LevelName: LevelName(ev.level, lang),
			Reason:    ev.rel.Group + " / " + ev.rel.Direction,
		})
	}
	// Sort best first, limit to 5
	sort.SliceStable(result, func(i, j int) bool {
		return levelRank(result[i].Level) < levelRank(result[j].Level)
	})
	if len(result) > 5 {
		result = result[:5]
	}
	return result
}

func rateAction(group, level string, action luckyAction) string {
	for _, bg := range action.badGroups {
		if group == bg {
			return "avoid"
		}
	}
	for _, gg := range action.goodGroups {
		if group == gg {
			if level == LevelDaikichi {
				return "excellent"
			}
			return "good"
		}
	}
	if level == LevelKyo {
		return "avoid"
	}
	return "fair"
}

func levelRank(level string) int {
	switch level {
	case LevelDaikichi:
		return 0
	case LevelKichi:
		return 1
	case LevelShokyo:
		return 2
	case LevelKyo:
		return 3
	default:
		return 4
	}
}

func betterLevel(a, b string) string {
	if levelRank(a) < levelRank(b) {
		return a
	}
	return b
}

func categoryName(cat, lang string) string {
	names := map[string]map[string]string{
		"career": {"zh": "事業", "ja": "キャリア", "en": "Career"},
		"social": {"zh": "社交", "ja": "社交", "en": "Social"},
		"health": {"zh": "健康", "ja": "健康", "en": "Health"},
		"move":   {"zh": "移動", "ja": "移動", "en": "Travel"},
	}
	k := langKey(lang)
	if m, ok := names[cat]; ok {
		return m[k]
	}
	return cat
}

func actionName(action luckyAction, lang string) string {
	// For now return zh name; can be extended with i18n
	return action.name
}
