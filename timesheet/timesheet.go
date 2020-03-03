package timesheet

import (
	"time"
)

type TimeEntry struct {
	Day         time.Time `json:"-"`
	Hours       float64   `json:"-"`
	AccountId   int       `json:"-" db:"account_id"`
	ProfileId   int       `json:"-" db:"profile_id"`
	ProjectId   int       `json:"-" db:"project_id"`
	TaskId      int       `json:"-" db:"task_id"`
	ClientName  string    `json:"-" db:"client_name"`
	ProjectName string    `json:"-" db:"project_name"`
	TaskName    string    `json:"-" db:"task_name"`
}

type Timesheet struct {
	Days      [7]TimeEntry `json:"-"`
	StartDate time.Time    `json:"-"`
	Status    string       `json:"-"`
}
