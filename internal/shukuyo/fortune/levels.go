// Package fortune implements the modern interpretation layer on top of the
// T21n1299 engine. All level mappings and fortune determinations in this
// package are modern reconstructions, not original sutra text.
package fortune

// --- Level system (modern interpretation) ---

// Level constants
const (
	LevelDaikichi = "daikichi"
	LevelKichi    = "kichi"
	LevelShokyo   = "shokyo"
	LevelKyo      = "kyo"
)

// levelOrder defines levels from best to worst (index 0 = best).
var levelOrder = [4]string{LevelDaikichi, LevelKichi, LevelShokyo, LevelKyo}

// levelIndex maps level string to ordinal (0=best, 3=worst).
var levelIndex = map[string]int{
	LevelDaikichi: 0, LevelKichi: 1, LevelShokyo: 2, LevelKyo: 3,
}

// RelationLevelMap maps relation group to base fortune level.
// Modern interpretation based on T21n1299 三九秘法 day descriptions.
var RelationLevelMap = map[string]string{
	"eishin": LevelDaikichi, // 栄親: 諸吉事並大吉 (p.397c)
	"yusui":  LevelKichi,    // 友衰: 宜結交 (p.391b) + 唯宜療病 (p.398a)
	"kisei":  LevelShokyo,   // 危成: 社交吉但行動受限
	"ankai":  LevelKyo,      // 安壊: 壊日餘並不堪 (p.398a)
	"gyotai": LevelShokyo,   // 業胎: 所作不成就 (p.397c) + 不宜百事 (p.391b)
	"mei":    LevelShokyo,   // 命: 不宜舉動百事 (p.397c)
}

// DirectionLevelOverride overrides level for specific directions within a group.
// 成日 has positive sutra text (修道學問吉), 安日 has limited positive scope.
var DirectionLevelOverride = map[string]string{
	"成": LevelKichi,  // 宜修道學問、作諸成就法並吉 (p.398a)
	"安": LevelShokyo, // 移徙吉 but limited scope (p.397c)
}

// RyouhanLevelFlip is the symmetric level inversion during ryouhan periods.
// 凌犯期間 reverses fortune: good becomes bad, bad becomes good.
var RyouhanLevelFlip = map[string]string{
	LevelDaikichi: LevelKyo,
	LevelKichi:    LevelShokyo,
	LevelShokyo:   LevelKichi,
	LevelKyo:      LevelDaikichi,
}

// LevelNames maps level to display name by language key (zh/ja/en).
var LevelNames = map[string]map[string]string{
	LevelDaikichi: {"zh": "大吉", "ja": "大吉", "en": "Great Fortune"},
	LevelKichi:    {"zh": "吉", "ja": "吉", "en": "Good Fortune"},
	LevelShokyo:   {"zh": "小凶", "ja": "小凶", "en": "Minor Caution"},
	LevelKyo:      {"zh": "凶", "ja": "凶", "en": "Caution"},
}

// LevelName returns the display name for a level in the given language.
func LevelName(level, lang string) string {
	key := "zh"
	switch lang {
	case "ja":
		key = "ja"
	case "en":
		key = "en"
	}
	if names, ok := LevelNames[level]; ok {
		return names[key]
	}
	return level
}

// DetermineLevel calculates the daily fortune level.
// Returns (finalLevel, baseLevel).
func DetermineLevel(group, direction string, ryouhanActive bool, specialDayType string) (string, string) {
	// Step 1: Base level from relation group
	base := RelationLevelMap[group]
	if base == "" {
		base = LevelShokyo
	}

	// Step 2: Direction override
	if override, ok := DirectionLevelOverride[direction]; ok {
		base = override
	}

	level := base

	// Step 3: Ryouhan flip
	if ryouhanActive {
		if flipped, ok := RyouhanLevelFlip[level]; ok {
			level = flipped
		}
	}

	// Step 4: Special day shift
	if specialDayType != "" {
		level = shiftBySpecialDay(level, specialDayType, ryouhanActive)
	}

	return level, base
}

// shiftBySpecialDay adjusts level based on special day type.
// Normal: kanro/kongou shift up, rasetsu shifts down.
// Ryouhan: reversed (kanro/kongou shift down, rasetsu shifts up).
func shiftBySpecialDay(level, sdType string, ryouhanActive bool) string {
	idx, ok := levelIndex[level]
	if !ok {
		return level
	}

	shift := 0
	switch sdType {
	case "kanro", "kongou":
		if ryouhanActive {
			shift = 1 // shift down (worse) during ryouhan
		} else {
			shift = -1 // shift up (better) normally
		}
	case "rasetsu":
		if ryouhanActive {
			shift = -1 // shift up (better) during ryouhan
		} else {
			shift = 1 // shift down (worse) normally
		}
	}

	newIdx := idx + shift
	if newIdx < 0 {
		newIdx = 0
	}
	if newIdx > 3 {
		newIdx = 3
	}
	return levelOrder[newIdx]
}

// AggregateLevel averages multiple daily levels into one.
// Used for monthly trend calculation.
func AggregateLevel(levels []string) string {
	if len(levels) == 0 {
		return LevelShokyo
	}
	sum := 0
	for _, lv := range levels {
		// ordinal: daikichi=3, kichi=2, shokyo=1, kyo=0
		sum += 3 - levelIndex[lv]
	}
	avg := float64(sum) / float64(len(levels))
	nearest := int(avg + 0.5)
	if nearest > 3 {
		nearest = 3
	}
	if nearest < 0 {
		nearest = 0
	}
	return levelOrder[3-nearest]
}
