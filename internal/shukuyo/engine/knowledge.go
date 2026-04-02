package engine

// MansionDetail is the rich mansion info for the knowledge base.
type MansionDetail struct {
	Index   int    `json:"index"`
	NameJP  string `json:"name_jp"`
	NameZH  string `json:"name_zh"`
	Reading string `json:"reading"`
	Yosei   string `json:"yosei"`

	// Personality descriptions (modern interpretation)
	Personality        string   `json:"personality"`
	PersonalityClassic string   `json:"personality_classic,omitempty"`
	Keywords           []string `json:"keywords"`

	// Day fortune from sutra
	DayFortune *DayFortuneInfo `json:"day_fortune,omitempty"`

	// Lunar date ranges (which lunar month/days map to this mansion)
	LunarDates []LunarDateRange `json:"lunar_dates"`
}

// DayFortuneInfo holds day-specific fortune from T21n1299.
type DayFortuneInfo struct {
	Auspicious       []string `json:"auspicious"`
	Inauspicious     []string `json:"inauspicious"`
	Summary          string   `json:"summary"`
	SummaryClassic   string   `json:"summary_classic,omitempty"`
	IsMostAuspicious bool     `json:"is_most_auspicious"`
}

// LunarDateRange represents a lunar month/day range that maps to a mansion.
type LunarDateRange struct {
	Month    int `json:"month"`
	StartDay int `json:"start_day"`
	EndDay   int `json:"end_day"`
}

// RelationTypeInfo is the rich relation info for the knowledge base.
type RelationTypeInfo struct {
	Type           string `json:"type"`
	Name           string `json:"name"`
	NameJP         string `json:"name_jp"`
	Reading        string `json:"reading"`
	Level          string `json:"level"`
	Description    string `json:"description"`
	DescClassic    string `json:"description_classic,omitempty"`
	Detailed       string `json:"detailed"`
	Advice         string `json:"advice"`
	Tips           []string `json:"tips"`
	Avoid          []string `json:"avoid"`
	GoodFor        []string `json:"good_for"`
}

// KnowledgeMetadata holds the origin/scripture/history info.
type KnowledgeMetadata struct {
	Name            string `json:"name"`
	Reading         string `json:"reading"`
	Origin          string `json:"origin"`
	OriginReading   string `json:"origin_reading"`
	Founder         string `json:"founder"`
	FounderReading  string `json:"founder_reading"`
	Scripture       string `json:"scripture"`
	ScriptureReading string `json:"scripture_reading"`
	Method          string `json:"method"`
	MethodReading   string `json:"method_reading"`
}

// GetAllMansions returns detailed info for all 27 mansions.
func GetAllMansions() []MansionDetail {
	mansions := make([]MansionDetail, 27)
	for i, m := range Mansions27 {
		mansions[i] = MansionDetail{
			Index:      m.Index,
			NameJP:     m.NameJP,
			NameZH:     m.Name,
			Reading:    m.Reading,
			Yosei:      m.Yosei,
			Personality: mansionPersonalities[i],
			Keywords:   mansionKeywords[i],
			LunarDates: computeLunarDates(i),
		}
	}
	return mansions
}

