package app

import (
	"context"
	"fmt"
	"github.com/bryanmorgan/time-tracking-api/api"
	"github.com/bryanmorgan/time-tracking-api/client"
	"github.com/bryanmorgan/time-tracking-api/config"
	"github.com/bryanmorgan/time-tracking-api/database"
	"github.com/bryanmorgan/time-tracking-api/logger"
	"github.com/bryanmorgan/time-tracking-api/middleware"
	"github.com/bryanmorgan/time-tracking-api/profile"
	"github.com/bryanmorgan/time-tracking-api/reporting"
	"github.com/bryanmorgan/time-tracking-api/task"
	"github.com/bryanmorgan/time-tracking-api/timesheet"
	"github.com/bryanmorgan/time-tracking-api/version"
	"github.com/go-chi/chi"
	cmiddleware "github.com/go-chi/chi/middleware"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

type App struct {
	Router *chi.Mux
	DB     *sqlx.DB
}

func NewApp() *App {
	config.InitConfig()
	logger.InitLogger()
	db := database.InitPostgres()

	logger.Log.Info("Application started")
	return &App{
		Router: newRouter(db),
		DB:     db,
	}
}

func (a *App) Run() {
	runServers(a.Router, a.DB)
}

func newRouter(db *sqlx.DB) *chi.Mux {
	// Create database stores
	profileStore := profile.NewProfileAccountStore(db)
	clientStore := client.NewClientStore(db)
	timeStore := timesheet.NewTimeStore(db)
	taskStore := task.NewTaskStore(db)
	reportingStore := reporting.NewReportingStore(db)

	// Create API service routers
	profileRouter := profile.NewRouter(profileStore)
	clientRouter := client.NewRouter(clientStore, timeStore, profileRouter)
	timeRouter := timesheet.NewRouter(timeStore, profileRouter)
	taskRouter := task.NewRouter(taskStore, profileRouter)
	reportingRouter := reporting.NewRouter(reportingStore, profileRouter)

	r := chi.NewRouter()

	r.Use(middleware.PanicHandler)
	r.Use(cmiddleware.RealIP)
	r.Use(logger.RequestLoggingHandler)

	if viper.GetInt("application.delayRequestMilliSeconds") > 0 {
		// Allow debugging delay
		r.Use(addDelay)
	}

	if viper.GetBool("cors.enabled") {
		r.Use(middleware.CorsHandler)
	}

	r.Route("/api", func(r chi.Router) {
		r.Mount("/auth", profileRouter.AuthenticationRouter())
		r.Mount("/profile", profileRouter.ProfileRouter())
		r.Mount("/account", profileRouter.AccountRouter())
		r.Mount("/client", clientRouter.Router())
		r.Mount("/time", timeRouter.Router())
		r.Mount("/task", taskRouter.Router())
		r.Mount("/report", reportingRouter.Router())
	})

	r.Get("/_ping", middleware.Ping(db))

	// Handle unexpected conditions
	r.NotFound(notFoundHandler)
	r.MethodNotAllowed(methodNotAllowedHandler)

	return r
}

func runServers(router *chi.Mux, db *sqlx.DB) {
	hostname := viper.GetString("application.hostname")
	port := viper.GetInt("application.port")

	if hostname == "" || port <= 0 {
		log.Fatal("Missing hostname or port in config settings")
	}

	// pprof profiling server exposed on an internal (non-public) port
	if viper.GetBool("application.profileEnabled") {
		profileHostname := viper.GetString("application.profileHostname")
		profilePort := viper.GetInt("application.profilePort")

		go func() {
			profileAddress := fmt.Sprintf("%s:%d", profileHostname, profilePort)
			logger.Log.Info("Start profile server: " + profileAddress)
			log.Fatal(http.ListenAndServe(profileAddress, http.DefaultServeMux))
		}()
	}

	appAddress := fmt.Sprintf("%s:%d", hostname, port)
	server := &http.Server{
		Addr:         appAddress,
		Handler:      router,
		ReadTimeout:  5 * time.Minute,
		WriteTimeout: 10 * time.Minute,
	}

	// Signal handler to gracefully shutdown
	go func() {
		// Handle SIGINT and SIGTERM and SIGQUIT
		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

		killSignal := <-interrupt
		switch killSignal {
		case os.Kill:
			log.Print("Got SIGKILL...")
		case os.Interrupt:
			log.Print("Got SIGINT...")
		case syscall.SIGTERM:
			log.Print("Got SIGTERM...")
		}

		// Stop the service gracefully.
		logger.Log.Info("Graceful shutdown...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			logger.Log.Fatal("Could not shutdown: " + err.Error())
		}
		logger.Log.Info("HTTP Server shutdown")

		if db != nil {
			if err := db.Close(); err != nil {
				logger.Log.Error("Failed to close database: " + err.Error())
			}
		}
		logger.Log.Info("Database closed")
		err := logger.Log.Sync()
		if err != nil {
			log.Println("Failed to sync logger: " + err.Error())
		}
		os.Exit(1)
	}()

	// Public HTTP server
	startMessage := fmt.Sprintf("[%s] app server started on %s  [release: %s] [build: %s] [commit: %s]", viper.GetString("GO_ENV"), appAddress, version.Release, version.BuildTime, version.Commit)
	logger.Log.Info(startMessage)
	log.Fatal(server.ListenAndServe())
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	warningMessage := api.NewError("Not Found", r.RequestURI, api.NotFound)
	api.ErrorJson(w, warningMessage, http.StatusNotFound)
}

func methodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	warningMessage := api.NewError(r.RequestURI, r.Method, api.MethodNotAllowed)
	api.ErrorJson(w, warningMessage, http.StatusMethodNotAllowed)
}

func addDelay(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "OPTIONS" {
			delay := viper.GetInt("application.delayRequestMilliSeconds")
			logger.Log.Debug("Delaying request " + r.RequestURI + " for " + strconv.Itoa(delay) + " ms")
			time.Sleep(time.Duration(delay) * time.Millisecond)
		}
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
