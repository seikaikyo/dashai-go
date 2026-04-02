package career

import (
	"sort"
	"sync"

	"github.com/seikaikyo/dashai-go/internal/shukuyo/engine"
)

// TeamMatrix builds a compatibility matrix for all candidates against a company
// and against each other.
func TeamMatrix(req TeamMatrixRequest) (*TeamMatrixResponse, error) {
	compDate, err := engine.ParseDate(req.CompanyDate)
	if err != nil {
		return nil, err
	}

	lang := normalizeLang(req.Lang)
	n := len(req.Candidates)

	// Parse all candidate birth dates
	type parsedCandidate struct {
		entry    CandidateEntry
		birthIdx int
		info     engine.PersonInfo
	}

	candidates := make([]parsedCandidate, 0, n)
	for _, c := range req.Candidates {
		bd, err := engine.ParseDate(c.BirthDate)
		if err != nil {
			continue
		}
		idx, _ := engine.MansionIndexFromDate(bd)
		m := engine.Mansions27[idx]
		candidates = append(candidates, parsedCandidate{
			entry:    c,
			birthIdx: idx,
			info: engine.PersonInfo{
				Date:    c.BirthDate,
				Mansion: m.NameJP,
				Reading: m.Reading,
				Yosei:   m.Yosei,
				Index:   idx,
			},
		})
	}

	n = len(candidates)
	compIdx, _ := engine.MansionIndexFromDate(compDate)

	// Build N x N matrix concurrently
	matrix := make([][]MatrixCell, n)
	for i := range matrix {
		matrix[i] = make([]MatrixCell, n)
	}

	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(row int) {
			defer wg.Done()
			for col := 0; col < n; col++ {
				if row == col {
					matrix[row][col] = MatrixCell{Group: "self", Direction: "-", Level: "-", Drain: "-"}
					continue
				}
				rel := engine.GetRelation(candidates[row].birthIdx, candidates[col].birthIdx)
				matrix[row][col] = MatrixCell{
					Group:     rel.Group,
					Direction: rel.Direction,
					Level:     CareerLevel(rel.Group),
					Drain:     DrainLevel(rel.Group, rel.Direction),
				}
			}
		}(i)
	}
	wg.Wait()

	// Members list
	members := make([]engine.PersonInfo, n)
	for i, c := range candidates {
		members[i] = c.info
	}

	// Rankings: candidate vs company + average team fit
	rankings := make([]TeamRanking, n)
	for i, c := range candidates {
		compRel := engine.GetRelation(c.birthIdx, compIdx)
		level := CareerLevel(compRel.Group)
		tier := CareerTier(compRel.Group)
		drain := DrainLevel(compRel.Group, compRel.Direction)

		drainTexts := LoadDrainTexts(ContextHR, lang)
		dirRomaji := DirectionRomaji(compRel.Direction)
		drainText := drainTexts[dirRomaji]

		// Team fit: average level score against all other candidates
		fit := teamFitScore(i, matrix, n)

		rankings[i] = TeamRanking{
			ID:   c.entry.ID,
			Name: c.entry.Name,
			Level: LevelInfo{
				Level:     level,
				LevelName: CareerLevelName(level, lang),
				Tier:      tier,
			},
			Drain: DrainInfo{
				Severity:    drain,
				Description: drainText.Relationship,
			},
			TeamFit: fit,
		}
	}

	// Sort rankings: by tier, then team fit (descending)
	sort.Slice(rankings, func(i, j int) bool {
		if rankings[i].Level.Tier != rankings[j].Level.Tier {
			return rankings[i].Level.Tier < rankings[j].Level.Tier
		}
		return rankings[i].TeamFit > rankings[j].TeamFit
	})

	return &TeamMatrixResponse{
		Matrix:   matrix,
		Members:  members,
		Rankings: rankings,
	}, nil
}

func teamFitScore(candidateIdx int, matrix [][]MatrixCell, n int) float64 {
	if n <= 1 {
		return 1.0
	}

	levelScores := map[string]float64{
		LevelDaikichi: 1.0,
		LevelKichi:    0.75,
		LevelShokyo:   0.5,
		LevelKyo:      0.25,
	}

	sum := 0.0
	count := 0
	for j := 0; j < n; j++ {
		if j == candidateIdx {
			continue
		}
		cell := matrix[candidateIdx][j]
		if score, ok := levelScores[cell.Level]; ok {
			sum += score
			count++
		}
	}

	if count == 0 {
		return 0.5
	}
	return sum / float64(count)
}
