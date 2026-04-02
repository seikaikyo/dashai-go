package fortune

import (
	"embed"
	"encoding/json"
	"sync"
)

//go:embed data/*.json data/i18n/*/*.json
var dataFS embed.FS

// --- Cached data loading ---

var (
	fortuneDataOnce sync.Once
	fortuneData     map[string]any

	i18nCache sync.Map // lang -> map[string]any (fortunes.json)
	fdCache   sync.Map // lang -> map[string]any (fortune_data.json)
	sankiCache sync.Map // lang -> map[string]any (sanki.json)
	kuyouCache sync.Map // lang -> map[string]any (kuyou.json)
)

func loadFortuneData() map[string]any {
	fortuneDataOnce.Do(func() {
		fortuneData = loadJSON("data/sukuyodo_fortune.json")
	})
	return fortuneData
}

func loadI18n(lang string) map[string]any {
	return loadCached(&i18nCache, lang, "fortunes.json")
}

func loadFortuneI18n(lang string) map[string]any {
	return loadCached(&fdCache, lang, "fortune_data.json")
}

func loadSankiI18n(lang string) map[string]any {
	return loadCached(&sankiCache, lang, "sanki.json")
}

func loadKuyouI18n(lang string) map[string]any {
	return loadCached(&kuyouCache, lang, "kuyou.json")
}

func loadCached(cache *sync.Map, lang, filename string) map[string]any {
	if v, ok := cache.Load(lang); ok {
		return v.(map[string]any)
	}
	lang = normalizeLang(lang)
	path := "data/i18n/" + lang + "/" + filename
	data := loadJSON(path)
	if data == nil {
		// Fallback to zh-TW
		data = loadJSON("data/i18n/zh-TW/" + filename)
	}
	if data == nil {
		data = map[string]any{}
	}
	cache.Store(lang, data)
	return data
}

func loadJSON(path string) map[string]any {
	b, err := dataFS.ReadFile(path)
	if err != nil {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return nil
	}
	return m
}

func normalizeLang(lang string) string {
	switch lang {
	case "ja", "en":
		return lang
	default:
		return "zh-TW"
	}
}

func langKey(lang string) string {
	switch lang {
	case "ja":
		return "ja"
	case "en":
		return "en"
	default:
		return "zh"
	}
}

// WeekdayShort returns the short weekday name.
var weekdayShort = map[string][7]string{
	"zh": {"日", "月", "火", "水", "木", "金", "土"},
	"ja": {"日", "月", "火", "水", "木", "金", "土"},
	"en": {"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"},
}

func WeekdayName(jpWeekday int, lang string) string {
	k := langKey(lang)
	if jpWeekday < 0 || jpWeekday > 6 {
		return ""
	}
	return weekdayShort[k][jpWeekday]
}
