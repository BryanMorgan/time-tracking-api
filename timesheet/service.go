package timesheet

import (
	"time"

	"github.com/bryanmorgan/time-tracking-api/api"
	"github.com/bryanmorgan/time-tracking-api/database"
)

// Compile Only: ensure interface is implemented
var _ TimeService = &TimeResource{}

type TimeService interface {
	GetTimeEntriesForRange(profileId int, accountId int, start time.Time, end time.Time) ([]*TimeEntry, *api.Error)

	SaveOrUpdateTimeEntries(entries []*TimeEntry) *api.Error
	UpdateTimeEntries(entries []*TimeEntry) *api.Error
	AddInitialProjectTimeEntries(profileId int, accountId int, start time.Time, end time.Time, projectId int, taskId int) *api.Error

	DeleteProjectForDates(profileId int, accountId int, projectId int, taskId int, start time.Time, end time.Time) *api.Error
}

type TimeResource struct {
	store TimeStore
}

func NewTimeService(store TimeStore) TimeService {
	return &TimeResource{store: store}
}

func (c *TimeResource) GetTimeEntriesForRange(profileId int, accountId int, start time.Time, end time.Time) ([]*TimeEntry, *api.Error) {
	timeEntries, err := c.store.GetTimeEntriesForRange(profileId, accountId, start, end)
	if err != nil {
		return nil, api.NewError(err, "Could not get time entries", api.SystemError)
	}

	return timeEntries, nil
}

func (c *TimeResource) SaveOrUpdateTimeEntries(entries []*TimeEntry) *api.Error {
	err := c.store.SaveOrUpdateTimeEntries(entries)
	if err != nil {
		return api.NewError(err, "Failed to save or update time entries", api.SystemError)
	}

	return nil
}

func (c *TimeResource) UpdateTimeEntries(entries []*TimeEntry) *api.Error {
	err := c.store.UpdateTimeEntries(entries)
	if err != nil {
		return api.NewError(err, "Failed to update time entries", api.SystemError)
	}

	return nil
}

func (c *TimeResource) AddInitialProjectTimeEntries(profileId int, accountId int, start time.Time, end time.Time, projectId int, taskId int) *api.Error {
	err := c.store.AddInitialProjectTimeEntries(profileId, accountId, start, end, projectId, taskId)
	if err != nil {
		return api.NewError(err, "Failed to add initial project time entries", api.SystemError)
	}

	return nil
}

func (c *TimeResource) DeleteProjectForDates(profileId int, accountId int, projectId int, taskId int, start time.Time, end time.Time) *api.Error {
	err := c.store.DeleteProjectForDates(profileId, accountId, projectId, taskId, start, end)
	if err == database.NoRowAffectedError {
		return api.NewError(err, "No matching project/task time entries found", api.InvalidField)
	}

	if err != nil {
		return api.NewError(err, "Failed to delete project from time entries", api.SystemError)
	}

	return nil
}
