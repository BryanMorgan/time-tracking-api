// +build integration

package integration_test

import (
	"github.com/go-chi/chi"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
	"os"
	"testing"

	"github.com/bryanmorgan/time-tracking-api/app"
)

var db *sqlx.DB
var router *chi.Mux

func TestMain(m *testing.M) {
	os.Setenv("GO_ENV", "test")

	viper.AddConfigPath("../config")
	server := app.NewApp()
	router = server.Router
	db = server.DB

	code := m.Run()

	db.Close()
	os.Exit(code)
}

