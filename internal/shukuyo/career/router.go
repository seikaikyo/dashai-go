package career

import "github.com/go-chi/chi/v5"

// Router returns a chi.Router for the shukuyo career endpoints.
func Router() chi.Router {
	r := chi.NewRouter()

	r.Post("/analyze", handleAnalyze)
	r.Post("/batch", handleBatch)
	r.Post("/comparison", handleComparison)
	r.Post("/interview-dates", handleInterviewDates)
	r.Post("/team-matrix", handleTeamMatrix)
	r.Post("/headhunter/match", handleHeadhunterMatch)
	r.Post("/104/company-jobs", handle104Jobs)
	r.Post("/company-search", handleCompanySearch)

	return r
}
