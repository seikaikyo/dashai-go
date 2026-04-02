---
title: Fortune 回應格式豐富化
type: feature
status: completed
created: 2026-04-03
---

# Fortune 回應格式豐富化

## 變更內容

Go fortune daily/weekly/monthly/yearly 回應目前只有 level 和基本宿位資訊，
前端期望的 advice、category descriptions、lucky items、weekday planets 等欄位都缺。
i18n JSON 裡已有這些資料，需要在計算函式中讀取並組裝進回應。

## 要補的欄位

### DailyFortune

| 欄位 | 資料來源 | 說明 |
|------|---------|------|
| weekday object | fortune_data.json weekday_planets | `{name, reading, yosei, planet}` |
| fortune.career_desc | fortunes.json daily_category_descriptions.career | 按 level 映射 |
| fortune.love_desc | fortunes.json daily_category_descriptions.love | 同上 |
| fortune.health_desc | fortunes.json daily_category_descriptions.health | 同上 |
| fortune.wealth_desc | fortunes.json daily_category_descriptions.wealth | 同上 |
| advice | fortune_data.json daily_advice | 按 level 隨機選一條 |
| lucky | fortune_data.json lucky_items | direction + color + numbers |
| day_mansion.day_fortune | fortunes.json daily_fortune_descriptions | auspicious/inauspicious |

### WeeklyFortune

| 欄位 | 資料來源 |
|------|---------|
| your_mansion | 從 days[0] 取 |
| fortune (aggregated) | AggregateLevel + category descs |
| daily_overview | 每天 {date, weekday, level, special_day, ryouhan, is_dark_week} |
| advice | fortunes.json weekly_fortune_focus |
| category_tips | fortunes.json weekly_category_tips |
| lucky | fortune_data.json lucky_items |

### MonthlyFortune

| 欄位 | 資料來源 |
|------|---------|
| your_mansion | birthIdx → mansion |
| month_mansion | 該月宿位 |
| relation | your_mansion vs month_mansion |
| theme | fortunes.json monthly_theme_descriptions |
| fortune (aggregated) | AggregateLevel + category descs |
| weekly[] | 分週聚合（每 7 天一組） |
| advice | fortunes.json monthly_fortune_advice |
| special_days[] | 該月特殊日列表 |
| ryouhan_info | 凌犯影響統計 |

### YearlyFortune

| 欄位 | 資料來源 |
|------|---------|
| your_mansion | birthIdx → mansion |
| kuyou_star (enriched) | kuyou.json stars[index] 加 description/fortune_name |
| fortune (aggregated) | AggregateLevel + category descs |
| theme | fortunes.json yearly_theme_descriptions |
| category_descriptions | career/love/health/wealth |
| opportunities[] | 從 bestMonths 生成 |
| warnings[] | 從 kyo months 生成 |
| advice | fortunes.json yearly_fortune_advice |
| monthly_trend (enriched) | 加 relation_type, ryouhan_ratio |

## 影響範圍

- `internal/shukuyo/fortune/daily.go` — DailyFortune struct + CalculateDaily
- `internal/shukuyo/fortune/weekly.go` — WeeklyFortune struct + CalculateWeekly
- `internal/shukuyo/fortune/monthly.go` — MonthlyFortune struct + CalculateMonthly
- `internal/shukuyo/fortune/yearly.go` — YearlyFortune struct + CalculateYearly
- `internal/shukuyo/fortune/data.go` — 可能需要新的 loader function

## 測試計畫

1. 現有 fortune 測試不能 break
2. Daily 回應包含 advice、lucky、category descriptions
3. Weekly 回應包含 daily_overview、advice
4. Monthly 回應包含 theme、weekly 分組
5. Yearly 回應包含 enriched kuyou_star、theme、advice
6. 三語切換正確

## Checklist

- [x] DailyFortune 豐富化（weekday/fortune categories/advice/lucky/day_fortune）
- [x] WeeklyFortune 豐富化（daily_overview/fortune/advice/focus/category_tips/lucky）
- [x] MonthlyFortune 豐富化（your_mansion/month_mansion/relation/theme/weekly/special_days/ryouhan_info/advice）
- [x] YearlyFortune 豐富化（your_mansion/enriched kuyou_star/theme/category_descriptions/opportunities/warnings/advice）
- [x] 測試通過（career 10/engine all/fortune all）
