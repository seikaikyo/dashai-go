package engine

import "time"

// Compatibility calculates the relationship between two people.
// Pure T21n1299: mansion indices → forward distance → position → group.
func Compatibility(date1, date2 time.Time) CompatibilityResult {
	idx1, _ := MansionIndexFromDate(date1)
	idx2, _ := MansionIndexFromDate(date2)

	m1 := Mansions27[idx1]
	m2 := Mansions27[idx2]
	rel := GetRelation(idx1, idx2)

	return CompatibilityResult{
		Person1: PersonInfo{
			Date:    date1.Format("2006-01-02"),
			Mansion: m1.NameJP,
			Reading: m1.Reading,
			Yosei:   m1.Yosei,
			Index:   idx1,
		},
		Person2: PersonInfo{
			Date:    date2.Format("2006-01-02"),
			Mansion: m2.NameJP,
			Reading: m2.Reading,
			Yosei:   m2.Yosei,
			Index:   idx2,
		},
		Relation: rel,
	}
}
