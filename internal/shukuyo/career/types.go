package career

import "github.com/seikaikyo/dashai-go/internal/shukuyo/engine"

// ContextType identifies one of six career analysis scenarios.
type ContextType string

const (
	ContextEmployment  ContextType = "employment"
	ContextConsulting  ContextType = "consulting"
	ContextOutsourcing ContextType = "outsourcing"
	ContextB2B         ContextType = "b2b"
	ContextHR          ContextType = "hr"
	ContextHeadhunter  ContextType = "headhunter"
)

// --- Request types ---

// AnalyzeRequest is the input for single career compatibility analysis.
type AnalyzeRequest struct {
	BirthDate   string      `json:"birth_date"`
	CompanyDate string      `json:"company_date"`
	Context     ContextType `json:"context"`
	Lang        string      `json:"lang"`
}

// CompanyEntry identifies a company for batch operations.
type CompanyEntry struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	FoundingDate string `json:"founding_date"`
}

// BatchRequest is the input for batch career analysis.
type BatchRequest struct {
	BirthDate   string         `json:"birth_date"`
	Year        int            `json:"year"`
	Companies   []CompanyEntry `json:"companies"`
	Context     ContextType    `json:"context"`
	Lang        string         `json:"lang"`
	IncludeDeep bool           `json:"include_deep"`
}

// ComparisonRequest is the input for company comparison.
type ComparisonRequest struct {
	BirthDate  string         `json:"birth_date"`
	Companies  []CompanyEntry `json:"companies"`
	StartYear  int            `json:"start_year"`
	EndYear    int            `json:"end_year"`
	Context    ContextType    `json:"context"`
	Lang       string         `json:"lang"`
}

// InterviewDatesRequest is the input for interview auspicious dates.
type InterviewDatesRequest struct {
	BirthDate   string `json:"birth_date"`
	CompanyDate string `json:"company_date"`
	Days        int    `json:"days"`
	Lang        string `json:"lang"`
}

// CandidateEntry identifies a candidate for team/headhunter operations.
type CandidateEntry struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	BirthDate string `json:"birth_date"`
}

// TeamMatrixRequest is the input for HR team compatibility matrix.
type TeamMatrixRequest struct {
	CompanyDate string           `json:"company_date"`
	Candidates  []CandidateEntry `json:"candidates"`
	Lang        string           `json:"lang"`
}

// HeadhunterMatchRequest is the input for headhunter matching.
type HeadhunterMatchRequest struct {
	CandidateDate string         `json:"candidate_date"`
	Companies     []CompanyEntry `json:"companies"`
	Lang          string         `json:"lang"`
}

// --- Response types ---

// AnalyzeResponse is the output for single career analysis.
type AnalyzeResponse struct {
	Person1     engine.PersonInfo `json:"person1"`
	Person2     engine.PersonInfo `json:"person2"`
	Relation    RelationDetail    `json:"relation"`
	Level       LevelInfo         `json:"level"`
	Drain       DrainInfo         `json:"drain"`
	Initiative  *TextBlock        `json:"initiative,omitempty"`
	RedFlags    []string          `json:"red_flags,omitempty"`
	GapGuidance *TextBlock        `json:"gap_guidance,omitempty"`
	Career      *CareerAdvice     `json:"career,omitempty"`
}

// RelationDetail extends engine.Relation with display names.
type RelationDetail struct {
	Group            string `json:"group"`
	Direction        string `json:"direction"`
	InverseDirection string `json:"inverse_direction"`
	ForwardDistance   int    `json:"forward_distance"`
	GroupName        string `json:"group_name"`
	Description      string `json:"description"`
}

// LevelInfo holds the career compatibility level.
type LevelInfo struct {
	Level     string `json:"level"`
	LevelName string `json:"level_name"`
	Tier      int    `json:"tier"`
}

// DrainInfo holds the energy drain assessment.
type DrainInfo struct {
	Severity    string `json:"severity"`
	Description string `json:"description"`
	LongTerm    string `json:"long_term_risk,omitempty"`
	BadYearNote string `json:"bad_year_impact,omitempty"`
}

// TextBlock is a reusable text container.
type TextBlock struct {
	Headline string `json:"headline"`
	Detail   string `json:"detail,omitempty"`
}

// CareerAdvice holds career-specific guidance for a direction.
type CareerAdvice struct {
	Headline string   `json:"headline"`
	Detail   string   `json:"detail"`
	Do       []string `json:"do,omitempty"`
	Avoid    []string `json:"avoid,omitempty"`
}

