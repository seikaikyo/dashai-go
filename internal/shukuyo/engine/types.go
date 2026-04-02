package engine

// Mansion represents a lunar mansion (nakshatra).
type Mansion struct {
	Index   int    `json:"index"`
	Name    string `json:"name"`
	NameJP  string `json:"name_jp"`
	Reading string `json:"reading"`
	Yosei   string `json:"yosei"`
}

// LunarDate holds a date in the Chinese lunar calendar.
type LunarDate struct {
	Year   int  `json:"year"`
	Month  int  `json:"month"`
	Day    int  `json:"day"`
	IsLeap bool `json:"is_leap"`
}

// MansionResult is the response for mansion index lookup.
type MansionResult struct {
	MansionIndex int       `json:"mansion_index"`
	Mansion      Mansion   `json:"mansion"`
	LunarDate    LunarDate `json:"lunar_date"`
	SolarDate    string    `json:"solar_date"`
}

// Relation describes the relationship between two mansions.
// All fields derived from T21n1299 original text.
type Relation struct {
	// Group is the paired relationship: eishin/yusui/ankai/kisei/mei/gyotai
	// (romaji of original kanji pairs: 栄親/友衰/安壊/危成/命/業胎)
	Group string `json:"group"`

	// Direction is the position name from person1's perspective.
	// One of: 命/栄/衰/安/危/成/壊/友/親/業/胎
	Direction string `json:"direction"`

	// InverseDirection is the position from person2's perspective.
	InverseDirection string `json:"inverse_direction"`

	// ForwardDistance is (idx2 - idx1) mod 27.
	ForwardDistance int `json:"forward_distance"`
}

// KuyouStar holds a nine-star year calculation result.
// Names, readings, and buddha from T21n1299 品四.
type KuyouStar struct {
	Index    int    `json:"index"`
	Name     string `json:"name"`
	Reading  string `json:"reading"`
	Yosei    string `json:"yosei,omitempty"`
	Buddha   string `json:"buddha"`
	KazoeAge int    `json:"kazoe_age"`
}

// CompatibilityResult is the response for compatibility calculation.
type CompatibilityResult struct {
	Person1  PersonInfo `json:"person1"`
	Person2  PersonInfo `json:"person2"`
	Relation Relation   `json:"relation"`
}

// PersonInfo holds mansion info for a person in compatibility.
type PersonInfo struct {
	Date    string `json:"date"`
	Mansion string `json:"mansion"`
	Reading string `json:"reading"`
	Yosei   string `json:"yosei"`
	Index   int    `json:"index"`
}

// SpecialDayType identifies a special day (kanro/kongou/rasetsu).
// From T21n1299 卷五.
type SpecialDayType struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Reading string `json:"reading"`
}

// RyouhanPeriod describes a ryouhan (inauspicious) period.
// From T21n1299 品三/品五.
type RyouhanPeriod struct {
	Active     bool `json:"active"`
	LunarMonth int  `json:"lunar_month"`
	StartDay   int  `json:"start_day"`
	EndDay     int  `json:"end_day"`
}

// SankiPosition describes where a day falls in the 27-day three-nine cycle.
// From T21n1299 品二.
type SankiPosition struct {
	PeriodIndex int    `json:"period_index"`
	PeriodName  string `json:"period_name"`
	DayInPeriod int    `json:"day_in_period"`
	DayType     string `json:"day_type"`
}
