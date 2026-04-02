package fortune

import "github.com/go-chi/chi/v5"

// Router returns a chi.Router for the shukuyo fortune endpoints.
func Router() chi.Router {
	r := chi.NewRouter()

	r.Get("/daily/{date}", handleDaily)
	r.Get("/weekly/{date}", handleWeekly)
	r.Get("/monthly/{year}/{month}", handleMonthly)
	r.Get("/yearly/{year}", handleYearly)
	r.Get("/yearly-range", handleYearlyRange)

	return r
}
