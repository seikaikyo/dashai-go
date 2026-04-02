package career

// Drain severity constants, ordered from least to most draining.
const (
	DrainNone        = "none"
	DrainMinimal     = "minimal"
	DrainLow         = "low"
	DrainModerate    = "moderate"
	DrainModerateHigh = "moderate_high"
	DrainHigh        = "high"
	DrainVeryHigh    = "very_high"
)

// drainSeverityOrder maps severity to ordinal (0=none, 6=very_high).
var drainSeverityOrder = map[string]int{
	DrainNone:         0,
	DrainMinimal:      1,
	DrainLow:          2,
	DrainModerate:     3,
	DrainModerateHigh: 4,
	DrainHigh:         5,
	DrainVeryHigh:     6,
}

// drainLevelMap maps (group, direction) to drain severity.
// Matches Python DRAIN_LEVEL_MAP exactly for backward compatibility.
var drainLevelMap = map[[2]string]string{
	{"mei", "命"}:    DrainNone,
	{"gyotai", "業"}: DrainLow,
	{"gyotai", "胎"}: DrainMinimal,
	{"eishin", "栄"}: DrainLow,
	{"eishin", "親"}: DrainMinimal,
	{"yusui", "友"}:  DrainLow,
	{"yusui", "衰"}:  DrainHigh,
	{"ankai", "安"}:  DrainModerate,
	{"ankai", "壊"}:  DrainVeryHigh,
	{"kisei", "危"}:  DrainModerateHigh,
	{"kisei", "成"}:  DrainLow,
}

// DrainLevel returns the drain severity for a (group, direction) pair.
func DrainLevel(group, direction string) string {
	if lv, ok := drainLevelMap[[2]string{group, direction}]; ok {
		return lv
	}
	return DrainLow
}

// DrainOrdinal returns the numeric severity (0-6) for comparison.
func DrainOrdinal(severity string) int {
	if ord, ok := drainSeverityOrder[severity]; ok {
		return ord
	}
	return 2
}

// IsHighDrain returns true if severity >= moderate_high (ordinal >= 4).
func IsHighDrain(severity string) bool {
	return DrainOrdinal(severity) >= 4
}

// CrossRisk evaluates the cross-risk when combining drain with bad-year status.
// Returns "danger" for high drain + bad year, "warning" for moderate drain + bad year.
func CrossRisk(drainSeverity string, isBadYear bool) string {
	if !isBadYear {
		return ""
	}
	ord := DrainOrdinal(drainSeverity)
	if ord >= 5 { // high or very_high
		return "danger"
	}
	if ord >= 3 { // moderate or moderate_high
		return "warning"
	}
	return ""
}
