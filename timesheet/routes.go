package timesheet

import (
	"github.com/bryanmorgan/time-tracking-api/profile"

	"github.com/go-chi/chi"
)

type TimeRouter struct {
	timeService   TimeService
	profileRouter *profile.ProfileRouter
}

func NewRouter(store TimeStore, profileRouter *profile.ProfileRouter) *TimeRouter {
	return &TimeRouter{
		timeService:   NewTimeService(store),
		profileRouter: profileRouter,
	}
}

func (a *TimeRouter) Router() *chi.Mux {
	r := chi.NewRouter()

	// Require authorization/token and valid account
	r.Group(func(r chi.Router) {
		r.Use(profile.TokenHandler)
		r.Use(a.profileRouter.ValidateProfileHandler)
		r.Use(a.profileRouter.ValidateSessionHandler)

		// Time Entries
		r.Get("/week", a.getTimeEntriesForWeek)
		r.Get("/week/{startDate}", a.getTimeEntriesForWeek)

		r.Put("/", a.updateTimeEntries)
		r.Post("/project/week", a.addProjectToWeek)
		r.Delete("/project/week", a.deleteProjectForWeek)
	})

	return r
}
