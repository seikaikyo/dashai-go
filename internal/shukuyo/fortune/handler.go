package fortune

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/seikaikyo/dashai-go/internal/shukuyo/engine"
	"github.com/seikaikyo/dashai-go/internal/response"
)

func handleDaily(w http.ResponseWriter, r *http.Request) {
	targetStr := chi.URLParam(r, "date")
	birthStr := r.URL.Query().Get("birth_date")
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "zh-TW"
	}

	if targetStr == "" || birthStr == "" {
		response.Err(w, http.StatusBadRequest, "date path and birth_date query are required")
		return
	}

	target, err := engine.ParseDate(targetStr)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "invalid date: "+err.Error())
		return
	}
	birth, err := engine.ParseDate(birthStr)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "invalid birth_date: "+err.Error())
		return
	}

	response.OK(w, CalculateDaily(birth, target, lang))
}

func handleWeekly(w http.ResponseWriter, r *http.Request) {
	targetStr := chi.URLParam(r, "date")
	birthStr := r.URL.Query().Get("birth_date")
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "zh-TW"
	}

	if targetStr == "" || birthStr == "" {
		response.Err(w, http.StatusBadRequest, "date path and birth_date query are required")
		return
	}

	target, err := engine.ParseDate(targetStr)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "invalid date: "+err.Error())
		return
	}
	birth, err := engine.ParseDate(birthStr)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "invalid birth_date: "+err.Error())
		return
	}

	response.OK(w, CalculateWeekly(birth, target, lang))
}

func handleMonthly(w http.ResponseWriter, r *http.Request) {
	yearStr := chi.URLParam(r, "year")
	monthStr := chi.URLParam(r, "month")
	birthStr := r.URL.Query().Get("birth_date")
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "zh-TW"
	}

	if yearStr == "" || monthStr == "" || birthStr == "" {
		response.Err(w, http.StatusBadRequest, "year, month path and birth_date query are required")
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "invalid year")
		return
	}
	month, err := strconv.Atoi(monthStr)
	if err != nil || month < 1 || month > 12 {
		response.Err(w, http.StatusBadRequest, "invalid month (1-12)")
		return
	}
	birth, err := engine.ParseDate(birthStr)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "invalid birth_date: "+err.Error())
		return
	}

	response.OK(w, CalculateMonthly(birth, year, month, lang))
}

func handleYearly(w http.ResponseWriter, r *http.Request) {
	yearStr := chi.URLParam(r, "year")
	birthStr := r.URL.Query().Get("birth_date")
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "zh-TW"
	}

	if yearStr == "" || birthStr == "" {
		response.Err(w, http.StatusBadRequest, "year path and birth_date query are required")
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "invalid year")
		return
	}
	birth, err := engine.ParseDate(birthStr)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "invalid birth_date: "+err.Error())
		return
	}

	response.OK(w, CalculateYearly(birth, year, lang))
}

func handleYearlyRange(w http.ResponseWriter, r *http.Request) {
	birthStr := r.URL.Query().Get("birth_date")
	startStr := r.URL.Query().Get("start_year")
	endStr := r.URL.Query().Get("end_year")
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "zh-TW"
	}

	if birthStr == "" || startStr == "" {
		response.Err(w, http.StatusBadRequest, "birth_date and start_year are required")
		return
	}

	birth, err := engine.ParseDate(birthStr)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "invalid birth_date: "+err.Error())
		return
	}

	startYear, err := strconv.Atoi(startStr)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "invalid start_year")
		return
	}

	endYear := startYear + 9
	if endStr != "" {
		endYear, _ = strconv.Atoi(endStr)
	}

	response.OK(w, CalculateYearlyRange(birth, startYear, endYear, lang))
}

func handleLuckyDaysSummary(w http.ResponseWriter, r *http.Request) {
	birthStr := chi.URLParam(r, "date")
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "zh-TW"
	}

	if birthStr == "" {
		response.Err(w, http.StatusBadRequest, "birth_date path param is required")
		return
	}

	result, err := CalculateLuckyDays(birthStr, lang)
	if err != nil {
		response.Err(w, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(w, result)
}

func handleLuckyDaysCalendar(w http.ResponseWriter, r *http.Request) {
	birthStr := chi.URLParam(r, "date")
	yearStr := chi.URLParam(r, "year")
	monthStr := chi.URLParam(r, "month")
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "zh-TW"
	}

	if birthStr == "" || yearStr == "" || monthStr == "" {
		response.Err(w, http.StatusBadRequest, "date, year, month path params are required")
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "invalid year")
		return
	}
	month, err := strconv.Atoi(monthStr)
	if err != nil || month < 1 || month > 12 {
		response.Err(w, http.StatusBadRequest, "invalid month (1-12)")
		return
	}

	result, err := CalculateLuckyCalendar(birthStr, year, month, lang)
	if err != nil {
		response.Err(w, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(w, result)
}

func handleCalendarMonthly(w http.ResponseWriter, r *http.Request) {
	yearStr := chi.URLParam(r, "year")
	monthStr := chi.URLParam(r, "month")
	birthStr := r.URL.Query().Get("birth_date")
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "zh-TW"
	}

	if yearStr == "" || monthStr == "" || birthStr == "" {
		response.Err(w, http.StatusBadRequest, "year, month path and birth_date query are required")
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "invalid year")
		return
	}
	month, err := strconv.Atoi(monthStr)
	if err != nil || month < 1 || month > 12 {
		response.Err(w, http.StatusBadRequest, "invalid month (1-12)")
		return
	}

	result, err := CalculateCalendarMonth(birthStr, year, month, lang)
	if err != nil {
		response.Err(w, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(w, result)
}

func handleStartupIndustry(w http.ResponseWriter, r *http.Request) {
	dateStr := chi.URLParam(r, "date")
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "zh-TW"
	}
	if dateStr == "" {
		response.Err(w, http.StatusBadRequest, "date path param is required")
		return
	}
	result, err := CalculateIndustryRecommendation(dateStr, lang)
	if err != nil {
		response.Err(w, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(w, result)
}

func handleCareerLuckyDates(w http.ResponseWriter, r *http.Request) {
	var req struct {
		BirthDate string `json:"birth_date"`
		Days      int    `json:"days"`
		Lang      string `json:"lang"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Err(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	if req.BirthDate == "" {
		response.Err(w, http.StatusBadRequest, "birth_date is required")
		return
	}
	if req.Lang == "" {
		req.Lang = "zh-TW"
	}
	result, err := CalculateCareerLuckyDates(req.BirthDate, req.Days, req.Lang)
	if err != nil {
		response.Err(w, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(w, result)
}

func handlePairLuckyDays(w http.ResponseWriter, r *http.Request) {
	date1 := chi.URLParam(r, "date1")
	date2 := chi.URLParam(r, "date2")
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "zh-TW"
	}

	if date1 == "" || date2 == "" {
		response.Err(w, http.StatusBadRequest, "date1 and date2 path params are required")
		return
	}

	result, err := CalculatePairLuckyDays(date1, date2, lang)
	if err != nil {
		response.Err(w, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(w, result)
}
