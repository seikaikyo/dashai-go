package career

import (
	"github.com/seikaikyo/dashai-go/internal/shukuyo/engine"
)

// Analyze performs a single career compatibility analysis.
func Analyze(req AnalyzeRequest) (*AnalyzeResponse, error) {
	birth, err := engine.ParseDate(req.BirthDate)
	if err != nil {
		return nil, err
	}
	company, err := engine.ParseDate(req.CompanyDate)
	if err != nil {
		return nil, err
	}

	ctx := DefaultContext(req.Context)
	lang := normalizeLang(req.Lang)

	// Engine calculation
	compat := engine.Compatibility(birth, company)
	rel := compat.Relation

	// Career-specific level and drain
	level := CareerLevel(rel.Group)
	tier := CareerTier(rel.Group)
	drain := DrainLevel(rel.Group, rel.Direction)

	// Load context-specific texts
	relations := LoadRelations(ctx, lang)
	careers := LoadCareer(ctx, lang)
	drainTexts := LoadDrainTexts(ctx, lang)
	redFlags := LoadRedFlags(ctx, lang)

	// Build relation detail
	relText := relations[rel.Group]
	relDetail := RelationDetail{
		Group:            rel.Group,
		Direction:        rel.Direction,
		InverseDirection: rel.InverseDirection,
		ForwardDistance:   rel.ForwardDistance,
		GroupName:        GroupName(rel.Group, lang),
		Description:      relText.Description,
	}

	// Build drain info
	dirRomaji := DirectionRomaji(rel.Direction)
	drainText := drainTexts[dirRomaji]
	drainInfo := DrainInfo{
		Severity:    drain,
		Description: drainText.Relationship,
		LongTerm:    drainText.LongTermRisk,
		BadYearNote: drainText.BadYearImpact,
	}

	// Build career advice
	careerText := careers[dirRomaji]
	var careerAdvice *CareerAdvice
	if careerText.Headline != "" {
		careerAdvice = &CareerAdvice{
			Headline: careerText.Headline,
			Detail:   careerText.Detail,
			Do:       careerText.Do,
			Avoid:    careerText.Avoid,
		}
	}

	// Build initiative (from career text context_note)
	var initiative *TextBlock
	if careerText.ContextNote != "" {
		initiative = &TextBlock{
			Headline: careerText.Headline,
			Detail:   careerText.ContextNote,
		}
	}

	// Build red flags
	flags := redFlags[rel.Group]

	// Build gap guidance from relation advice
	var gapGuidance *TextBlock
	if relText.Advice != "" {
		gapGuidance = &TextBlock{
			Headline: relText.Advice,
		}
	}

	return &AnalyzeResponse{
		Person1: compat.Person1,
		Person2: compat.Person2,
		Relation: relDetail,
		Level: LevelInfo{
			Level:     level,
			LevelName: CareerLevelName(level, lang),
			Tier:      tier,
		},
		Drain:       drainInfo,
		Initiative:  initiative,
		RedFlags:    flags,
		GapGuidance: gapGuidance,
		Career:      careerAdvice,
	}, nil
}
