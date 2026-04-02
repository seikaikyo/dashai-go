package career

import (
	"sort"
	"sync"
	"time"

	"github.com/seikaikyo/dashai-go/internal/shukuyo/engine"
)

// HeadhunterMatch ranks companies for a candidate from a headhunter's perspective.
func HeadhunterMatch(req HeadhunterMatchRequest) (*HeadhunterMatchResponse, error) {
	candDate, err := engine.ParseDate(req.CandidateDate)
	if err != nil {
		return nil, err
	}

	lang := normalizeLang(req.Lang)
	ctx := ContextHeadhunter

	type rankResult struct {
		idx    int
		result HeadhunterRanking
		err    error
	}

	results := make([]rankResult, len(req.Companies))
	var wg sync.WaitGroup

	for i, comp := range req.Companies {
		wg.Add(1)
		go func(idx int, c CompanyEntry) {
			defer wg.Done()
			r, e := matchOneCompany(candDate, c, ctx, lang)
			results[idx] = rankResult{idx: idx, result: r, err: e}
		}(i, comp)
	}
	wg.Wait()

	rankings := make([]HeadhunterRanking, 0, len(req.Companies))
	for _, r := range results {
		if r.err != nil {
			continue
		}
		rankings = append(rankings, r.result)
	}

	// Sort by tier then drain
	sort.Slice(rankings, func(i, j int) bool {
		if rankings[i].Level.Tier != rankings[j].Level.Tier {
			return rankings[i].Level.Tier < rankings[j].Level.Tier
		}
		return DrainOrdinal(rankings[i].Drain.Severity) < DrainOrdinal(rankings[j].Drain.Severity)
	})

	return &HeadhunterMatchResponse{
		Rankings: rankings,
	}, nil
}

func matchOneCompany(candDate time.Time, comp CompanyEntry, ctx ContextType, lang string) (HeadhunterRanking, error) {
	compDate, err := engine.ParseDate(comp.FoundingDate)
	if err != nil {
		return HeadhunterRanking{}, err
	}

	compat := engine.Compatibility(candDate, compDate)
	rel := compat.Relation

	level := CareerLevel(rel.Group)
	tier := CareerTier(rel.Group)
	drain := DrainLevel(rel.Group, rel.Direction)

	// Load texts (fallback to employment if headhunter texts don't exist yet)
	relations := LoadRelations(ctx, lang)
	if len(relations) == 0 {
		relations = LoadRelations(ContextEmployment, lang)
	}
	drainTexts := LoadDrainTexts(ctx, lang)
	if len(drainTexts) == 0 {
		drainTexts = LoadDrainTexts(ContextEmployment, lang)
	}

	relText := relations[rel.Group]
	dirRomaji := DirectionRomaji(rel.Direction)
	drainText := drainTexts[dirRomaji]

	// Generate pitch texts
	pitchCandidate := buildCandidatePitch(rel.Group, relText, lang)
	pitchCompany := buildCompanyPitch(rel.Group, drain, relText, lang)

	return HeadhunterRanking{
		Company: comp,
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
		},
		PitchCandidate: pitchCandidate,
		PitchCompany:   pitchCompany,
	}, nil
}

func buildCandidatePitch(group string, relText RelationText, lang string) string {
	// Simple pitch based on relation group
	if len(relText.Tips) > 0 {
		return relText.Tips[0]
	}
	return relText.Description
}

func buildCompanyPitch(group, drain string, relText RelationText, lang string) string {
	if len(relText.GoodFor) > 0 {
		return relText.GoodFor[0]
	}
	return relText.Description
}