// --- Batch response ---

// BatchResponse is the output for batch career analysis.
type BatchResponse struct {
	User             UserSummary        `json:"user"`
	Companies        []BatchCompany     `json:"companies"`
	TierSummary      map[int]int        `json:"tier_summary"`
	DrainSummary     map[string]int     `json:"drain_summary"`
	StrategicSummary *StrategicSummary  `json:"strategic_summary,omitempty"`
}

// UserSummary holds the requesting user's basic info.
type UserSummary struct {
	BirthDate string            `json:"birth_date"`
	Mansion   engine.PersonInfo `json:"mansion"`
	Kuyou     *engine.KuyouStar `json:"kuyou,omitempty"`
}

// BatchCompany is one company result in a batch.
type BatchCompany struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Relation    RelationDetail `json:"relation"`
	Level       LevelInfo      `json:"level"`
	Drain       DrainInfo      `json:"drain"`
	Initiative  *TextBlock     `json:"initiative,omitempty"`
	CrossRisk   string         `json:"cross_risk,omitempty"`
	DeepYears   []YearCross    `json:"deep_years,omitempty"`
}

// YearCross is a single-year cross-risk entry for deep analysis.
type YearCross struct {
	Year         int    `json:"year"`
	UserLevel    string `json:"user_level"`
	CompanyLevel string `json:"company_level"`
	Risk         string `json:"risk"`
}

// StrategicSummary is the top-level recommendation.
type StrategicSummary struct {
	TopPick       string   `json:"top_pick"`
	Reason        string   `json:"reason"`
	Categories    map[string][]string `json:"categories"`
}

// --- Comparison response ---

// ComparisonResponse is the output for company comparison.
type ComparisonResponse struct {
	User      UserSummary         `json:"user"`
	Companies []ComparisonCompany `json:"companies"`
	Verdict   string              `json:"verdict,omitempty"`
}

// ComparisonCompany is one company in a comparison.
type ComparisonCompany struct {
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	Relation      RelationDetail `json:"relation"`
	Level         LevelInfo      `json:"level"`
	Drain         DrainInfo      `json:"drain"`
	DecadeChart   []YearCross    `json:"decade_chart"`
	BothGoodYears int            `json:"both_good_years"`
	BothBadYears  int            `json:"both_bad_years"`
}

// --- Interview dates response ---

// InterviewDatesResponse is the output for interview auspicious dates.
type InterviewDatesResponse struct {
	Target       InterviewDay   `json:"target"`
	Alternatives []InterviewDay `json:"alternatives"`
}

// InterviewDay is a single day evaluation.
type InterviewDay struct {
	Date       string `json:"date"`
	Weekday    string `json:"weekday"`
	Relation   string `json:"relation"`
	Level      string `json:"level"`
	LevelName  string `json:"level_name"`
	Verdict    string `json:"verdict"`
	SpecialDay string `json:"special_day,omitempty"`
	Ryouhan    bool   `json:"ryouhan"`
}

// --- Team matrix response ---

// TeamMatrixResponse is the output for HR team matrix.
type TeamMatrixResponse struct {
	Matrix   [][]MatrixCell      `json:"matrix"`
	Members  []engine.PersonInfo `json:"members"`
	Rankings []TeamRanking       `json:"rankings"`
}

// MatrixCell is one cell in the team compatibility matrix.
type MatrixCell struct {
	Group     string `json:"group"`
	Direction string `json:"direction"`
	Level     string `json:"level"`
	Drain     string `json:"drain"`
}

// TeamRanking ranks a candidate for the company.
type TeamRanking struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Level      LevelInfo `json:"level"`
	Drain      DrainInfo `json:"drain"`
	TeamFit    float64   `json:"team_fit"`
}

// --- Headhunter response ---

// HeadhunterMatchResponse is the output for headhunter matching.
type HeadhunterMatchResponse struct {
	Rankings []HeadhunterRanking `json:"rankings"`
}

// HeadhunterRanking is one company match for the candidate.
type HeadhunterRanking struct {
	Company         CompanyEntry   `json:"company"`
	Relation        RelationDetail `json:"relation"`
	Level           LevelInfo      `json:"level"`
	Drain           DrainInfo      `json:"drain"`
	PitchCandidate  string         `json:"pitch_to_candidate"`
	PitchCompany    string         `json:"pitch_to_company"`
}
