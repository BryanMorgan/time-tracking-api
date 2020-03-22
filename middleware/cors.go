package middleware

import (
	"net/http"

	"github.com/bryanmorgan/time-tracking-api/logger"

	"github.com/go-chi/cors"
	"github.com/spf13/viper"
)

func CorsHandler(next http.Handler) http.Handler {
	corsDomains := viper.GetStringSlice("cors.domains")
	corsMaxAge := viper.GetInt("cors.maxAge")

	c := cors.New(cors.Options{
		AllowedOrigins:   corsDomains,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           corsMaxAge,
	})

	logger.Log.Debug("CORS enabled")

	return c.Handler(next)
}
