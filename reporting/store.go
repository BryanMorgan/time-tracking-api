package reporting

import (
	"database/sql"
	"time"

	"github.com/bryanmorgan/time-tracking-api/config"
	"github.com/bryanmorgan/time-tracking-api/database"
	"github.com/jmoiron/sqlx"
)

// Compile Only: ensure interface is implemented
var _ ReportingStore = &ReportingData{}

const ReportPaginationLimit = 100

type ReportingStore interface {
	GetTimeByClient(accountId int, fromDate time.Time, toDate time.Time, offset int) ([]*ClientReport, error)
	GetTimeByProject(accountId int, fromDate time.Time, toDate time.Time, offset int) ([]*ProjectReport, error)
	GetTimeByTask(accountId int, fromDate time.Time, toDate time.Time, offset int) ([]*TaskReport, error)
	GetTimeByPerson(accountId int, fromDate time.Time, toDate time.Time, offset int) ([]*PersonReport, error)
}

type ReportingData struct {
	db *sqlx.DB
}

func NewReportingStore(db *sqlx.DB) ReportingStore {
	return &ReportingData{
		db: db,
	}
}

func (c *ReportingData) GetTimeByClient(accountId int, fromDate time.Time, toDate time.Time, offset int) ([]*ClientReport, error) {
	sqlStatement := `
		SELECT c.client_name,
		       c.client_id,
       		   sum(t.hours) filter (where not pt.billable)       as non_billable_hours,
       		   sum(t.hours) filter (where pt.billable)           as billable_hours,
       		   sum(t.hours * pt.rate) filter (where pt.billable) as billable_total
		FROM time t,
     		 project_task pt,
       		 project p,
     	  	 client c
		WHERE t.account_id = $1
  		  AND t.account_id = pt.account_id
  		  AND t.project_id = pt.project_id
  		  AND t.task_id = pt.task_id
  		  AND pt.project_id = p.project_id
  		  AND c.client_id = p.client_id
		  AND t.hours > 0.0
		  AND day >= $2
		  AND day <= $3
		GROUP BY c.client_id
		ORDER BY c.client_name
		LIMIT $4
		OFFSET $5		
`
	rows, err := c.db.Queryx(sqlStatement,
		accountId,
		fromDate.Format(config.ISOShortDateFormat),
		toDate.Format(config.ISOShortDateFormat),
		ReportPaginationLimit,
		offset*ReportPaginationLimit)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	defer database.CloseRows(rows)

	var clientRows []*ClientReport
	for rows.Next() {
		var r ClientReport
		err := rows.StructScan(&r)
		if err != nil {
			return nil, err
		}
		clientRows = append(clientRows, &r)
	}

	return clientRows, nil
}

func (c *ReportingData) GetTimeByProject(accountId int, fromDate time.Time, toDate time.Time, offset int) ([]*ProjectReport, error) {
	sqlStatement := `
	  SELECT p.project_id, 
	         p.project_name, 
	         c.client_name, 
	         bt.non_billable_hours, 
	         bt.billable_hours, 
	         bt.billable_total
      FROM (SELECT t.project_id,
                   sum(t.hours) FILTER (WHERE NOT pt.billable)       AS non_billable_hours,
                   sum(t.hours) FILTER (WHERE pt.billable)           AS billable_hours,
                   sum(t.hours * pt.rate) FILTER (WHERE pt.billable) AS billable_total
            FROM time t,
                 project_task pt
            WHERE t.account_id = $1
              AND t.account_id = pt.account_id
              AND t.project_id = pt.project_id
              AND t.task_id = pt.task_id
              AND t.hours > 0.0
              AND day >= $2
              AND day <= $3
            GROUP BY t.project_id
            ORDER BY t.project_id
            LIMIT $4
            OFFSET $5) bt,
           project p,
           client c
      WHERE bt.project_id = p.project_id
        AND p.client_id = c.client_id
      ORDER BY p.project_name`

	rows, err := c.db.Queryx(sqlStatement,
		accountId,
		fromDate.Format(config.ISOShortDateFormat),
		toDate.Format(config.ISOShortDateFormat),
		ReportPaginationLimit,
		offset*ReportPaginationLimit)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	defer database.CloseRows(rows)

	var projectRows []*ProjectReport
	for rows.Next() {
		var p ProjectReport
		err := rows.StructScan(&p)
		if err != nil {
			return nil, err
		}
		projectRows = append(projectRows, &p)
	}

	return projectRows, nil
}

func (c *ReportingData) GetTimeByTask(accountId int, fromDate time.Time, toDate time.Time, offset int) ([]*TaskReport, error) {
	sqlStatement := `
	  SELECT t.task_id,
	         t.task_name, 
	         bt.non_billable_hours, 
	         bt.billable_hours, 
	         bt.billable_total
      FROM (SELECT t.task_id,
                   sum(t.hours) FILTER (WHERE NOT pt.billable)       AS non_billable_hours,
                   sum(t.hours) FILTER (WHERE pt.billable)           AS billable_hours,
                   sum(t.hours * pt.rate) FILTER (WHERE pt.billable) AS billable_total
            FROM time t,
                 project_task pt
            WHERE t.account_id = $1
              AND t.account_id = pt.account_id
              AND t.project_id = pt.project_id
              AND t.task_id = pt.task_id
              AND t.hours > 0.0
              AND day >= $2
              AND day <= $3
            GROUP BY t.task_id
            ORDER BY t.task_id
            LIMIT $4
            OFFSET $5) bt,
           task t
      WHERE bt.task_id = t.task_id
      ORDER BY t.task_name;		
`
	rows, err := c.db.Queryx(sqlStatement,
		accountId,
		fromDate.Format(config.ISOShortDateFormat),
		toDate.Format(config.ISOShortDateFormat),
		ReportPaginationLimit,
		offset*ReportPaginationLimit)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	defer database.CloseRows(rows)

	var taskRows []*TaskReport
	for rows.Next() {
		var t TaskReport
		err := rows.StructScan(&t)
		if err != nil {
			return nil, err
		}
		taskRows = append(taskRows, &t)
	}

	return taskRows, nil
}

func (c *ReportingData) GetTimeByPerson(accountId int, fromDate time.Time, toDate time.Time, offset int) ([]*PersonReport, error) {
	sqlStatement := `
		SELECT p.profile_id,
		       p.first_name,
		       p.last_name,
       		   sum(t.hours) filter (where not pt.billable)       as non_billable_hours,
       		   sum(t.hours) filter (where pt.billable)           as billable_hours,
       		   sum(t.hours * pt.rate) filter (where pt.billable) as billable_total
		FROM time t,
     		 project_task pt,
       		 profile p
		WHERE t.account_id = $1
  		  AND t.account_id = pt.account_id
  		  AND t.project_id = pt.project_id
  		  AND t.task_id = pt.task_id
  		  AND t.profile_id = p.profile_id
		  AND t.hours > 0.0
		  AND day >= $2
		  AND day <= $3
		GROUP BY p.profile_id
		ORDER BY p.last_name
		LIMIT $4
		OFFSET $5		
`
	rows, err := c.db.Queryx(sqlStatement,
		accountId,
		fromDate.Format(config.ISOShortDateFormat),
		toDate.Format(config.ISOShortDateFormat),
		ReportPaginationLimit,
		offset*ReportPaginationLimit)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	defer database.CloseRows(rows)

	var personRows []*PersonReport
	for rows.Next() {
		var p PersonReport
		err := rows.StructScan(&p)
		if err != nil {
			return nil, err
		}
		personRows = append(personRows, &p)
	}

	return personRows, nil
}
