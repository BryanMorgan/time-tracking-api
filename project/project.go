package project

import (
	"database/sql"

	"github.com/bryanmorgan/time-tracking-api/client"
	"github.com/bryanmorgan/time-tracking-api/task"
)

type Project struct {
	client.Client
	ProjectId     int                `json:"-" db:"project_id"`
	ProjectName   string             `json:"-" db:"project_name"`
	Code          sql.NullString     `json:"-"`
	ProjectActive bool               `json:"-" db:"project_active"`
	Tasks         []task.ProjectTask `json:"-"`
}

type ProjectTaskEntry struct {
	ProjectId int `json:"-" db:"project_id"`
	TaskId    int `json:"-" db:"task_id"`
}
