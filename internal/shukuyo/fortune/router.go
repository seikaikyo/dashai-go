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

	// Lucky days
	r.Get("/lucky-days/summary/{date}", handleLuckyDaysSummary)
	r.Get("/lucky-days/calendar/{date}/{year}/{month}", handleLuckyDaysCalendar)
	r.Get("/lucky-days/pair/{date1}/{date2}", handlePairLuckyDays)

	// Calendar
	r.Get("/calendar/{year}/{month}", handleCalendarMonthly)

	return r
}
