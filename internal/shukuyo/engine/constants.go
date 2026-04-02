package engine

// === T21n1299 品二: 月宿傍通曆 ===

// MonthStartMansion maps lunar month (1-12) to starting mansion index (0-26).
var MonthStartMansion = map[int]int{
	1: 11, 2: 13, 3: 15, 4: 17, 5: 19, 6: 21,
	7: 24, 8: 0, 9: 2, 10: 4, 11: 7, 12: 9,
}

// === T21n1299: 二十七宿 ===

// Mansions27 defines the 27 lunar mansions in order (index 0-26).
var Mansions27 = [27]Mansion{
	{0, "角", "角宿", "かくしゅく", "木"},
	{1, "亢", "亢宿", "こうしゅく", "金"},
	{2, "氐", "氐宿", "ていしゅく", "土"},
	{3, "房", "房宿", "ぼうしゅく", "日"},
	{4, "心", "心宿", "しんしゅく", "月"},
	{5, "尾", "尾宿", "びしゅく", "火"},
	{6, "箕", "箕宿", "きしゅく", "水"},
	{7, "斗", "斗宿", "としゅく", "木"},
	{8, "女", "女宿", "じょしゅく", "金"},
	{9, "虛", "虛宿", "きょしゅく", "土"},
	{10, "危", "危宿", "きしゅく", "日"},
	{11, "室", "室宿", "しつしゅく", "月"},
	{12, "壁", "壁宿", "へきしゅく", "火"},
	{13, "奎", "奎宿", "けいしゅく", "水"},
	{14, "婁", "婁宿", "ろうしゅく", "木"},
	{15, "胃", "胃宿", "いしゅく", "金"},
	{16, "昴", "昴宿", "ぼうしゅく", "土"},
	{17, "畢", "畢宿", "ひつしゅく", "日"},
	{18, "觜", "觜宿", "ししゅく", "月"},
	{19, "参", "参宿", "さんしゅく", "火"},
	{20, "井", "井宿", "せいしゅく", "水"},
	{21, "鬼", "鬼宿", "きしゅく", "木"},
	{22, "柳", "柳宿", "りゅうしゅく", "金"},
	{23, "星", "星宿", "せいしゅく", "土"},
	{24, "張", "張宿", "ちょうしゅく", "日"},
	{25, "翼", "翼宿", "よくしゅく", "月"},
	{26, "軫", "軫宿", "しんしゅく", "火"},
}

// === T21n1299 品二: 三九秘法位名 ===

// SankuPositionNames maps forward distance (0-26) to position name.
// Three groups of 9: 一九(命行), 二九(業行), 三九(胎行).
// Each group: start position + 栄→衰→安→危→成→壊→友→親.
var SankuPositionNames = [27]string{
	"命", "栄", "衰", "安", "危", "成", "壊", "友", "親",
	"業", "栄", "衰", "安", "危", "成", "壊", "友", "親",
	"胎", "栄", "衰", "安", "危", "成", "壊", "友", "親",
}

// positionToGroup maps a position name to its paired relationship group.
// Groups are the original paired kanji: 栄親/友衰/安壊/危成/命/業胎.
var positionToGroup = map[string]string{
	"栄": "eishin", "親": "eishin",
	"友": "yusui", "衰": "yusui",
	"安": "ankai", "壊": "ankai",
	"危": "kisei", "成": "kisei",
	"命": "mei",
	"業": "gyotai", "胎": "gyotai",
}

// DirectionInverse maps a direction to its counterpart (person2's perspective).
// From T21n1299: paired positions are symmetric.
var DirectionInverse = map[string]string{
	"栄": "親", "親": "栄",
	"友": "衰", "衰": "友",
	"安": "壊", "壊": "安",
	"危": "成", "成": "危",
	"命": "命",
	"業": "胎", "胎": "業",
}

// === T21n1299 品二: 三九日型 ===

// SankiPeriodNames holds the period names.
var SankiPeriodNames = [3]string{"善期（一九）", "惡期（二九）", "中期（三九）"}

