package middleware

import (
	"github.com/bryanmorgan/time-tracking-api/logger"
	"github.com/jmoiron/sqlx"
	"net/http"
)

func Ping(db *sqlx.DB) http.HandlerFunc {
	fn := func (w http.ResponseWriter, r * http.Request) {
		if r.Method != "GET" {
			writeOutput(w, http.StatusBadRequest, "invalid method")
			return
		}

		err := db.Ping()
		if err != nil {
			logger.Log.Error("Failed to ping Postgres database: " + err.Error())
			writeOutput(w, http.StatusInternalServerError, "error")
			return
		}

		writeOutput(w, http.StatusOK, "ok")
	}
	return fn
}

func writeOutput(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(statusCode)
	if _, err := w.Write([]byte(message)); err != nil {
		logger.Log.Error("Failed to write ping bytes: " + err.Error())
	}
}
