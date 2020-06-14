package timesheet

import (
	"errors"
	"time"

	"github.com/bryanmorgan/time-tracking-api/config"
	"github.com/bryanmorgan/time-tracking-api/database"
	"github.com/bryanmorgan/time-tracking-api/logger"

	"github.com/jmoiron/sqlx"
)

// Compile Only: ensure interface is implemented
var _ TimeStore = &TimeData{}

type TimeStore interface {
	GetTimeEntriesForRange(profileId int, accountId int, start time.Time, end time.Time) ([]*TimeEntry, error)

	SaveOrUpdateTimeEntries(entries []*TimeEntry) error
	UpdateTimeEntries(entries []*TimeEntry) error

	AddInitialProjectTimeEntries(profileId int, accountId int, start time.Time, end time.Time, projectId int, taskId int) error

	DeleteProjectForDates(profileId int, accountId int, projectId int, taskId int, start time.Time, end time.Time) error
}

// ProfileData implements database operations for user profiles
type TimeData struct {
	db *sqlx.DB
}

func NewTimeStore(db *sqlx.DB) TimeStore {
	return &TimeData{
		db: db,
	}
}

func (c *TimeData) SaveOrUpdateTimeEntries(entries []*TimeEntry) error {
	if entries == nil {
		logger.Log.Warn("No entries to save or update. Do nothing.")
		return nil
	}

	tx, err := c.db.Begin()
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.Day.IsZero() {
			database.RollbackTransaction(tx)
			return errors.New("invalid time entry day: " + entry.Day.String())
		}

		if entry.Hours < 0 {
			entry.Hours = 0.0
		}

		upsertSql := `
		INSERT INTO time (account_id, profile_id, project_id, task_id, day, hours)
 				  VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (account_id, profile_id, project_id, task_id, day)
		DO UPDATE SET hours = $6
		WHERE time.account_id = $1
		  AND time.profile_id = $2
		  AND time.project_id = $3
		  AND time.task_id = $4
		  AND time.day = $5
		`

		results, err := c.db.Exec(upsertSql, entry.AccountId, entry.ProfileId, entry.ProjectId, entry.TaskId, entry.Day.Format(config.ISOShortDateFormat), entry.Hours)
		if err != nil {
			database.RollbackTransaction(tx)
			return err
		}

		rows, err := results.RowsAffected()
		if err != nil {
			database.RollbackTransaction(tx)
			return err
		}

		if rows == 0 {
			database.RollbackTransaction(tx)
			return database.NoRowAffectedError
		}
	}

	if err := tx.Commit(); err != nil {
		logger.Log.Error("Error in commit transaction: " + err.Error())
		return err
	}

	return nil
}

func (c *TimeData) UpdateTimeEntries(entries []*TimeEntry) error {
	if entries == nil {
		logger.Log.Warn("No time entries to update")
		return nil
	}

	tx, err := c.db.Begin()
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.Day.IsZero() {
			database.RollbackTransaction(tx)
			return errors.New("invalid time entry day: " + entry.Day.String())
		}

		if entry.Hours < 0 {
			entry.Hours = 0.0
		}

		if entry.Hours > 9999 {
			logger.Log.Warn("Time entry for hours is too big. Capping at 9999.")
			entry.Hours = 9999
		}

		updateSql := `
			UPDATE time SET hours = $6
			WHERE time.account_id = $1
			  AND time.profile_id = $2
			  AND time.project_id = $3
			  AND time.task_id = $4
			  AND time.day = $5
		`

		results, err := c.db.Exec(updateSql, entry.AccountId, entry.ProfileId, entry.ProjectId, entry.TaskId, entry.Day.Format(config.ISOShortDateFormat), entry.Hours)
		if err != nil {
			database.RollbackTransaction(tx)
			return err
		}

		rows, err := results.RowsAffected()
		if err != nil {
			database.RollbackTransaction(tx)
			return err
		}

		if rows == 0 {
			logger.Log.Warn("[Update-Insert] Update time entry failed: trying INSERT")

			insertSql := `
				INSERT INTO time(account_id, profile_id, project_id, task_id, day, hours)
					  VALUES ($1, $2, $3, $4, $5, $6)`
			results, err = c.db.Exec(insertSql, entry.AccountId, entry.ProfileId, entry.ProjectId, entry.TaskId, entry.Day.Format(config.ISOShortDateFormat), entry.Hours)
			if err != nil {
				logger.Log.Error("Update failed and insert failed: " + err.Error())
				database.RollbackTransaction(tx)
				return err
			}

			rows, err := results.RowsAffected()
			if err != nil {
				database.RollbackTransaction(tx)
				return err
			}

			if rows == 0 {
				database.RollbackTransaction(tx)
				return database.NoRowAffectedError
			}
		}
	}

	if err := tx.Commit(); err != nil {
		logger.Log.Error("Error in commit transaction: " + err.Error())
		return err
	}

	return nil
}