// GetAllRelationTypes returns info for all 6 relation groups.
func GetAllRelationTypes() []RelationTypeInfo {
	return []RelationTypeInfo{
		{Type: "eishin", Name: "栄親", NameJP: "栄親", Reading: "えいしん", Level: "daikichi",
			Description: "互相提升的最佳組合", Detailed: "栄意指繁榮興旺，親意指親密無間。雙方都會往更好的方向發展。",
			Advice: "品質很高但仍需經營，定期表達感謝。",
			Tips: []string{"積極合作", "互相扶持", "長期投資"}, Avoid: []string{"視為理所當然", "疏於維護"}, GoodFor: []string{"長期夥伴", "核心團隊"}},
		{Type: "gyotai", Name: "業胎", NameJP: "業胎", Reading: "ごうたい", Level: "kichi",
			Description: "前世因緣的深層默契", Detailed: "業方背負推動對方的課題，胎方接受培育。合作效率極高。",
			Advice: "善用天生默契但不要省略正式溝通。",
			Tips: []string{"善用默契", "定期溝通確認"}, Avoid: []string{"只靠默契不說"}, GoodFor: []string{"高默契團隊", "穩定夥伴"}},
		{Type: "mei", Name: "命", NameJP: "命", Reading: "めい", Level: "kichi",
			Description: "同宿同體的鏡像組合", Detailed: "兩人的本命宿完全相同，理解彼此幾乎不費力。但互補性不強。",
			Advice: "享受零障礙溝通但要引入外部觀點。",
			Tips: []string{"善用溝通優勢", "引入外部視角"}, Avoid: []string{"忽略互補不足", "一起鑽牛角尖"}, GoodFor: []string{"高共識決策", "文化適配"}},
		{Type: "yusui", Name: "友衰", NameJP: "友衰", Reading: "ゆうすい", Level: "shokyo",
			Description: "友善但不對等的付出關係", Detailed: "友方主動付出社交能量，衰方享受但回饋有限。時間久了友方會疲累。",
			Advice: "友方設好界限，衰方主動回饋。",
			Tips: []string{"友方設界限", "短期合作"}, Avoid: []string{"無限制付出", "長期深度綁定"}, GoodFor: []string{"短期專案", "明確分工"}},
		{Type: "kisei", Name: "危成", NameJP: "危成", Reading: "きせい", Level: "shokyo",
			Description: "互補但需要磨合的搭檔", Detailed: "能力和觀點互補，磨合過了之後會是高效搭檔。",
			Advice: "給磨合期足夠耐心，初期不適不代表不合適。",
			Tips: []string{"給時間磨合", "善用互補"}, Avoid: []string{"因初期不適放棄"}, GoodFor: []string{"互補團隊", "長期合作"}},
		{Type: "ankai", Name: "安壊", NameJP: "安壊", Reading: "あんかい", Level: "kyo",
			Description: "張力最大的高風險組合", Detailed: "安方持續輸出維穩，壊方從安方抽取能量。短期可刺激成長，長期風險高。",
			Advice: "安方設界限，壊方節制，需要外部平衡機制。",
			Tips: []string{"安方設界限", "短期比長期安全"}, Avoid: []string{"無限制付出", "長期綁定"}, GoodFor: []string{"短期高壓專案"}},
	}
}

// GetMetadata returns the knowledge base metadata.
func GetMetadata() KnowledgeMetadata {
	return KnowledgeMetadata{
		Name:             "宿曜道",
		Reading:          "すくようどう",
		Origin:           "印度・中國・日本",
		OriginReading:    "いんど・ちゅうごく・にほん",
		Founder:          "不空三藏",
		FounderReading:   "ふくうさんぞう",
		Scripture:        "文殊師利菩薩及諸仙所說吉凶時日善惡宿曜經",
		ScriptureReading: "もんじゅしりぼさつきゅうしょせんしょせつきっきょうじじつぜんあくすくようきょう",
		Method:           "二十七宿 × 三九秘法",
		MethodReading:    "にじゅうしちしゅく × さんくひほう",
	}
}

// CompatibilityBatchEntry is one pair in a batch compatibility request.
type CompatibilityBatchEntry struct {
	ID   string `json:"id"`
	Date string `json:"date"`
}

// CompatibilityBatchResult is one result from batch compatibility.
type CompatibilityBatchResult struct {
	ID    string              `json:"id"`
	Data  *CompatibilityResult `json:"data,omitempty"`
	Error string              `json:"error,omitempty"`
}

