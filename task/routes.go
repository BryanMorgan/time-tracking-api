package task

import (
	"github.com/bryanmorgan/time-tracking-api/profile"

	"github.com/go-chi/chi"
)

type TaskRouter struct {
	taskService   TaskService
	profileRouter *profile.ProfileRouter
}

func NewRouter(store TaskStore, profileRouter *profile.ProfileRouter) *TaskRouter {
	return &TaskRouter{
		taskService:   NewTaskService(store),
		profileRouter: profileRouter,
	}
}

func (a *TaskRouter) Router() *chi.Mux {
	r := chi.NewRouter()

	// Require authorization/token and valid account
	r.Group(func(r chi.Router) {
		r.Use(profile.TokenHandler)
		r.Use(a.profileRouter.ValidateProfileHandler)
		r.Use(a.profileRouter.ValidateSessionHandler)

		r.Get("/{taskId}", a.getTask)
		r.Get("/all", a.getAllTasks)
		r.Get("/archived", a.getArchivedTasks)
		r.Post("/", a.saveTask)
		r.Put("/", a.updateTask)
		r.Put("/archive", a.archiveTaskHandler)
		r.Put("/restore", a.restoreTaskHandler)
		r.Delete("/", a.deleteTaskHandler)
	})

	return r
}
