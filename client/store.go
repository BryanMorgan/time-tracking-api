package client

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/bryanmorgan/time-tracking-api/config"
	"github.com/bryanmorgan/time-tracking-api/database"
	"github.com/bryanmorgan/time-tracking-api/logger"
	"github.com/bryanmorgan/time-tracking-api/task"
	"github.com/bryanmorgan/time-tracking-api/valid"

	"github.com/jmoiron/sqlx"
)

// Compile Only: ensure interface is implemented
var _ ClientStore = &ClientData{}

type ClientStore interface {
	GetClient(clientId int, accountId int) (*Client, error)
	GetAllClients(accountId int, active bool) ([]*Client, error)
	GetProject(projectId int, accountId int) (*Project, error)
	GetAllProjects(accountId int, active bool) ([]*Project, error)

	CreateClient(client Client) (int, error)
	CreateProject(project *Project) (int, error)
	UpdateClient(client *Client) error
	UpdateProject(project *Project) error

	ArchiveClient(clientId int, accountId int) error
	RestoreClient(clientId int, accountId int) error
	UpdateProjectActive(projectId int, accountId int, active bool) error

	DeleteClient(clientId int, accountId int) error
	DeleteProject(projectId int, accountId int) error

	CopyProjectsFromDateRanges(profileId int, accountId int, fromStart time.Time, fromEnd time.Time, toStart time.Time, toEnd time.Time) (bool, error)
}

// ProfileData implements database operations for user profiles
type ClientData struct {
	db *sqlx.DB
}

func NewClientStore(db *sqlx.DB) ClientStore {
	return &ClientData{
		db: db,
	}
}

func (c *ClientData) GetClient(clientId int, accountId int) (*Client, error) {
	sqlStatement := `
		SELECT client_id, account_id, client_name, address, client_active
		FROM client
 		WHERE client_id=$1 and account_id=$2`

	clientData := Client{}
	err := c.db.Get(&clientData, sqlStatement, clientId, accountId)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &clientData, nil
}

func (c *ClientData) GetAllClients(accountId int, active bool) ([]*Client, error) {
	sqlStatement := `SELECT client_id, account_id, client_name, address, client_active
		FROM client
 		WHERE account_id=$1
          AND client_active=$2`

	rows, err := c.db.Queryx(sqlStatement, accountId, active)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}
	defer database.CloseRows(rows)

	var clients []*Client
	for rows.Next() {
		var c Client
		err := rows.StructScan(&c)
		if err != nil {
			return nil, err
		}
		clients = append(clients, &c)
	}

	return clients, nil
}

func (c *ClientData) CreateClient(newClient Client) (int, error) {
	if valid.IsNull(newClient.ClientName) || newClient.AccountId <= 0 {
		return 0, errors.New("invalid client name or account")
	}

	sqlStatement := `
		INSERT INTO client (account_id, client_name, address, client_active)
		VALUES ($1, $2, $3, TRUE)
		RETURNING client_id`

	var clientId int
	err := c.db.QueryRow(sqlStatement, newClient.AccountId, newClient.ClientName, newClient.Address).Scan(&clientId)
	if err != nil {
		return 0, err
	}

	return clientId, nil
}

