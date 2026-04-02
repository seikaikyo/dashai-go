package engine

import (
	"fmt"
	"time"

	"github.com/6tail/lunar-go/calendar"
)

// SolarToLunar converts a Gregorian date to a Chinese lunar date.
// In lunar-go, leap months have negative month values (e.g. -4 = leap month 4).
func SolarToLunar(t time.Time) LunarDate {
	solar := calendar.NewSolar(t.Year(), int(t.Month()), t.Day(), 0, 0, 0)
	lunar := solar.GetLunar()
	month := lunar.GetMonth()
	isLeap := month < 0
	if isLeap {
		month = -month
	}
	return LunarDate{
		Year:   lunar.GetYear(),
		Month:  month,
		Day:    lunar.GetDay(),
		IsLeap: isLeap,
	}
}

// LunarToSolar converts a Chinese lunar date to Gregorian. Returns zero time on error.
func LunarToSolar(year, month, day int) (time.Time, error) {
	defer func() {
		recover() // lunar-go may panic on invalid dates
	}()
	lunar := calendar.NewLunar(year, month, day, 0, 0, 0)
	solar := lunar.GetSolar()
	return time.Date(solar.GetYear(), time.Month(solar.GetMonth()), solar.GetDay(), 0, 0, 0, 0, time.UTC), nil
}

// MansionIndex calculates the natal mansion index (0-26) from a lunar month and day.
// Uses the bouchutsureki (month mansion almanac): each month starts at a fixed mansion,
// then advances one mansion per day.
func MansionIndex(lunarMonth, lunarDay int) int {
	month := lunarMonth
	if month < 1 || month > 12 {
		month = 1
	}
	start := MonthStartMansion[month]
	return (start + lunarDay - 1) % 27
}

// MansionIndexFromDate calculates the natal mansion index from a Gregorian birth date.
func MansionIndexFromDate(birthDate time.Time) (int, LunarDate) {
	ld := SolarToLunar(birthDate)
	idx := MansionIndex(ld.Month, ld.Day)
	return idx, ld
}

// DayMansionIndex calculates the day mansion index for a given solar date.
// Same algorithm as natal mansion: solar -> lunar -> mansion lookup.
func DayMansionIndex(solarDate time.Time) int {
	ld := SolarToLunar(solarDate)
	return MansionIndex(ld.Month, ld.Day)
}

// GetMansion returns full mansion info for a birth date.
func GetMansion(birthDate time.Time) MansionResult {
	idx, ld := MansionIndexFromDate(birthDate)
	m := Mansions27[idx]
	return MansionResult{
		MansionIndex: idx,
		Mansion:      m,
		LunarDate:    ld,
		SolarDate:    birthDate.Format("2006-01-02"),
	}
}

// JPWeekday converts Go's time.Weekday to Japanese weekday (0=Sun, 1=Mon, ..., 6=Sat).
func JPWeekday(t time.Time) int {
	return int(t.Weekday())
}

// CheckSpecialDay checks if a date + mansion index is a special day (kanro/kongou/rasetsu).
func CheckSpecialDay(solarDate time.Time, dayMansionIdx int) *SpecialDayType {
	jpwd := JPWeekday(solarDate)
	key := specialDayKey{jpwd, dayMansionIdx}
	sdType, ok := SpecialDayMap[key]
	if !ok {
		return nil
	}
	info := SpecialDayInfo[sdType]
	return &info
}

// CheckRyouhanPeriod checks if a date falls within a ryouhan (inauspicious) period.
func CheckRyouhanPeriod(solarDate time.Time) *RyouhanPeriod {
	ld := SolarToLunar(solarDate)

	// Get the solar date of the 1st day of this lunar month
	firstDaySolar, err := LunarToSolar(ld.Year, ld.Month, 1)
	if err != nil {
		return nil
	}

	// Japanese weekday of the 1st day
	jpwd := JPWeekday(firstDaySolar)

	key := ryouhanKey{ld.Month, jpwd}
	period, ok := RyouhanMap[key]
	if !ok {
		return nil
	}

	startDay, endDay := period[0], period[1]
	if ld.Day < startDay || ld.Day > endDay {
		return nil
	}

	return &RyouhanPeriod{
		Active:     true,
		LunarMonth: ld.Month,
		StartDay:   startDay,
		EndDay:     endDay,
	}
}

// GetSankiPosition calculates where a day falls in the 27-day three-nine cycle
// relative to the person's birth mansion.
func GetSankiPosition(birthMansionIdx, dayMansionIdx int) SankiPosition {
	dist := (dayMansionIdx - birthMansionIdx + 27) % 27
	periodIdx := dist / 9
	dayInPeriod := dist%9 + 1

	var dayType string
	if dayInPeriod == 1 {
		dayType = SankiPeriodStartNames[periodIdx]
	} else {
		dayType = SankiDayTypeShared[dayInPeriod]
	}

	return SankiPosition{
		PeriodIndex: periodIdx,
		PeriodName:  SankiPeriodNames[periodIdx],
		DayInPeriod: dayInPeriod,
		DayType:     dayType,
	}
}

// ParseDate parses a date string in "2006-01-02" format.
func ParseDate(s string) (time.Time, error) {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err)
	}
	return t, nil
}
