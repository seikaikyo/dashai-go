package career

import (
	"sort"
	"sync"
	"time"

	"github.com/seikaikyo/dashai-go/internal/shukuyo/engine"
)

// BatchAnalyze performs career analysis for multiple companies concurrently.
func BatchAnalyze(req BatchRequest) (*BatchResponse, error) {
	birth, err := engine.ParseDate(req.BirthDate)
	if err != nil {
		return nil, err
	}

	ctx := DefaultContext(req.Context)
	lang := normalizeLang(req.Lang)
	year := req.Year
	if year == 0 {
		year = time.Now().Year()
	}

	// User info
	mansion := engine.GetMansion(birth)
	kuyou := engine.GetKuyouStar(birth, year)
	userIsBadYear := isBadYear(kuyou)

	user := UserSummary{
		BirthDate: req.BirthDate,
		Mansion: engine.PersonInfo{
			Date:    req.BirthDate,
			Mansion: mansion.Mansion.NameJP,
			Reading: mansion.Mansion.Reading,
			Yosei:   mansion.Mansion.Yosei,
			Index:   mansion.MansionIndex,
		},
		Kuyou: &kuyou,
	}

	// Concurrent company analysis
	type compResult struct {
		idx    int
		result BatchCompany
		err    error
	}

	results := make([]compResult, len(req.Companies))
	var wg sync.WaitGroup

	for i, comp := range req.Companies {
		wg.Add(1)
		go func(idx int, c CompanyEntry) {
			defer wg.Done()
			r, e := analyzeOneCompany(birth, c, ctx, lang, year, userIsBadYear, req.IncludeDeep)
			results[idx] = compResult{idx: idx, result: r, err: e}
		}(i, comp)
	}
	wg.Wait()

	// Collect results
	companies := make([]BatchCompany, 0, len(req.Companies))
	tierSummary := map[int]int{}
	drainSummary := map[string]int{}

	for _, r := range results {
		if r.err != nil {
			continue
		}
		companies = append(companies, r.result)
		tierSummary[r.result.Level.Tier]++
		drainSummary[r.result.Drain.Severity]++
	}

	// Sort by tier (best first)
	sort.Slice(companies, func(i, j int) bool {
		if companies[i].Level.Tier != companies[j].Level.Tier {
			return companies[i].Level.Tier < companies[j].Level.Tier
		}
		return DrainOrdinal(companies[i].Drain.Severity) < DrainOrdinal(companies[j].Drain.Severity)
	})

	// Strategic summary
	var strategic *StrategicSummary
	if len(companies) > 0 {
		strategic = buildStrategicSummary(companies, lang)
	}

	return &BatchResponse{
		User:             user,
		Companies:        companies,
		TierSummary:      tierSummary,
		DrainSummary:     drainSummary,
		StrategicSummary: strategic,
	}, nil
}

func analyzeOneCompany(birth time.Time, comp CompanyEntry, ctx ContextType, lang string, year int, userBadYear, includeDeep bool) (BatchCompany, error) {
	compDate, err := engine.ParseDate(comp.FoundingDate)
	if err != nil {
		return BatchCompany{}, err
	}

	compat := engine.Compatibility(birth, compDate)
	rel := compat.Relation

	level := CareerLevel(rel.Group)
	tier := CareerTier(rel.Group)
	drain := DrainLevel(rel.Group, rel.Direction)

	relations := LoadRelations(ctx, lang)
	careers := LoadCareer(ctx, lang)
	drainTexts := LoadDrainTexts(ctx, lang)

	relText := relations[rel.Group]
	dirRomaji := DirectionRomaji(rel.Direction)
	drainText := drainTexts[dirRomaji]
	careerText := careers[dirRomaji]

	crossRisk := CrossRisk(drain, userBadYear)

	var initiative *TextBlock
	if careerText.Headline != "" {
		initiative = &TextBlock{
			Headline: careerText.Headline,
			Detail:   careerText.ContextNote,
		}
	}

	bc := BatchCompany{
		ID:   comp.ID,
		Name: comp.Name,
		Relation: RelationDetail{
			Group:            rel.Group,
			Direction:        rel.Direction,
			InverseDirection: rel.InverseDirection,
			ForwardDistance:   rel.ForwardDistance,
			GroupName:        GroupName(rel.Group, lang),
			Description:      relText.Description,
		},
		Level: LevelInfo{
			Level:     level,
			LevelName: CareerLevelName(level, lang),
			Tier:      tier,
		},
		Drain: DrainInfo{
			Severity:    drain,
			Description: drainText.Relationship,
			LongTerm:    drainText.LongTermRisk,
			BadYearNote: drainText.BadYearImpact,
		},
		Initiative: initiative,
		CrossRisk:  crossRisk,
	}

	// Deep analysis: 7-year range (past 2 + future 5)
	if includeDeep {
		bc.DeepYears = buildDeepYears(birth, compDate, year)
	}

	return bc, nil
}

func buildDeepYears(birth, compDate time.Time, currentYear int) []YearCross {
	startYear := currentYear - 2
	endYear := currentYear + 5
	years := make([]YearCross, 0, endYear-startYear)

	for y := startYear; y < endYear; y++ {
		userKuyou := engine.GetKuyouStar(birth, y)
		compKuyou := engine.GetKuyouStar(compDate, y)

		userLv := kuyouToLevel(userKuyou)
		compLv := kuyouToLevel(compKuyou)

		risk := ""
		if userLv == "kyo" && compLv == "kyo" {
			risk = "danger"
		} else if userLv == "kyo" || compLv == "kyo" {
			risk = "warning"
		}

		years = append(years, YearCross{
			Year:         y,
			UserLevel:    userLv,
			CompanyLevel: compLv,
			Risk:         risk,
		})
	}
	return years
}

func isBadYear(kuyou engine.KuyouStar) bool {
	// Index 4 (五黄殺) is the worst year in nine-star cycle
	return kuyou.Index == 4
}

func kuyouToLevel(kuyou engine.KuyouStar) string {
	// Simplified mapping based on nine-star fortune
	switch kuyou.Index {
	case 4:
		return "kyo" // 五黄殺: worst
	case 2, 6:
		return "shokyo" // challenging
	case 0, 8:
		return "daikichi" // best
	default:
		return "kichi"
	}
}

func buildStrategicSummary(companies []BatchCompany, lang string) *StrategicSummary {
	categories := map[string][]string{
		"best_match":      {},
		"growth_potential": {},
		"safe_bet":        {},
		"watch_out":       {},
	}

	for _, c := range companies {
		switch {
		case c.Level.Tier == 1:
			categories["best_match"] = append(categories["best_match"], c.Name)
		case c.Level.Tier == 2:
			categories["growth_potential"] = append(categories["growth_potential"], c.Name)
		case c.Level.Tier == 3 && !IsHighDrain(c.Drain.Severity):
			categories["safe_bet"] = append(categories["safe_bet"], c.Name)
		default:
			categories["watch_out"] = append(categories["watch_out"], c.Name)
		}
	}

	topPick := ""
	reason := ""
	if len(companies) > 0 {
		topPick = companies[0].Name
		reason = companies[0].Relation.GroupName + " / " + companies[0].Drain.Severity + " drain"
	}

	return &StrategicSummary{
		TopPick:    topPick,
		Reason:     reason,
		Categories: categories,
	}
}
