package engine

import (
	"github.com/go-chi/chi/v5"
)

// Router returns a chi.Router for the shukuyo engine endpoints.
func Router() chi.Router {
	r := chi.NewRouter()

	r.Get("/mansion", handleMansion)
	r.Get("/mansion/{date}", handleMansionByDate)
	r.Get("/mansions", handleMansions)
	r.Get("/relations", handleRelations)
	r.Get("/metadata", handleMetadata)
	r.Get("/kuyou", handleKuyou)
	r.Post("/compatibility", handleCompatibility)
	r.Post("/compatibility-batch", handleCompatibilityBatch)
	r.Get("/compatibility-finder/{date}", handleCompatibilityFinder)
	r.Get("/relation", handleRelation)
	r.Get("/special-day", handleSpecialDay)

	return r
}
