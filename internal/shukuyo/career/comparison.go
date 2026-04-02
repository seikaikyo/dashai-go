package career

import (
	"github.com/seikaikyo/dashai-go/internal/shukuyo/engine"
)

// Compare performs multi-year comparison between companies.
func Compare(req ComparisonRequest) (*ComparisonResponse, error) {
	birth, err := engine.ParseDate(req.BirthDate)
	if err != nil {
		return nil, err
	}

	ctx := DefaultContext(req.Context)
	lang := normalizeLang(req.Lang)

	mansion := engine.GetMansion(birth)
	user := UserSummary{
		BirthDate: req.BirthDate,
		Mansion: engine.PersonInfo{
			Date:    req.BirthDate,
			Mansion: mansion.Mansion.NameJP,
			Reading: mansion.Mansion.Reading,
			Yosei:   mansion.Mansion.Yosei,
			Index:   mansion.MansionIndex,
		},
	}

	companies := make([]ComparisonCompany, 0, len(req.Companies))

	for _, comp := range req.Companies {
		compDate, err := engine.ParseDate(comp.FoundingDate)
		if err != nil {
			continue
		}

		compat := engine.Compatibility(birth, compDate)
		rel := compat.Relation

		level := CareerLevel(rel.Group)
		tier := CareerTier(rel.Group)
		drain := DrainLevel(rel.Group, rel.Direction)

		relations := LoadRelations(ctx, lang)
		drainTexts := LoadDrainTexts(ctx, lang)

		relText := relations[rel.Group]
		dirRomaji := DirectionRomaji(rel.Direction)
		drainText := drainTexts[dirRomaji]

		// Build decade chart
		decadeChart := make([]YearCross, 0, req.EndYear-req.StartYear+1)
		bothGood := 0
		bothBad := 0

		for y := req.StartYear; y <= req.EndYear; y++ {
			userKuyou := engine.GetKuyouStar(birth, y)
			compKuyou := engine.GetKuyouStar(compDate, y)

			userLv := kuyouToLevel(userKuyou)
			compLv := kuyouToLevel(compKuyou)

			risk := ""
			if userLv == "kyo" && compLv == "kyo" {
				risk = "danger"
				bothBad++
			} else if userLv == "kyo" || compLv == "kyo" {
				risk = "warning"
			}

			if (userLv == "daikichi" || userLv == "kichi") && (compLv == "daikichi" || compLv == "kichi") {
				bothGood++
			}

			decadeChart = append(decadeChart, YearCross{
				Year:         y,
				UserLevel:    userLv,
				CompanyLevel: compLv,
				Risk:         risk,
			})
		}

		companies = append(companies, ComparisonCompany{
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
			DecadeChart:   decadeChart,
			BothGoodYears: bothGood,
			BothBadYears:  bothBad,
		})
	}

	return &ComparisonResponse{
		User:      user,
		Companies: companies,
	}, nil
}
