package app

import (
	"github.com/bryanmorgan/time-tracking-api/config"
	"github.com/bryanmorgan/time-tracking-api/logger"
)

type App struct {
}

func NewApp() *App {
	config.InitConfig()
	logger.InitLogger()

	logger.Log.Info("Application started")
	return &App{}
}

func (a *App) Run() {
}
