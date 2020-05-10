package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/bryanmorgan/time-tracking-api/logger"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

var NoRowAffectedError = errors.New("no rows affected")

func timeTrack(start time.Time, name string) {
	logger.Log.Info(name, logger.Duration("duration", time.Since(start)))
}

func InitPostgres() *sqlx.DB {
	start := time.Now()
	defer timeTrack(start, "Postgres Startup")

	username := viper.GetString("postgres.username")
	password := viper.GetString("postgres.password")
	database := viper.GetString("postgres.database")
	sslmode := viper.GetString("postgres.sslmode")
	timeout := viper.GetInt("postgres.timeout")
	host := viper.GetString("postgres.master.host")
	port := viper.GetInt("postgres.master.port")

	dataSource := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%d sslmode=%s connect_timeout=%d",
		username, password, database, host, port, sslmode, timeout)
	db, err := sqlx.Connect("postgres", dataSource)
	if err != nil {
		log.Panicf("Initialize database failed: %s", err.Error())
		return nil
	}

	connectionLifetime := viper.GetDuration("postgres.connectLifetime")
	maxIdle := viper.GetInt("postgres.maxIdle")
	maxOpen := viper.GetInt("postgres.maxOpen")
	db.SetConnMaxLifetime(time.Minute * connectionLifetime)
	db.SetMaxIdleConns(maxIdle)
	db.SetMaxOpenConns(maxOpen)

	logger.Log.Info("Postgres Configuration",
		logger.Int("maxIdle", maxIdle),
		logger.Int("maxOpen", maxOpen),
		logger.Duration("connectionMax", time.Minute*connectionLifetime),
		logger.Int("timeout", timeout))

	logger.Log.Info("Postgres Connect",
		logger.Duration("duration", time.Since(start)),
		logger.String("host", host),
		logger.Int("port", port),
		logger.String("database", database))

	start = time.Now()
	err = db.Ping()
	if err != nil {
		log.Panicf("Failed to ping Postgres database: %s", err.Error())
		return nil
	}

	logger.Log.Info("Postgres Ping DB", logger.Duration("duration", time.Since(start)))

	return db
}

func CloseRows(rows *sqlx.Rows) {
	if err := rows.Close(); err != nil {
		logger.Log.Error("Failed to close rows: " + err.Error())
	}
}

func RollbackTransaction(tx *sql.Tx) {
	if err := tx.Rollback(); err != nil {
		logger.Log.Error("Failed to rollback transaction: " + err.Error())
	}
}
