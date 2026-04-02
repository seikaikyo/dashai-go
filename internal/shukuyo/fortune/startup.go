package fortune

import (
	"time"

	"github.com/seikaikyo/dashai-go/internal/shukuyo/engine"
)

// IndustryRecommendation is the response for startup industry advice.
type IndustryRecommendation struct {
	Mansion         IndustryMansion        `json:"mansion"`
	CareerTags      []string               `json:"career_tags"`
	CareerDesc      string                 `json:"career_description"`
	FavorableMansions []FavorableMansion    `json:"favorable_mansions"`
}

// IndustryMansion is the mansion info for industry recommendation.
type IndustryMansion struct {
	NameZH  string `json:"name_zh"`
	NameJP  string `json:"name_jp"`
	Reading string `json:"reading"`
	Index   int    `json:"index"`
}

// FavorableMansion is a business-partner mansion suggestion.
type FavorableMansion struct {
	Index   int    `json:"index"`
	NameZH  string `json:"name_zh"`
	NameJP  string `json:"name_jp"`
	Reading string `json:"reading"`
	Summary string `json:"summary"`
}

// CareerLuckyDate is a single date for career lucky dates.
type CareerLuckyDate struct {
	Date       string `json:"date"`
	Weekday    string `json:"weekday"`
	Level      string `json:"level"`
	LevelLabel string `json:"level_label"`
	DayMansion string `json:"day_mansion"`
	Relation   string `json:"relation"`
	Flags      []string `json:"flags"`
	Reason     string `json:"reason"`
}

// CareerLuckyDatesResult is the response for career lucky dates.
type CareerLuckyDatesResult struct {
	GoodDates []CareerLuckyDate `json:"good_dates"`
	BadDates  []CareerLuckyDate `json:"bad_dates"`
}

// CalculateIndustryRecommendation generates career/industry advice based on mansion.
func CalculateIndustryRecommendation(birthDateStr, lang string) (*IndustryRecommendation, error) {
	bd, err := engine.ParseDate(birthDateStr)
	if err != nil {
		return nil, err
	}

	idx, _ := engine.MansionIndexFromDate(bd)
	m := engine.Mansions27[idx]

	// Career tags based on mansion yosei (elemental affinity)
	tags := yoseiCareerTags[m.Yosei]
	desc := yoseiCareerDesc[m.Yosei]

	// Find eishin mansions as favorable business partners
	var favorable []FavorableMansion
	for targetIdx := 0; targetIdx < 27; targetIdx++ {
		rel := engine.GetRelation(idx, targetIdx)
		if rel.Group == "eishin" {
			tm := engine.Mansions27[targetIdx]
			favorable = append(favorable, FavorableMansion{
				Index:   targetIdx,
				NameZH:  tm.Name,
				NameJP:  tm.NameJP,
				Reading: tm.Reading,
				Summary: "栄親：互相提升的最佳合作夥伴",
			})
		}
	}

	return &IndustryRecommendation{
		Mansion: IndustryMansion{
			NameZH:  m.Name,
			NameJP:  m.NameJP,
			Reading: m.Reading,
			Index:   idx,
		},
		CareerTags: tags,
		CareerDesc: desc,
		FavorableMansions: favorable,
	}, nil
}

// CalculateCareerLuckyDates scans N days for career-relevant auspicious/inauspicious dates.
func CalculateCareerLuckyDates(birthDateStr string, days int, lang string) (*CareerLuckyDatesResult, error) {
	bd, err := engine.ParseDate(birthDateStr)
	if err != nil {
		return nil, err
	}
	lang = normalizeLang(lang)

	if days <= 0 {
		days = 30
	}
	if days > 90 {
		days = 90
	}

	birthIdx, _ := engine.MansionIndexFromDate(bd)

	var goodDates, badDates []CareerLuckyDate

	for i := 0; i < days; i++ {
		d := timeNow().AddDate(0, 0, i)
		dateStr := d.Format("2006-01-02")
		dayIdx := engine.DayMansionIndex(d)
		dayM := engine.Mansions27[dayIdx]
		rel := engine.GetRelation(birthIdx, dayIdx)

		sd := engine.CheckSpecialDay(d, dayIdx)
		rp := engine.CheckRyouhanPeriod(d)
		sdType := ""
		if sd != nil {
			sdType = sd.Type
		}
		ryouhanActive := rp != nil && rp.Active
		level, _ := DetermineLevel(rel.Group, rel.Direction, ryouhanActive, sdType)
		jpwd := engine.JPWeekday(d)

		var flags []string
		if sd != nil {
			flags = append(flags, sd.Name)
		}
		if ryouhanActive {
			flags = append(flags, "凌犯期間")
		}

		entry := CareerLuckyDate{
			Date:       dateStr,
			Weekday:    WeekdayName(jpwd, lang),
			Level:      level,
			LevelLabel: LevelName(level, lang),
			DayMansion: dayM.NameJP,
			Relation:   rel.Group,
			Flags:      flags,
			Reason:     rel.Group + " / " + rel.Direction,
		}

		switch level {
		case LevelDaikichi, LevelKichi:
			if rel.Group != "ankai" {
				goodDates = append(goodDates, entry)
			}
		case LevelKyo:
			badDates = append(badDates, entry)
		}
	}

	return &CareerLuckyDatesResult{
		GoodDates: goodDates,
		BadDates:  badDates,
	}, nil
}

// timeNow is a package-level function for testability.
var timeNow = func() time.Time {
	return time.Now().Truncate(24 * time.Hour)
}

// Yosei-based career recommendations
var yoseiCareerTags = map[string][]string{
	"木": {"教育", "出版", "農業", "環保", "醫療", "社會福利", "研發"},
	"金": {"金融", "法律", "珠寶", "精密工業", "品質管理", "會計"},
	"土": {"建設", "不動產", "陶藝", "食品", "農產加工", "倉儲物流"},
	"日": {"能源", "表演藝術", "領導管理", "政治", "媒體", "公關"},
	"月": {"餐飲", "旅遊", "服務業", "照護", "心理諮商", "藝術"},
	"火": {"科技", "工程", "軍警", "運動", "機械", "電子"},
	"水": {"貿易", "物流", "漁業", "旅行", "外交", "翻譯", "進出口"},
}

var yoseiCareerDesc = map[string]string{
	"木": "木性宿位適合需要成長、培育、教化的行業。你的能量在「生長」的場域最能發揮。",
	"金": "金性宿位適合需要精確、判斷、收斂的行業。你的能量在講求標準與品質的場域最能發揮。",
	"土": "土性宿位適合需要穩定、累積、承載的行業。你的能量在建設與滋養的場域最能發揮。",
	"日": "日性宿位適合需要光明、領導、展現的行業。你的能量在舞台與決策的場域最能發揮。",
	"月": "月性宿位適合需要感受、包容、服務的行業。你的能量在照顧與創意的場域最能發揮。",
	"火": "火性宿位適合需要動力、突破、技術的行業。你的能量在挑戰與行動的場域最能發揮。",
	"水": "水性宿位適合需要流動、溝通、連結的行業。你的能量在交流與貿易的場域最能發揮。",
}
