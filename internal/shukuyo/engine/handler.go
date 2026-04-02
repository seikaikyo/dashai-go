package engine

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/seikaikyo/dashai-go/internal/response"
)

type compatibilityRequest struct {
	Date1 string `json:"date1"`
	Date2 string `json:"date2"`
}

func handleMansion(w http.ResponseWriter, r *http.Request) {
	birthDateStr := r.URL.Query().Get("birth_date")
	if birthDateStr == "" {
		response.Err(w, http.StatusBadRequest, "birth_date is required")
		return
	}

	bd, err := ParseDate(birthDateStr)
	if err != nil {
		response.Err(w, http.StatusBadRequest, err.Error())
		return
	}

	response.OK(w, GetMansion(bd))
}

func handleKuyou(w http.ResponseWriter, r *http.Request) {
	birthDateStr := r.URL.Query().Get("birth_date")
	yearStr := r.URL.Query().Get("year")

	if birthDateStr == "" {
		response.Err(w, http.StatusBadRequest, "birth_date is required")
		return
	}
	if yearStr == "" {
		response.Err(w, http.StatusBadRequest, "year is required")
		return
	}

	bd, err := ParseDate(birthDateStr)
	if err != nil {
		response.Err(w, http.StatusBadRequest, err.Error())
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "year must be an integer")
		return
	}

	response.OK(w, GetKuyouStar(bd, year))
}

func handleCompatibility(w http.ResponseWriter, r *http.Request) {
	var req compatibilityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Err(w, http.StatusBadRequest, err.Error())
		return
	}

	d1, err := ParseDate(req.Date1)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "date1: "+err.Error())
		return
	}
	d2, err := ParseDate(req.Date2)
	if err != nil {
		response.Err(w, http.StatusBadRequest, "date2: "+err.Error())
		return
	}

	response.OK(w, Compatibility(d1, d2))
}

func handleRelation(w http.ResponseWriter, r *http.Request) {
	idx1Str := r.URL.Query().Get("idx1")
	idx2Str := r.URL.Query().Get("idx2")

	if idx1Str == "" || idx2Str == "" {
		response.Err(w, http.StatusBadRequest, "idx1 and idx2 are required")
		return
	}

	idx1, err := strconv.Atoi(idx1Str)
	if err != nil || idx1 < 0 || idx1 > 26 {
		response.Err(w, http.StatusBadRequest, "idx1 must be 0-26")
		return
	}
	idx2, err := strconv.Atoi(idx2Str)
	if err != nil || idx2 < 0 || idx2 > 26 {
		response.Err(w, http.StatusBadRequest, "idx2 must be 0-26")
		return
	}

	response.OK(w, GetRelation(idx1, idx2))
}

func handleSpecialDay(w http.ResponseWriter, r *http.Request) {
	dateStr := r.URL.Query().Get("date")
	birthDateStr := r.URL.Query().Get("birth_date")

	if dateStr == "" {
		response.Err(w, http.StatusBadRequest, "date is required")
		return
	}

	d, err := ParseDate(dateStr)
	if err != nil {
		response.Err(w, http.StatusBadRequest, err.Error())
		return
	}

	dayMansionIdx := DayMansionIndex(d)
	specialDay := CheckSpecialDay(d, dayMansionIdx)
	ryouhan := CheckRyouhanPeriod(d)

	result := map[string]any{
		"date":              dateStr,
		"day_mansion_index": dayMansionIdx,
		"day_mansion":       Mansions27[dayMansionIdx].NameJP,
		"special_day":       specialDay,
		"ryouhan":           ryouhan,
	}

	if birthDateStr != "" {
		bd, err := ParseDate(birthDateStr)
		if err == nil {
			birthIdx, _ := MansionIndexFromDate(bd)
			result["sanki"] = GetSankiPosition(birthIdx, dayMansionIdx)
		}
	}

	response.OK(w, result)
}
