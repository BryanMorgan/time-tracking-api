package reporting

import "database/sql"

type ClientReport struct {
	ClientId         int             `json:"-" db:"client_id"`
	ClientName       string          `json:"-" db:"client_name"`
	NonBillableHours sql.NullFloat64 `json:"-" db:"non_billable_hours"`
	BillableHours    sql.NullFloat64 `json:"-" db:"billable_hours"`
	BillableTotal    sql.NullFloat64 `json:"-" db:"billable_total"`
}

type ProjectReport struct {
	ProjectId        int             `json:"-" db:"project_id"`
	ProjectName      string          `json:"-" db:"project_name"`
	ClientName       string          `json:"-" db:"client_name"`
	NonBillableHours sql.NullFloat64 `json:"-" db:"non_billable_hours"`
	BillableHours    sql.NullFloat64 `json:"-" db:"billable_hours"`
	BillableTotal    sql.NullFloat64 `json:"-" db:"billable_total"`
}

type TaskReport struct {
	TaskId           int             `json:"-" db:"task_id"`
	ClientId         int             `json:"-" db:"client_id"`
	TaskName         string          `json:"-" db:"task_name"`
	NonBillableHours sql.NullFloat64 `json:"-" db:"non_billable_hours"`
	BillableHours    sql.NullFloat64 `json:"-" db:"billable_hours"`
	BillableTotal    sql.NullFloat64 `json:"-" db:"billable_total"`
}

type PersonReport struct {
	ProfileId        int             `json:"-" db:"profile_id"`
	FirstName        string          `json:"-" db:"first_name"`
	LastName         string          `json:"-" db:"last_name"`
	NonBillableHours sql.NullFloat64 `json:"-" db:"non_billable_hours"`
	BillableHours    sql.NullFloat64 `json:"-" db:"billable_hours"`
	BillableTotal    sql.NullFloat64 `json:"-" db:"billable_total"`
}
