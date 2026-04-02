---
title: Shukuyo Career Phase 3 -- 情境化職涯相性系統
type: feature
status: completed
created: 2026-04-02
---

# Shukuyo Career Phase 3 -- 情境化職涯相性系統

## 變更內容

從 Python dashai-api 的單一「求職者 vs 公司」模式，重新設計為 6 種情境的職涯相性系統。
同一組宿曜關係在不同情境有不同解讀（如安壊在直僱 = 權力失衡警告，在顧問 = 短期信任快速建立）。

### 6 種情境

| Context Key | 名稱 | 主體 | 核心解讀 |
|-------------|------|------|----------|
| employment | 直僱 | 我 vs 公司 | 長期相處、日常摩擦、職涯發展 |
| consulting | 顧問/派遣 | 我 vs 案場 | 短期合作、專案導向、溝通效率 |
| outsourcing | 外包/承攬 | 我的公司 vs 客戶公司 | 契約關係、業務互補、付款風險 |
| b2b | 企業合作 | 公司 A vs 公司 B | 策略聯盟、長期夥伴、互利 |
| hr | HR 選才 | 候選人 vs 公司 | 適配度、團隊互補、留任率 |
| headhunter | 獵頭配對 | 候選人 vs 目標公司 | 第三方視角、匹配率、pitch |

### 核心設計原則

- career/ import engine/（原典計算），不碰 fortune/（等級系統留在 fortune）
- 每個情境有獨立的 JSON 文案（三語 zh-TW/ja/en）
- 直僱模式結果與 Python 現有 API backward compatible
- 先做直僱（沿用現有文案結構），其他情境漸進加入

## 影響範圍

### 新增檔案

```
internal/shukuyo/career/
├── types.go            -- 所有 request/response 型別
├── context.go          -- 6 種情境定義與切換邏輯
├── analyze.go          -- 核心相性分析（含 drain、red_flags、initiative）
├── batch.go            -- 批次分析（20 間公司 < 2 秒）
├── comparison.go       -- 公司比較（十年對照表）
├── interview_dates.go  -- 面試吉日
├── team_matrix.go      -- HR 團隊矩陣（50 人 < 5 秒）
├── headhunter.go       -- 獵頭配對（排名 + pitch）
├── drain.go            -- 消耗度分析（relation + direction → severity）
├── data.go             -- embed FS + i18n 載入
├── handler.go          -- HTTP handlers
├── router.go           -- chi.Router
└── career_test.go      -- 測試

internal/shukuyo/data/career/
├── drain.json                      -- drain level 對照表
├── contexts.json                   -- 情境 metadata
└── employment/{zh-TW,ja,en}/
    ├── relations.json              -- 6 組關係的情境化解讀
    ├── red_flags.json              -- 紅旗警告
    ├── initiative.json             -- 主動建議
    ├── drain.json                  -- 消耗度文案
    └── gap_guidance.json           -- 缺口補救建議
```

### 修改檔案

- `cmd/server/main.go` -- 掛載 `/shukuyo/career` 路由

## API 端點

| Method | Path | 說明 |
|--------|------|------|
| POST | /shukuyo/career/analyze | 單一相性分析 |
| POST | /shukuyo/career/batch | 批次分析（max 50） |
| POST | /shukuyo/career/comparison | 公司比較（2-5 間，含十年對照） |
| POST | /shukuyo/career/interview-dates | 面試吉日（30 天內） |
| POST | /shukuyo/career/team-matrix | HR 團隊矩陣 |
| POST | /shukuyo/career/headhunter/match | 獵頭配對排名 |

## 測試計畫

1. 直僱模式結果與 Python API 交叉比對（相同 birth_date + company_date）
2. 同一組日期，切換 6 種情境，確認文案不同
3. 批次 20 間公司 benchmark < 2 秒
4. 團隊矩陣 50 人 benchmark < 5 秒
5. drain level 對照表與 Python DRAIN_LEVEL_MAP 一致
6. 三語切換正確（zh-TW/ja/en）

## Checklist

- [x] types.go -- request/response 型別定義
- [x] context.go -- 情境定義
- [x] levels.go -- career 等級映射（與 fortune 不同）
- [x] drain.go -- 消耗度計算
- [x] analyze.go -- 核心分析邏輯
- [x] batch.go -- 批次分析（20 公司 0.26ms）
- [x] comparison.go -- 公司比較（十年對照 + cross-risk）
- [x] interview_dates.go -- 面試吉日
- [x] team_matrix.go -- 團隊矩陣（50 人 4.0ms）
- [x] headhunter.go -- 獵頭配對
- [x] data.go + JSON 文案 -- i18n 資料層（zh-TW/ja/en 三語完整）
- [x] handler.go + router.go -- HTTP 層（6 個 POST 端點）
- [x] career_test.go -- 7 個測試 + 2 個 benchmark 全過
- [x] main.go 路由掛載 /shukuyo/career
- [x] Python API backward compatibility 驗證（3 間公司 group/direction/tier/drain 全部 match）
