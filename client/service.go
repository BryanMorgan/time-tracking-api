package client

import (
	"time"

	"github.com/bryanmorgan/time-tracking-api/api"
	"github.com/bryanmorgan/time-tracking-api/database"
	"github.com/bryanmorgan/time-tracking-api/timesheet"
	"github.com/bryanmorgan/time-tracking-api/valid"
)

// Compile Only: ensure interface is implemented
var _ ClientService = &ClientResource{}

type ClientService interface {
	GetClient(clientId int, accountId int) (*Client, *api.Error)
	GetAllClients(accountId int, active bool) ([]*Client, *api.Error)
	GetProject(projectId int, accountId int) (*Project, *api.Error)
	GetAllProjects(accountId int, active bool) ([]*Project, *api.Error)

	CreateClient(accountId int, name string, address string) (*Client, *api.Error)
	CreateProject(newProject *Project) (*Project, *api.Error)
	UpdateClient(updateClient *Client) *api.Error
	UpdateProject(updateProject *Project) *api.Error

	ArchiveClient(clientId int, accountId int) *api.Error
	RestoreClient(clientId int, accountId int) *api.Error
	UpdateProjectActive(projectId int, accountId int, active bool) *api.Error

	DeleteClient(clientId int, accountId int) *api.Error
	DeleteProject(projectId int, accountId int) *api.Error

	CopyProjectsFromDateRanges(profileId int, accountId int, fromStart time.Time, fromEnd time.Time, toStart time.Time, toEnd time.Time) ([]*timesheet.TimeEntry, *api.Error)
}

type ClientResource struct {
	store     ClientStore
	timeStore timesheet.TimeStore
}

func NewClientService(store ClientStore, timeStore timesheet.TimeStore) ClientService {
	return &ClientResource{
		store:     store,
		timeStore: timeStore,
	}
}

func (c *ClientResource) GetClient(clientId int, accountId int) (*Client, *api.Error) {
	clientData, err := c.store.GetClient(clientId, accountId)
	if err != nil {
		return nil, api.NewError(err, "Could not get client data", api.SystemError)
	}

	return clientData, nil
}

func (c *ClientResource) GetAllClients(accountId int, active bool) ([]*Client, *api.Error) {
	clients, err := c.store.GetAllClients(accountId, active)
	if err != nil {
		return nil, api.NewError(err, "Could not get all clients", api.SystemError)
	}

	return clients, nil
}

func (c *ClientResource) CreateClient(accountId int, name string, address string) (*Client, *api.Error) {
	newClient := Client{
		AccountId:  accountId,
		ClientName: name,
		Address:    valid.ToNullString(address),
	}

	clientId, err := c.store.CreateClient(newClient)
	if err != nil {
		return nil, api.NewError(err, "Could not create client", api.SystemError)
	}

	newClient.ClientId = clientId
	return &newClient, nil
}

func (c *ClientResource) UpdateClient(updateClient *Client) *api.Error {
	err := c.store.UpdateClient(updateClient)
	if err != nil {
		return api.NewError(err, "Could not update client", api.SystemError)
	}

	return nil
}

func (c *ClientResource) ArchiveClient(clientId int, accountId int) *api.Error {
	err := c.store.ArchiveClient(clientId, accountId)
	if err != nil {
		return api.NewError(err, "Could not archive client", api.SystemError)
	}

	return nil
}

func (c *ClientResource) RestoreClient(clientId int, accountId int) *api.Error {
	err := c.store.RestoreClient(clientId, accountId)
	if err != nil {
		return api.NewError(err, "Could not restore archived client", api.SystemError)
	}

	return nil
}

func (c *ClientResource) DeleteClient(clientId int, accountId int) *api.Error {
	err := c.store.DeleteClient(clientId, accountId)
	if err != nil {
		return api.NewError(err, "Could not delete client", api.SystemError)
	}

	return nil
}

// --- Project

func (c *ClientResource) GetProject(projectId int, accountId int) (*Project, *api.Error) {
	project, err := c.store.GetProject(projectId, accountId)
	if err != nil {
		return nil, api.NewError(err, "Could not get project data", api.SystemError)
	}

	return project, nil
}

func (c *ClientResource) GetAllProjects(accountId int, active bool) ([]*Project, *api.Error) {
	projects, err := c.store.GetAllProjects(accountId, active)
	if err != nil {
		return nil, api.NewError(err, "Could not get all projects", api.SystemError)
	}

	return projects, nil
}

func (c *ClientResource) CreateProject(newProject *Project) (*Project, *api.Error) {
	projectId, err := c.store.CreateProject(newProject)
	if err != nil {
		return nil, api.NewError(err, "Could not create project", api.SystemError)
	}

	newProject.ProjectId = projectId
	return newProject, nil
}

func (c *ClientResource) UpdateProject(updateProject *Project) *api.Error {
	existingProject, appErr := c.GetProject(updateProject.ProjectId, updateProject.Client.AccountId)
	if appErr != nil {
		return api.NewError(appErr, "Error retrieving existing project", api.InvalidProject)
	}

	if existingProject == nil {
		return api.NewError(appErr, "No project found", api.InvalidProject)
	}

	err := c.store.UpdateProject(updateProject)
	if err != nil {
		return api.NewError(err, "Could not update project", api.SystemError)
	}

	return nil
}

func (c *ClientResource) UpdateProjectActive(projectId int, accountId int, active bool) *api.Error {
	err := c.store.UpdateProjectActive(projectId, accountId, active)
	if err == database.NoRowAffectedError {
		return api.NewError(err, "Project not found", api.InvalidProject)
	} else if err != nil {
		return api.NewError(err, "Could not archive project", api.SystemError)
	}

	return nil
}

func (c *ClientResource) DeleteProject(projectId int, accountId int) *api.Error {
	err := c.store.DeleteProject(projectId, accountId)
	if err == database.NoRowAffectedError {
		return api.NewError(err, "Project not found", api.InvalidProject)
	} else if err != nil {
		return api.NewError(err, "Could not delete project", api.SystemError)
	}

	return nil
}

func (c *ClientResource) CopyProjectsFromDateRanges(profileId int, accountId int, fromStart time.Time, fromEnd time.Time, toStart time.Time, toEnd time.Time) ([]*timesheet.TimeEntry, *api.Error) {
	var timeEntries []*timesheet.TimeEntry
	var serviceErr error
	success, err := c.store.CopyProjectsFromDateRanges(profileId, accountId, fromStart, fromEnd, toStart, toEnd)
	if err != nil {
		return nil, api.NewError(err, "Error copying projects from prior date range", api.SystemError)
	}

	if success {
		// Get all time/projects/tasks for the date range
		timeEntries, serviceErr = c.timeStore.GetTimeEntriesForRange(profileId, accountId, toStart, toEnd)
		if serviceErr != nil {
			return nil, api.NewError(serviceErr, "Failed to get time entries for 'to' date range", api.SystemError)

		}
	}

	return timeEntries, nil
}
