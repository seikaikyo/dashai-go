// Package career implements the contextual career compatibility system.
// Level mappings here are for CAREER compatibility, distinct from fortune/levels.go
// which handles daily fortune. Same relation group maps to different levels because
// the interpretive context differs (e.g. gyotai = karma bond = good for relationships,
// but gyotai days = unlucky per sutra text).
package career

// Career compatibility level constants (same strings, different mapping).
const (
	LevelDaikichi = "daikichi"
	LevelKichi    = "kichi"
	LevelShokyo   = "shokyo"
	LevelKyo      = "kyo"
)

// CareerLevelMap maps relation group to career compatibility level.
// Differs from fortune daily levels:
//   - gyotai: kichi (career) vs shokyo (daily) -- karma bond = good compatibility
//   - mei:    kichi (career) vs shokyo (daily) -- mirror = neutral-positive compatibility
//   - yusui:  shokyo (career) vs kichi (daily) -- comfortable but stagnation risk
var CareerLevelMap = map[string]string{
	"eishin": LevelDaikichi, // tier 1: mutual elevation
	"gyotai": LevelKichi,    // tier 2: past-life karma bond
	"mei":    LevelKichi,    // tier 2: mirror reflection
	"yusui":  LevelShokyo,   // tier 3: comfortable but draining
	"kisei":  LevelShokyo,   // tier 3: complementary but friction
	"ankai":  LevelKyo,      // tier 4: power imbalance
}

// CareerTierMap maps relation group to tier number (1=best, 4=worst).
var CareerTierMap = map[string]int{
	"eishin": 1,
	"gyotai": 2,
	"mei":    2,
	"yusui":  3,
	"kisei":  3,
	"ankai":  4,
}

// CareerLevelNames maps level to display name by language.
var CareerLevelNames = map[string]map[string]string{
	LevelDaikichi: {"zh": "大吉", "ja": "大吉", "en": "Excellent"},
	LevelKichi:    {"zh": "吉", "ja": "吉", "en": "Good"},
	LevelShokyo:   {"zh": "小凶", "ja": "小凶", "en": "Fair"},
	LevelKyo:      {"zh": "凶", "ja": "凶", "en": "Caution"},
}

// GroupNames maps relation group to display name by language.
var GroupNames = map[string]map[string]string{
	"eishin": {"zh": "栄親", "ja": "栄親", "en": "Eishin"},
	"gyotai": {"zh": "業胎", "ja": "業胎", "en": "Gyotai"},
	"mei":    {"zh": "命", "ja": "命", "en": "Mei"},
	"yusui":  {"zh": "友衰", "ja": "友衰", "en": "Yusui"},
	"kisei":  {"zh": "危成", "ja": "危成", "en": "Kisei"},
	"ankai":  {"zh": "安壊", "ja": "安壊", "en": "Ankai"},
}

// CareerLevel returns the career compatibility level for a relation group.
func CareerLevel(group string) string {
	if lv, ok := CareerLevelMap[group]; ok {
		return lv
	}
	return LevelShokyo
}

// CareerTier returns the tier number (1-4) for a relation group.
func CareerTier(group string) int {
	if t, ok := CareerTierMap[group]; ok {
		return t
	}
	return 3
}

// CareerLevelName returns the display name for a level in the given language.
func CareerLevelName(level, lang string) string {
	key := langKey(lang)
	if names, ok := CareerLevelNames[level]; ok {
		return names[key]
	}
	return level
}

// GroupName returns the display name for a relation group in the given language.
func GroupName(group, lang string) string {
	key := langKey(lang)
	if names, ok := GroupNames[group]; ok {
		return names[key]
	}
	return group
}

func langKey(lang string) string {
	switch lang {
	case "ja":
		return "ja"
	case "en":
		return "en"
	default:
		return "zh"
	}
}