// SankiDayTypeShared maps position within period (2-9) to day type name.
var SankiDayTypeShared = [10]string{
	"", "",
	"栄の日", "衰の日", "安の日", "危の日",
	"成の日", "壊の日", "友の日", "親の日",
}

// SankiPeriodStartNames maps period index (0,1,2) to start day name.
var SankiPeriodStartNames = [3]string{"命の日", "業の日", "胎の日"}

// === T21n1299 品四: 九曜星 ===

type kuyouStarDef struct {
	Name    string
	Reading string
	Yosei   string // empty for rahula/ketu (eclipse nodes)
	Buddha  string
}

// KuyouStarDefs: star names, readings, and guardian buddha from 品四.
var KuyouStarDefs = [9]kuyouStarDef{
	{"羅喉星", "らごうせい", "", "不動明王"},
	{"土曜星", "どようせい", "土", "聖觀音"},
	{"水曜星", "すいようせい", "水", "彌勒菩薩"},
	{"金曜星", "きんようせい", "金", "阿彌陀如來"},
	{"日曜星", "にちようせい", "日", "千手觀音"},
	{"火曜星", "かようせい", "火", "虛空藏菩薩"},
	{"計都星", "けいとせい", "", "釋迦如來"},
	{"月曜星", "げつようせい", "月", "勢至菩薩"},
	{"木曜星", "もくようせい", "木", "藥師如來"},
}

// === T21n1299 卷五: 甘露日/金剛峯日/羅刹日 ===

type specialDayKey struct {
	weekday      int // 0=Sun, 1=Mon, ..., 6=Sat
	mansionIndex int
}

// SpecialDayMap maps (weekday, mansion_index) to special day type.
var SpecialDayMap = map[specialDayKey]string{
	// 甘露日
	{0, 26}: "kanro", {1, 17}: "kanro", {2, 5}: "kanro",
	{3, 22}: "kanro", {4, 21}: "kanro", {5, 3}: "kanro", {6, 23}: "kanro",
	// 金剛峯日
	{0, 5}: "kongou", {1, 4}: "kongou", {2, 12}: "kongou",
	{3, 16}: "kongou", {4, 20}: "kongou", {5, 24}: "kongou", {6, 1}: "kongou",
	// 羅刹日
	{0, 15}: "rasetsu", {1, 21}: "rasetsu", {2, 25}: "rasetsu",
	{3, 19}: "rasetsu", {4, 2}: "rasetsu", {5, 13}: "rasetsu", {6, 22}: "rasetsu",
}

// SpecialDayInfo holds display info for special day types.
var SpecialDayInfo = map[string]SpecialDayType{
	"kanro":   {"kanro", "甘露日", "かんろにち"},
	"kongou":  {"kongou", "金剛峯日", "こんごうぶにち"},
	"rasetsu": {"rasetsu", "羅刹日", "らせつにち"},
}

// === T21n1299 品三/品五: 凌犯 (七曜陵逼) ===

type ryouhanKey struct {
	month   int
	weekday int
}

// RyouhanMap maps (lunar_month, first_day_weekday) to (start_day, end_day).
var RyouhanMap = map[ryouhanKey][2]int{
	{1, 6}: {1, 16}, {1, 0}: {17, 30},
	{2, 1}: {1, 14}, {2, 2}: {15, 30},
	{3, 3}: {1, 12}, {3, 4}: {13, 30},
	{4, 5}: {1, 10}, {4, 6}: {11, 30},
	{5, 0}: {1, 8}, {5, 1}: {9, 30},
	{6, 2}: {1, 6}, {6, 3}: {7, 30},
	{7, 5}: {1, 3}, {7, 6}: {4, 30},
	{8, 2}: {1, 27},
	{9, 4}: {1, 25}, {9, 5}: {26, 30},
	{10, 6}: {1, 23}, {10, 0}: {24, 30},
	{11, 2}: {1, 20}, {11, 3}: {21, 30},
	{12, 4}: {1, 18}, {12, 5}: {19, 30},
}
