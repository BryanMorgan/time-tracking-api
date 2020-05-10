package profile

import (
	"github.com/go-chi/chi"
)

type ProfileRouter struct {
	profileService ProfileService
}

// Returns a configured authentication profileService
func NewRouter(store ProfileStore) *ProfileRouter {
	return &ProfileRouter{
		profileService: NewProfileService(store),
	}
}

func (pr *ProfileRouter) AuthenticationRouter() *chi.Mux {
	r := chi.NewRouter()

	r.Post("/login", pr.loginHandler)
	r.Post("/forgot", pr.forgotPasswordHandler)
	r.Post("/forgot/validate", pr.validateForgotTokenHandler)
	r.Put("/setup", pr.setupNewUserAccountHandler)

	// Require authorization/token
	r.Group(func(r chi.Router) {
		r.Use(TokenHandler)

		r.Post("/token", pr.validateTokenHandler)
		r.Post("/logout", pr.logoutHandler)

		r.Group(func(r chi.Router) {
			r.Use(pr.ValidateProfileHandler)
			r.Use(pr.ValidateSessionHandler)
		})
	})

	return r
}

func (pr *ProfileRouter) ProfileRouter() *chi.Mux {
	r := chi.NewRouter()

	// Require authorization/token
	r.Group(func(r chi.Router) {
		r.Use(TokenHandler)
		r.Use(pr.ValidateProfileHandler)
		r.Use(pr.ValidateSessionHandler)

		r.Get("/", pr.getProfileHandler)
		r.Put("/", pr.updateProfileHandler)
		r.Put("/password", pr.updatePasswordHandler)

	})

	return r
}

func (pr *ProfileRouter) AccountRouter() *chi.Mux {
	r := chi.NewRouter()

	r.Post("/", pr.createAccountHandler)

	// Require authorization token
	r.Group(func(r chi.Router) {
		r.Use(TokenHandler)

		// Require valid profile
		r.Group(func(r chi.Router) {
			r.Use(pr.ValidateProfileHandler)
			r.Use(pr.ValidateSessionHandler)

			// Admin-level access
			r.Group(func(r chi.Router) {
				r.Use(pr.AdminPermissionHandler)
				r.Put("/", pr.updateAccountHandler)
				r.Get("/", pr.getAccountHandler)

				// Require a valid, active account
				r.Group(func(r chi.Router) {
					r.Use(pr.requireValidAccount)

					r.Delete("/", pr.closeAccountHandler)
					r.Get("/users", pr.getUsersHandler)
					r.Post("/user", pr.addUserHandler)
					r.Delete("/user", pr.removeUserHandler)
				})
			})
		})

	})

	return r
}
