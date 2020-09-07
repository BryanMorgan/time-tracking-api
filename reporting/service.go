package reporting

import (
	"time"

	"github.com/bryanmorgan/time-tracking-api/api"
)

// Compile Only: ensure interface is implemented
var _ ReportingService = &ReportingResource{}

type ReportingService interface {
	GetTimeByClient(accountId int, fromDate time.Time, toDate time.Time, offset int) ([]*ClientReport, *api.Error)
	GetTimeByProject(accountId int, fromDate time.Time, toDate time.Time, offset int) ([]*ProjectReport, *api.Error)
	GetTimeByTask(accountId int, fromDate time.Time, toDate time.Time, offset int) ([]*TaskReport, *api.Error)
	GetTimeByPerson(accountId int, fromDate time.Time, toDate time.Time, offset int) ([]*PersonReport, *api.Error)
}

type ReportingResource struct {
	store ReportingStore
}

func NewReportingService(store ReportingStore) ReportingService {
	return &ReportingResource{store: store}
}

func (c *ReportingResource) GetTimeByClient(accountId int, fromDate time.Time, toDate time.Time, offset int) ([]*ClientReport, *api.Error) {
	clientReportRows, err := c.store.GetTimeByClient(accountId, fromDate, toDate, offset)
	if err != nil {
		return nil, api.NewError(err, "Failed to get time by client", api.SystemError)
	}

	return clientReportRows, nil
}

func (c *ReportingResource) GetTimeByProject(accountId int, fromDate time.Time, toDate time.Time, offset int) ([]*ProjectReport, *api.Error) {
	projectReportRows, err := c.store.GetTimeByProject(accountId, fromDate, toDate, offset)
	if err != nil {
		return nil, api.NewError(err, "Failed to get time by project", api.SystemError)
	}

	return projectReportRows, nil
}

func (c *ReportingResource) GetTimeByTask(accountId int, fromDate time.Time, toDate time.Time, offset int) ([]*TaskReport, *api.Error) {
	taskReportRows, err := c.store.GetTimeByTask(accountId, fromDate, toDate, offset)
	if err != nil {
		return nil, api.NewError(err, "Failed to get time by project", api.SystemError)
	}

	return taskReportRows, nil
}

func (c *ReportingResource) GetTimeByPerson(accountId int, fromDate time.Time, toDate time.Time, offset int) ([]*PersonReport, *api.Error) {
	personReportRows, err := c.store.GetTimeByPerson(accountId, fromDate, toDate, offset)
	if err != nil {
		return nil, api.NewError(err, "Failed to get time by person", api.SystemError)
	}

	return personReportRows, nil
}