func (c *ClientData) UpdateClient(updateData *Client) error {
	if updateData.ClientId <= 0 {
		return errors.New("invalid client id: " + strconv.Itoa(updateData.ClientId))
	}

	// Make sure this client and account are valid before adding the project
	selectStatement := `SELECT count(*) FROM client WHERE client_id=$1 and account_id=$2`
	var count int
	err := c.db.Get(&count, selectStatement, updateData.ClientId, updateData.AccountId)
	if err != nil {
		return err
	}

	if count == 0 {
		return errors.New("invalid client id for account")
	}

	sqlStatement := `
	UPDATE client SET client_name=$1, address=$2, client_active=$3 WHERE client_id=$4`

	result, err := c.db.Exec(sqlStatement, updateData.ClientName, updateData.Address, updateData.ClientActive, updateData.ClientId)
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

func (c *ClientData) ArchiveClient(clientId int, accountId int) error {
	sqlStatement := `UPDATE client SET client_active=FALSE WHERE client_id=$1 AND account_id=$2`

	result, err := c.db.Exec(sqlStatement, clientId, accountId)
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

func (c *ClientData) RestoreClient(clientId int, accountId int) error {
	sqlStatement := `UPDATE client SET client_active=TRUE WHERE client_id=$1 AND account_id=$2`

	result, err := c.db.Exec(sqlStatement, clientId, accountId)
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

func (c *ClientData) DeleteClient(clientId int, accountId int) error {
	if clientId <= 0 || accountId <= 0 {
		return errors.New("invalid client id: " + strconv.Itoa(clientId) + " or account id: " + strconv.Itoa(accountId))
	}

	sqlStatement := `DELETE FROM client WHERE client_id=$1 AND account_id=$2`

	result, err := c.db.Exec(sqlStatement, clientId, accountId)
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

// --- Project

func (c *ClientData) GetProject(projectId int, accountId int) (*Project, error) {
	projectSql := `
	SELECT p.project_id, p.account_id, p.project_active, code, p.project_name,
		   c.client_id, c.client_name
	FROM project p,
         client c
	WHERE p.project_id=$1
      AND p.account_id = $2
	  AND p.client_id = c.client_id
      AND c.client_active = true`

	taskSql := `
	SELECT t.*, pt.*
    FROM task t,
	     project_task pt
	WHERE pt.project_id = $1
	  AND pt.account_id = $2
	  AND pt.account_id = t.account_id
      AND pt.task_id = t.task_id
	  AND t.task_active = true
	  ORDER BY LOWER(t.task_name)`

	project := Project{}

	// First get the project info
	err := c.db.Get(&project, projectSql, projectId, accountId)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	project.Tasks = []task.ProjectTask{}

	// Next get the project's tasks
	err = c.db.Select(&project.Tasks, taskSql, projectId, accountId)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &project, nil
}

func (c *ClientData) GetAllProjects(accountId int, active bool) ([]*Project, error) {
	projectSql := `
	SELECT p.project_id, p.account_id, p.project_active, code, p.project_name,
		   c.client_id, c.client_name
	FROM project p, client c
	WHERE p.account_id=$1
	  AND p.account_id = c.account_id
	  AND p.client_id = c.client_id
      AND c.client_active = true
      AND p.project_active = $2
	ORDER BY LOWER(client_name), 
			 LOWER(project_name);`

	taskSql := `
	SELECT t.*, pt.*
    FROM task t,
	     project_task pt
	WHERE pt.account_id = $1
	  AND pt.account_id = t.account_id
      AND pt.task_id = t.task_id
	  AND t.task_active = true
      ORDER BY LOWER(task_name)`

	rows, err := c.db.Queryx(projectSql, accountId, active)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}
	defer database.CloseRows(rows)

	var allTasks []task.ProjectTask

	// Next get the project's tasks
	err = c.db.Select(&allTasks, taskSql, accountId)
	if err != nil {
		return nil, err
	}

	var projects []*Project
	for rows.Next() {
		var p Project
		err := rows.StructScan(&p)

		if err != nil {
			return nil, err
		}
		p.Tasks = []task.ProjectTask{}
		for _, currentTask := range allTasks {

			if currentTask.ProjectId == p.ProjectId {
				p.Tasks = append(p.Tasks, currentTask)
			}
		}

		projects = append(projects, &p)
	}

	return projects, nil
}

func (c *ClientData) CreateProject(newProject *Project) (int, error) {
	if valid.IsNull(newProject.ProjectName) || newProject.Client.AccountId <= 0 || newProject.Client.ClientId <= 0 {
		logger.Log.Error("Invalid project: " + fmt.Sprintf("%+v", newProject))
		return 0, errors.New("invalid project name, account id, or client id")
	}

	// Make sure this client and account are valid before adding the project
	selectStatement := `SELECT count(*) FROM client WHERE client_id=$1 and account_id=$2`
	var count int
	err := c.db.Get(&count, selectStatement, newProject.Client.ClientId, newProject.Client.AccountId)
	if err != nil {
		return 0, err
	}

	if count == 0 {
		return 0, errors.New("invalid client id for account")
	}

	projectSql := `
		INSERT INTO project (account_id, client_id, project_name, code, project_active)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING project_id`

	var projectId int
	err = c.db.QueryRow(projectSql, newProject.Client.AccountId, newProject.Client.ClientId, newProject.ProjectName, newProject.Code, newProject.ProjectActive).Scan(&projectId)
	if err != nil {
		return 0, err
	}

	taskSql := `INSERT INTO project_task (project_id, task_id, account_id, rate, billable, project_active) VALUES ($1, $2, $3, $4, $5, $6)`
	for _, taskData := range newProject.Tasks {
		result, err := c.db.Exec(taskSql,
			projectId,
			taskData.TaskId,
			newProject.Client.AccountId,
			taskData.Rate,
			taskData.Billable,
			taskData.ProjectActive)

		if err != nil {
			return 0, err
		}

		rows, err := result.RowsAffected()
		if err != nil {
			return 0, err
		}

		if rows == 0 {
			return 0, database.NoRowAffectedError
		}
	}

	return projectId, nil
}

func (c *ClientData) UpdateProject(updateProject *Project) error {
	if updateProject.ProjectId <= 0 {
		return errors.New("invalid project id: " + strconv.Itoa(updateProject.ProjectId))
	}

	// Make sure this project and account are valid before updating the project
	selectStatement := `SELECT count(*) FROM project WHERE project_id=$1 and account_id=$2`
	var count int
	err := c.db.Get(&count, selectStatement, updateProject.ProjectId, updateProject.Client.AccountId)
	if err != nil {
		return err
	}

	if count == 0 {
		return errors.New("invalid project id for account")
	}

	sqlStatement := `UPDATE project SET project_name=$1, client_id=$2, project_active=$3 WHERE project_id=$4`

	result, err := c.db.Exec(sqlStatement, updateProject.ProjectName, updateProject.ClientId, updateProject.ProjectActive, updateProject.ProjectId)
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

	// Delete all tasks from this project
	deleteSql := `
	DELETE FROM project_task
	WHERE project_id=$1
      AND account_id = $2`
	result, err = c.db.Exec(deleteSql, updateProject.ProjectId, updateProject.AccountId)
	if err != nil {
		return err
	}

	// Add all tasks for this project
	taskSql := `INSERT INTO project_task 
			    (project_id, account_id, task_id, rate, billable, project_active) 
				VALUES ($1, $2, $3, $4, $5, $6)`
	for _, taskData := range updateProject.Tasks {
		result, err := c.db.Exec(taskSql,
			updateProject.ProjectId,
			updateProject.AccountId,
			taskData.TaskId,
			taskData.Rate,
			taskData.Billable,
			taskData.ProjectActive,
		)
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
	}

	return nil
}

func (c *ClientData) UpdateProjectActive(projectId int, accountId int, active bool) error {
	sqlStatement := `UPDATE project SET project_active=$3 WHERE project_id=$1 AND account_id=$2`

	result, err := c.db.Exec(sqlStatement, projectId, accountId, active)
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

func (c *ClientData) DeleteProject(projectId int, accountId int) error {
	if projectId <= 0 || accountId <= 0 {
		return errors.New("invalid project id: " + strconv.Itoa(projectId) + " or account id: " + strconv.Itoa(accountId))
	}

	sqlStatement := `DELETE FROM project WHERE project_id=$1 AND account_id=$2`

	result, err := c.db.Exec(sqlStatement, projectId, accountId)
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

// Copy projects and tasks for all the days between prior start and end dates into the new date range
func (c *ClientData) CopyProjectsFromDateRanges(profileId int, accountId int, fromStart time.Time, fromEnd time.Time, toStart time.Time, toEnd time.Time) (bool, error) {
	// Get all the project/task entries from the "From" start/end date range
	sqlStatement := `
		SELECT DISTINCT project_id, task_id
		FROM time
 		WHERE profile_id = $1
          AND account_id = $2
          AND day >= $3
          AND day <= $4`

	rows, err := c.db.Queryx(sqlStatement, profileId, accountId, fromStart.Format(config.ISOShortDateFormat), fromEnd.Format(config.ISOShortDateFormat))
	if err == sql.ErrNoRows {
		return false, nil
	}

	if err != nil {
		return false, err
	}
	defer database.CloseRows(rows)

	var projectTaskEntries []*ProjectTaskEntry
	for rows.Next() {
		var pt ProjectTaskEntry
		err := rows.StructScan(&pt)
		if err != nil {
			return false, err
		}
		projectTaskEntries = append(projectTaskEntries, &pt)
	}

	// If there weren't any entries for the prior date range, return false
	if len(projectTaskEntries) == 0 {
		return false, nil
	}

	// Loop through all the "To" dates and add the project/task entries we collected
	tx, err := c.db.Begin()
	if err != nil {
		return false, err
	}

	for day := toStart; day.Before(toEnd) || day.Equal(toEnd); day = day.AddDate(0, 0, 1) {
		for _, entry := range projectTaskEntries {
			insertSql := `INSERT INTO time (account_id, profile_id, project_id, task_id, day, hours)
 				  VALUES ($1, $2, $3, $4, $5, 0.0)`

			results, err := c.db.Exec(insertSql, accountId, profileId, entry.ProjectId, entry.TaskId, day.Format(config.ISOShortDateFormat))
			if err != nil {
				database.RollbackTransaction(tx)
				return false, err
			}

			rows, err := results.RowsAffected()
			if err != nil {
				database.RollbackTransaction(tx)
				return false, err
			}

			if rows == 0 {
				database.RollbackTransaction(tx)
				return false, database.NoRowAffectedError
			}
		}
	}

	if err := tx.Commit(); err != nil {
		logger.Log.Error("Error in commit transaction: " + err.Error())
		return false, err
	}

	// Found prior entries and successfully inserted them
	return true, nil
}
