package task

import (
	"database/sql"
)

type Task struct {
	TaskId          int             `json:"-" db:"task_id"`
	AccountId       int             `json:"-" db:"account_id"`
	Name            string          `json:"-" db:"task_name"`
	DefaultRate     sql.NullFloat64 `json:"-" db:"default_rate"`
	DefaultBillable bool            `json:"-" db:"default_billable"`
	Common          bool            `json:"-"`
	TaskActive      bool            `json:"-" db:"task_active"`
}

type ProjectTask struct {
	Task
	ProjectId     int             `json:"-" db:"project_id"`
	Rate          sql.NullFloat64 `json:"-"`
	Billable      bool            `json:"-"`
	ProjectActive bool            `json:"-" db:"project_active"`
}