// CompatibilityBatch calculates compatibility for multiple pairs.
func CompatibilityBatch(date1Str string, partners []CompatibilityBatchEntry) []CompatibilityBatchResult {
	d1, err := ParseDate(date1Str)
	if err != nil {
		results := make([]CompatibilityBatchResult, len(partners))
		for i, p := range partners {
			results[i] = CompatibilityBatchResult{ID: p.ID, Error: "invalid date1"}
		}
		return results
	}

	results := make([]CompatibilityBatchResult, len(partners))
	for i, p := range partners {
		d2, err := ParseDate(p.Date)
		if err != nil {
			results[i] = CompatibilityBatchResult{ID: p.ID, Error: "invalid date: " + p.Date}
			continue
		}
		compat := Compatibility(d1, d2)
		results[i] = CompatibilityBatchResult{ID: p.ID, Data: &compat}
	}
	return results
}

// CompatibilityFinderResult groups all 27 mansions by relation type for a birth date.
type CompatibilityFinderResult struct {
	YourMansion MansionResult                        `json:"your_mansion"`
	Mei         CompatibilityFinderGroup             `json:"mei"`
	Gyotai      CompatibilityFinderGroup             `json:"gyotai"`
	Eishin      CompatibilityFinderGroup             `json:"eishin"`
	Yusui       CompatibilityFinderGroup             `json:"yusui"`
	Ankai       CompatibilityFinderGroup             `json:"ankai"`
	Kisei       CompatibilityFinderGroup             `json:"kisei"`
}

// CompatibilityFinderGroup holds mansions in one relation group.
type CompatibilityFinderGroup struct {
	Relation    string              `json:"relation"`
	Reading     string              `json:"reading"`
	Level       string              `json:"level"`
	Description string              `json:"description"`
	Mansions    []CompatibleMansion `json:"mansions"`
}

// CompatibleMansion is a mansion in the finder result.
type CompatibleMansion struct {
	NameJP     string           `json:"name_jp"`
	NameZH     string           `json:"name_zh"`
	Reading    string           `json:"reading"`
	Index      int              `json:"index"`
	Yosei      string           `json:"yosei"`
	Keywords   []string         `json:"keywords"`
	LunarDates []LunarDateRange `json:"lunar_dates"`
}

// FindCompatibility generates the full compatibility finder for a birth date.
func FindCompatibility(birthDateStr string) (*CompatibilityFinderResult, error) {
	bd, err := ParseDate(birthDateStr)
	if err != nil {
		return nil, err
	}

	mansion := GetMansion(bd)
	idx := mansion.MansionIndex

	// Group all 27 mansions by their relation to this person
	groups := map[string]*CompatibilityFinderGroup{
		"mei":    {Relation: "命", Reading: "めい", Level: "kichi", Description: "同宿同體的鏡像組合"},
		"gyotai": {Relation: "業胎", Reading: "ごうたい", Level: "kichi", Description: "前世因緣的深層默契"},
		"eishin": {Relation: "栄親", Reading: "えいしん", Level: "daikichi", Description: "互相提升的最佳組合"},
		"yusui":  {Relation: "友衰", Reading: "ゆうすい", Level: "shokyo", Description: "友善但不對等的付出"},
		"ankai":  {Relation: "安壊", Reading: "あんかい", Level: "kyo", Description: "張力最大的高風險組合"},
		"kisei":  {Relation: "危成", Reading: "きせい", Level: "shokyo", Description: "互補但需要磨合"},
	}

	for targetIdx := 0; targetIdx < 27; targetIdx++ {
		rel := GetRelation(idx, targetIdx)
		m := Mansions27[targetIdx]
		cm := CompatibleMansion{
			NameJP:     m.NameJP,
			NameZH:     m.Name,
			Reading:    m.Reading,
			Index:      targetIdx,
			Yosei:      m.Yosei,
			Keywords:   mansionKeywords[targetIdx],
			LunarDates: computeLunarDates(targetIdx),
		}
		if g, ok := groups[rel.Group]; ok {
			g.Mansions = append(g.Mansions, cm)
		}
	}

	return &CompatibilityFinderResult{
		YourMansion: mansion,
		Mei:         *groups["mei"],
		Gyotai:      *groups["gyotai"],
		Eishin:      *groups["eishin"],
		Yusui:       *groups["yusui"],
		Ankai:       *groups["ankai"],
		Kisei:       *groups["kisei"],
	}, nil
}

