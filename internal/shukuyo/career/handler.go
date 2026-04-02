package career

import (
	"encoding/json"
	"net/http"

	"github.com/seikaikyo/dashai-go/internal/response"
)

func handleAnalyze(w http.ResponseWriter, r *http.Request) {
	var req AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Err(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	if req.BirthDate == "" || req.CompanyDate == "" {
		response.Err(w, http.StatusBadRequest, "birth_date and company_date are required")
		return
	}

	result, err := Analyze(req)
	if err != nil {
		response.Err(w, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(w, result)
}

func handleBatch(w http.ResponseWriter, r *http.Request) {
	var req BatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Err(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	if req.BirthDate == "" || len(req.Companies) == 0 {
		response.Err(w, http.StatusBadRequest, "birth_date and companies are required")
		return
	}
	if len(req.Companies) > 50 {
		response.Err(w, http.StatusBadRequest, "maximum 50 companies per request")
		return
	}

	result, err := BatchAnalyze(req)
	if err != nil {
		response.Err(w, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(w, result)
}

func handleComparison(w http.ResponseWriter, r *http.Request) {
	var req ComparisonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Err(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	if req.BirthDate == "" || len(req.Companies) < 2 {
		response.Err(w, http.StatusBadRequest, "birth_date and at least 2 companies are required")
		return
	}
	if len(req.Companies) > 5 {
		response.Err(w, http.StatusBadRequest, "maximum 5 companies for comparison")
		return
	}
	if req.EndYear-req.StartYear > 10 {
		response.Err(w, http.StatusBadRequest, "maximum 10-year range")
		return
	}

	result, err := Compare(req)
	if err != nil {
		response.Err(w, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(w, result)
}

func handleInterviewDates(w http.ResponseWriter, r *http.Request) {
	var req InterviewDatesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Err(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	if req.BirthDate == "" || req.CompanyDate == "" {
		response.Err(w, http.StatusBadRequest, "birth_date and company_date are required")
		return
	}

	result, err := InterviewDates(req)
	if err != nil {
		response.Err(w, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(w, result)
}

func handleTeamMatrix(w http.ResponseWriter, r *http.Request) {
	var req TeamMatrixRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Err(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	if req.CompanyDate == "" || len(req.Candidates) == 0 {
		response.Err(w, http.StatusBadRequest, "company_date and candidates are required")
		return
	}
	if len(req.Candidates) > 50 {
		response.Err(w, http.StatusBadRequest, "maximum 50 candidates per request")
		return
	}

	result, err := TeamMatrix(req)
	if err != nil {
		response.Err(w, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(w, result)
}

func handleHeadhunterMatch(w http.ResponseWriter, r *http.Request) {
	var req HeadhunterMatchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Err(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	if req.CandidateDate == "" || len(req.Companies) == 0 {
		response.Err(w, http.StatusBadRequest, "candidate_date and companies are required")
		return
	}

	result, err := HeadhunterMatch(req)
	if err != nil {
		response.Err(w, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(w, result)
}
