package client

import (
	"github.com/bryanmorgan/time-tracking-api/profile"
	"github.com/bryanmorgan/time-tracking-api/timesheet"

	"github.com/go-chi/chi"
)

type ClientRouter struct {
	clientService ClientService
	profileRouter *profile.ProfileRouter
}

// Returns a configured authentication profileService
func NewRouter(store ClientStore, timeStore timesheet.TimeStore, profileRouter *profile.ProfileRouter) *ClientRouter {
	return &ClientRouter{
		clientService: NewClientService(store, timeStore),
		profileRouter: profileRouter,
	}
}

func (a *ClientRouter) Router() *chi.Mux {
	r := chi.NewRouter()

	// Require authorization/token
	r.Group(func(r chi.Router) {
		r.Use(profile.TokenHandler)
		r.Use(a.profileRouter.ValidateProfileHandler)
		r.Use(a.profileRouter.ValidateSessionHandler)

		// Client
		r.Get("/{clientId}", a.getClientHandler)
		r.Get("/all", a.getAllClientsHandler)
		r.Get("/archived", a.getArchivedClientsHandler)
		r.Post("/", a.createClientHandler)
		r.Put("/", a.updateClientHandler)
		r.Put("/archive", a.archiveClientHandler)
		r.Put("/restore", a.restoreClientHandler)
		r.Delete("/", a.deleteClientHandler)

		// Project
		r.Route("/project", func(r chi.Router) {
			r.Get("/{projectId}", a.getProjectHandler)
			r.Get("/all", a.getAllProjectsHandler)
			r.Get("/archived", a.getArchivedProjectsHandler)
			r.Post("/", a.createProjectHandler)
			r.Put("/", a.updateProjectHandler)
			r.Put("/archive", a.archiveProjectHandler)
			r.Put("/restore", a.restoreProjectHandler)
			r.Delete("/", a.deleteProjectHandler)
			r.Post("/copy/last/week", a.copyProjectsFromLastWeek)
		})

	})

	return r
}