// computeLunarDates returns which lunar month/day ranges map to a given mansion index.
func computeLunarDates(targetIdx int) []LunarDateRange {
	var ranges []LunarDateRange
	for month := 1; month <= 12; month++ {
		start := MonthStartMansion[month]
		// Find which days in this month map to targetIdx
		// day d maps to mansion (start + d - 1) % 27
		// We need (start + d - 1) % 27 == targetIdx
		// d = (targetIdx - start + 1 + 27) % 27
		d := (targetIdx - start + 1 + 27) % 27
		if d == 0 {
			d = 27
		}
		// In a lunar month (29-30 days), this mansion appears once (d) or twice (d, d+27)
		if d <= 30 {
			ranges = append(ranges, LunarDateRange{Month: month, StartDay: d, EndDay: d})
		}
		d2 := d + 27
		if d2 <= 30 {
			ranges = append(ranges, LunarDateRange{Month: month, StartDay: d2, EndDay: d2})
		}
	}
	return ranges
}

// mansionPersonalities: modern personality descriptions for each mansion.
var mansionPersonalities = [27]string{
	"領導力強，行動果斷，適合開創事業",
	"品味高雅，嚴謹精確，適合品質管理",
	"穩重踏實，基礎紮實，適合長期規劃",
	"熱情主動，直覺敏銳，適合業務開發",
	"感受力強，情緒豐富，適合創意工作",
	"堅韌不拔，耐力持久，適合技術深耕",
	"靈活變通，善於溝通，適合協調工作",
	"志向遠大，計畫周密，適合戰略規劃",
	"謹慎細膩，分析力強，適合研究工作",
	"理想主義，追求完美，適合學術研究",
	"冒險精神，不怕挑戰，適合危機處理",
	"創造力強，想像豐富，適合設計創新",
	"社交能力強，人脈廣闊，適合公關行銷",
	"知識廣博，好學不倦，適合教育培訓",
	"親和力強，善於協調，適合人力資源",
	"決斷力強，執行力高，適合專案管理",
	"藝術天分，審美獨到，適合文化創意",
	"實事求是，重視效率，適合工程製造",
	"觀察入微，判斷精準，適合投資分析",
	"勇於突破，不畏困難，適合探索創新",
	"邏輯清晰，條理分明，適合系統開發",
	"直覺靈敏，洞察人心，適合心理諮商",
	"表達力強，善於說服，適合法律辯護",
	"光明磊落，領袖魅力，適合高階管理",
	"氣勢恢宏，格局宏大，適合企業經營",
	"默默耕耘，厚積薄發，適合幕後工作",
	"善始善終，收尾完美，適合品質保證",
}

// mansionKeywords: keyword tags for each mansion.
var mansionKeywords = [27][]string{
	{"領導", "行動", "開創"},
	{"品味", "精確", "和諧"},
	{"穩重", "基礎", "踏實"},
	{"熱情", "直覺", "主動"},
	{"感受", "創意", "情緒"},
	{"堅韌", "耐力", "專注"},
	{"靈活", "溝通", "變通"},
	{"志向", "計畫", "遠見"},
	{"謹慎", "分析", "細膩"},
	{"理想", "完美", "學術"},
	{"冒險", "挑戰", "危機"},
	{"創造", "想像", "設計"},
	{"社交", "人脈", "行銷"},
	{"知識", "好學", "教育"},
	{"親和", "協調", "人資"},
	{"決斷", "執行", "管理"},
	{"藝術", "審美", "創意"},
	{"效率", "實務", "工程"},
	{"觀察", "判斷", "分析"},
	{"突破", "探索", "創新"},
	{"邏輯", "系統", "開發"},
	{"直覺", "洞察", "心理"},
	{"表達", "說服", "法律"},
	{"光明", "領袖", "管理"},
	{"氣勢", "格局", "經營"},
	{"耕耘", "累積", "幕後"},
	{"收尾", "品質", "完善"},
}
