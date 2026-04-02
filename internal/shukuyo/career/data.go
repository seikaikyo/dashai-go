package career

import (
	"embed"
	"encoding/json"
	"sync"
)

//go:embed data/*.json data/employment/*/*.json
var dataFS embed.FS

// --- Cached data loading ---

var (
	relationsCache sync.Map // "context/lang" -> map[string]RelationText
	careerCache    sync.Map // "context/lang" -> map[string]CareerText
	drainCache     sync.Map // "context/lang" -> map[string]DrainText
	redFlagsCache  sync.Map // "context/lang" -> map[string][]string
)

// RelationText holds the i18n text for a relation group.
type RelationText struct {
	Description      string   `json:"description"`
	Detailed         string   `json:"detailed"`
	Advice           string   `json:"advice"`
	Tips             []string `json:"tips"`
	Avoid            []string `json:"avoid"`
	GoodFor          []string `json:"good_for"`
	DescriptionClassic string `json:"description_classic"`
}

// CareerText holds career advice for a direction.
type CareerText struct {
	SutraKey    string   `json:"sutra_key"`
	Headline    string   `json:"headline"`
	ContextNote string   `json:"context_note"`
	Detail      string   `json:"detail"`
	Do          []string `json:"do"`
	Avoid       []string `json:"avoid"`
}

// DrainText holds drain analysis text for a direction.
type DrainText struct {
	Relationship  string    `json:"relationship"`
	DailyFeel     string    `json:"daily_feel"`
	EnergyFlow    string    `json:"energy_flow_desc"`
	LongTermRisk  string    `json:"long_term_risk"`
	Growth        string    `json:"growth"`
	Stability     string    `json:"stability"`
	BadYearImpact string    `json:"bad_year_impact"`
	FitFor        string    `json:"fit_for"`
	HR            *DrainHR  `json:"hr,omitempty"`
}

// DrainHR holds HR-perspective drain text.
type DrainHR struct {
	Relationship  string `json:"relationship"`
	DailyFeel     string `json:"daily_feel"`
	BadYearImpact string `json:"bad_year_impact"`
	FitFor        string `json:"fit_for"`
}

// directionToRomaji maps kanji direction to romaji key used in JSON files.
var directionToRomaji = map[string]string{
	"命": "mei", "栄": "ei", "衰": "sui", "安": "an",
	"壊": "kai", "危": "ki", "成": "sei", "友": "yu",
	"親": "shin", "業": "gyo", "胎": "tai",
}

// DirectionRomaji converts a kanji direction to its romaji key.
func DirectionRomaji(direction string) string {
	if r, ok := directionToRomaji[direction]; ok {
		return r
	}
	return direction
}

// LoadRelations loads relation texts for a context and language.
func LoadRelations(ctx ContextType, lang string) map[string]RelationText {
	lang = normalizeLang(lang)
	key := string(ctx) + "/" + lang
	if v, ok := relationsCache.Load(key); ok {
		return v.(map[string]RelationText)
	}
	data := loadTypedJSON[map[string]RelationText]("data/" + string(ctx) + "/" + lang + "/relations.json")
	if data == nil {
		data = loadTypedJSON[map[string]RelationText]("data/" + string(ctx) + "/zh-TW/relations.json")
	}
	if data == nil {
		empty := map[string]RelationText{}
		relationsCache.Store(key, empty)
		return empty
	}
	relationsCache.Store(key, *data)
	return *data
}

// LoadCareer loads career advice texts for a context and language.
func LoadCareer(ctx ContextType, lang string) map[string]CareerText {
	lang = normalizeLang(lang)
	key := string(ctx) + "/" + lang
	if v, ok := careerCache.Load(key); ok {
		return v.(map[string]CareerText)
	}
	data := loadTypedJSON[map[string]CareerText]("data/" + string(ctx) + "/" + lang + "/career_seeker.json")
	if data == nil {
		data = loadTypedJSON[map[string]CareerText]("data/" + string(ctx) + "/zh-TW/career_seeker.json")
	}
	if data == nil {
		empty := map[string]CareerText{}
		careerCache.Store(key, empty)
		return empty
	}
	careerCache.Store(key, *data)
	return *data
}

// LoadDrainTexts loads drain analysis texts for a context and language.
func LoadDrainTexts(ctx ContextType, lang string) map[string]DrainText {
	lang = normalizeLang(lang)
	key := string(ctx) + "/" + lang
	if v, ok := drainCache.Load(key); ok {
		return v.(map[string]DrainText)
	}
	data := loadTypedJSON[map[string]DrainText]("data/" + string(ctx) + "/" + lang + "/drain_analysis.json")
	if data == nil {
		data = loadTypedJSON[map[string]DrainText]("data/" + string(ctx) + "/zh-TW/drain_analysis.json")
	}
	if data == nil {
		empty := map[string]DrainText{}
		drainCache.Store(key, empty)
		return empty
	}
	drainCache.Store(key, *data)
	return *data
}

// LoadRedFlags loads red flag warnings for a context and language.
func LoadRedFlags(ctx ContextType, lang string) map[string][]string {
	lang = normalizeLang(lang)
	key := string(ctx) + "/" + lang
	if v, ok := redFlagsCache.Load(key); ok {
		return v.(map[string][]string)
	}
	data := loadTypedJSON[map[string][]string]("data/" + string(ctx) + "/" + lang + "/red_flags.json")
	if data == nil {
		data = loadTypedJSON[map[string][]string]("data/" + string(ctx) + "/zh-TW/red_flags.json")
	}
	if data == nil {
		empty := map[string][]string{}
		redFlagsCache.Store(key, empty)
		return empty
	}
	redFlagsCache.Store(key, *data)
	return *data
}

func loadTypedJSON[T any](path string) *T {
	b, err := dataFS.ReadFile(path)
	if err != nil {
		return nil
	}
	var v T
	if err := json.Unmarshal(b, &v); err != nil {
		return nil
	}
	return &v
}

func normalizeLang(lang string) string {
	switch lang {
	case "ja", "en":
		return lang
	default:
		return "zh-TW"
	}
}
