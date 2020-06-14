package task

import (
	"database/sql"
	"errors"
	"strconv"

	"github.com/bryanmorgan/time-tracking-api/database"
	"github.com/jmoiron/sqlx"
)

// Compile Only: ensure interface is implemented
var _ TaskStore = &TaskData{}

type TaskStore interface {
	GetTask(taskId int, accountId int) (*Task, error)
	GetAllTasks(accountId int, active bool) ([]*Task, error)

	SaveTask(*Task) (int, error)
	UpdateTask(*Task) error

	ArchiveTask(taskId int, accountId int) error
	RestoreTask(taskId int, accountId int) error
	DeleteTask(taskId int, accountId int) error
}

type TaskData struct {
	db *sqlx.DB
}

func NewTaskStore(db *sqlx.DB) TaskStore {
	return &TaskData{
		db: db,
	}
}

func (c *TaskData) GetTask(taskId int, accountId int) (*Task, error) {
	sqlStatement := `
		SELECT *
		FROM task
 		WHERE task_id=$1 and account_id=$2`

	task := Task{}
	err := c.db.Get(&task, sqlStatement, taskId, accountId)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &task, nil
}

func (c *TaskData) GetAllTasks(accountId int, active bool) ([]*Task, error) {
	sqlStatement := `
		SELECT *
		FROM task
 		WHERE account_id=$1
          AND task_active=$2
        ORDER BY LOWER(task_name)`

	rows, err := c.db.Queryx(sqlStatement, accountId, active)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}
	defer database.CloseRows(rows)

	var tasks []*Task
	for rows.Next() {
		var t Task
		err := rows.StructScan(&t)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, &t)
	}

	return tasks, nil
}

func (c *TaskData) SaveTask(task *Task) (int, error) {
	if task.AccountId <= 0 {
		return 0, errors.New("invalid account id")
	}

	var taskId int
	sqlStatement := `
		INSERT INTO task (account_id, task_name, common, default_rate, default_billable)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING task_id`

	err := c.db.QueryRow(sqlStatement, task.AccountId, task.Name, task.Common, task.DefaultRate, task.DefaultBillable).Scan(&taskId)
	if err != nil {
		return 0, err
	}

	return taskId, nil
}

func (c *TaskData) UpdateTask(task *Task) error {
	if task.AccountId <= 0 {
		return errors.New("invalid account id")
	}

	if task.TaskId <= 0 {
		return errors.New("invalid task id")
	}

	sqlStatement := `
		UPDATE task SET task_name=$1, default_billable=$2, default_rate=$3, common=$4, task_active=$5
		WHERE account_id=$6
		  AND task_id=$7`

	results, err := c.db.Exec(sqlStatement, task.Name, task.DefaultBillable, task.DefaultRate, task.Common, task.TaskActive, task.AccountId, task.TaskId)
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

func (c *TaskData) ArchiveTask(taskId int, accountId int) error {
	sqlStatement := `UPDATE task SET task_active=false WHERE task_id=$1 and account_id=$2`

	result, err := c.db.Exec(sqlStatement, taskId, accountId)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return database.NoRowAffectedError
	}

	return nil
}

func (c *TaskData) RestoreTask(taskId int, accountId int) error {
	sqlStatement := `UPDATE task SET task_active=true WHERE task_id=$1 and account_id=$2`

	result, err := c.db.Exec(sqlStatement, taskId, accountId)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return database.NoRowAffectedError
	}

	return nil
}

func (c *TaskData) DeleteTask(taskId int, accountId int) error {
	if taskId <= 0 || accountId <= 0 {
		return errors.New("invalid task id: " + strconv.Itoa(taskId) + " or account id: " + strconv.Itoa(accountId))
	}

	sqlStatement := `DELETE FROM task WHERE task_id=$1 AND account_id=$2`

	result, err := c.db.Exec(sqlStatement, taskId, accountId)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return database.NoRowAffectedError
	}

	return nil
}
