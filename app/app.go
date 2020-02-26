package app

import (
	"github.com/bryanmorgan/time-tracking-api/config"
	"github.com/bryanmorgan/time-tracking-api/database"
	"github.com/bryanmorgan/time-tracking-api/logger"
	"github.com/jmoiron/sqlx"
)

type App struct {
	DB     *sqlx.DB
}

func NewApp() *App {
	config.InitConfig()
	logger.InitLogger()
	db := database.InitPostgres()

	logger.Log.Info("Application started")
	return &App{
		DB:     db,
	}
}

func (a *App) Run() {
}
