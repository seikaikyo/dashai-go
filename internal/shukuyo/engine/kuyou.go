package engine

import "time"

// GetKuyouStar calculates the nine-star (kuyou) for a given birth date and year.
// T21n1299 品四: kazoe-doshi (counting age) method, 9-year cycle.
func GetKuyouStar(birthDate time.Time, year int) KuyouStar {
	kazoeAge := year - birthDate.Year() + 1
	starIndex := (kazoeAge - 1) % 9
	if starIndex < 0 {
		starIndex += 9
	}

	def := KuyouStarDefs[starIndex]
	return KuyouStar{
		Index:    starIndex,
		Name:     def.Name,
		Reading:  def.Reading,
		Yosei:    def.Yosei,
		Buddha:   def.Buddha,
		KazoeAge: kazoeAge,
	}
}