// Add 0.0 values for all the days between start and end for the project/task
func (c *TimeData) AddInitialProjectTimeEntries(profileId int, accountId int, start time.Time, end time.Time, projectId int, taskId int) error {
	tx, err := c.db.Begin()
	if err != nil {
		return err
	}

	for day := start; day.Before(end) || day.Equal(end); day = day.AddDate(0, 0, 1) {
		insertSql := `
		INSERT INTO time (account_id, profile_id, project_id, task_id, day, hours)
 				  VALUES ($1, $2, $3, $4, $5, 0.0)`

		results, err := c.db.Exec(insertSql, accountId, profileId, projectId, taskId, day.Format(config.ISOShortDateFormat))
		if err != nil {
			database.RollbackTransaction(tx)
			return err
		}

		rows, err := results.RowsAffected()
		if err != nil {
			database.RollbackTransaction(tx)
			return err
		}

		if rows == 0 {
			database.RollbackTransaction(tx)
			return database.NoRowAffectedError
		}
	}

	if err := tx.Commit(); err != nil {
		logger.Log.Error("Error in commit transaction: " + err.Error())
		return err
	}

	return nil
}

func (c *TimeData) GetTimeEntriesForRange(profileId int, accountId int, start time.Time, end time.Time) ([]*TimeEntry, error) {
	sqlStatement := `
		SELECT
		  c.client_name,
		  p.project_name,
		  k.task_name,
		  t.hours,
		  t.day,
		  p.project_id,
		  t.task_id
		FROM time t,
		     project p,
		     client c,
		     task k
		WHERE c.client_id = p.client_id
		  AND t.project_id = p.project_id
		  AND t.task_id = k.task_id
		  AND t.account_id = $1
		  AND t.profile_id = $2
		  AND t.day >= $3
		  AND t.day <= $4
		ORDER BY t.day`

	rows, err := c.db.Queryx(sqlStatement, accountId, profileId, start.Format(config.ISOShortDateFormat), end.Format(config.ISOShortDateFormat))
	if err != nil {
		return nil, err
	}
	defer database.CloseRows(rows)

	var timeEntries []*TimeEntry
	for rows.Next() {
		var timeEntry TimeEntry
		err := rows.StructScan(&timeEntry)
		if err != nil {
			return nil, err
		}
		timeEntries = append(timeEntries, &timeEntry)
	}

	return timeEntries, nil
}

func (c *TimeData) DeleteProjectForDates(profileId int, accountId int, projectId int, taskId int, start time.Time, end time.Time) error {
	sql := `
		DELETE FROM time
		WHERE profile_id=$1
          AND account_id=$2
          AND project_id=$3
          AND task_id=$4
          AND day >= $5
          AND day <= $6
	`
	results, err := c.db.Exec(sql, profileId, accountId, projectId, taskId, start.Format(config.ISOShortDateFormat), end.Format(config.ISOShortDateFormat))
	if err != nil {
		return err
	}

	rows, err := results.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return database.NoRowAffectedError
	}

	return nil
}
