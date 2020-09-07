package reporting

import (
	"github.com/bryanmorgan/time-tracking-api/profile"

	"github.com/go-chi/chi"
)

type ReportingRouter struct {
	ReportingService ReportingService
	profileRouter    *profile.ProfileRouter
}

func NewRouter(store ReportingStore, profileRouter *profile.ProfileRouter) *ReportingRouter {
	return &ReportingRouter{
		ReportingService: NewReportingService(store),
		profileRouter:    profileRouter,
	}
}

func (a *ReportingRouter) Router() *chi.Mux {
	r := chi.NewRouter()

	// Require authorization/token and valid account
	r.Group(func(r chi.Router) {
		r.Use(profile.TokenHandler)
		r.Use(a.profileRouter.ValidateProfileHandler)
		r.Use(a.profileRouter.ValidateSessionHandler)

		r.Get("/time/client", a.getTimeByClient)
		r.Get("/time/project", a.getTimeByProject)
		r.Get("/time/task", a.getTimeByTask)
		r.Get("/time/person", a.getTimeByPerson)
		r.Get("/time/export/client", a.exportTimeByClient)
		r.Get("/time/export/project", a.exportTimeByProject)
		r.Get("/time/export/task", a.exportTimeByTask)
		r.Get("/time/export/person", a.exportTimeByPerson)

	})

	return r
}
