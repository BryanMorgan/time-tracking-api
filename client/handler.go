package client

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/bryanmorgan/time-tracking-api/api"
	"github.com/bryanmorgan/time-tracking-api/config"
	"github.com/bryanmorgan/time-tracking-api/profile"
	"github.com/bryanmorgan/time-tracking-api/task"
	"github.com/bryanmorgan/time-tracking-api/timesheet"
	"github.com/bryanmorgan/time-tracking-api/valid"

	"github.com/go-chi/chi"
)

type ClientRequest struct {
	Id      int
	Name    string
	Address string
}

type ClientResponse struct {
	ClientId int    `json:"id,omitempty"`
	Name     string `json:"name"`
	Address  string `json:"address,omitempty"`
}

type ProjectIdRequest struct {
	ProjectId int
}

type ProjectRequest struct {
	Id       int
	ClientId int
	Name     string
}

type ProjectContainerRequest struct {
	Id       int
	ClientId int
	Name     string
	Tasks    []TaskRequest
}

type TaskRequest struct {
	Id       int
	Billable bool
	Name     string
	Rate     float64
}

type ProjectTaskResponse struct {
	TaskId   int     `json:"id"`
	Name     string  `json:"name"`
	Rate     float64 `json:"rate,omitempty"`
	Billable bool    `json:"billable"`
	Active   bool    `json:"active"`
}

type ProjectResponse struct {
	ProjectId     int                   `json:"id,omitempty"`
	ProjectName   string                `json:"name"`
	ProjectActive bool                  `json:"active"`
	ClientId      int                   `json:"clientId,omitempty"`
	Code          string                `json:"code,omitempty"`
	ClientName    string                `json:"clientName,omitempty"`
	Tasks         []ProjectTaskResponse `json:"tasks,omitempty"`
}

type StartAndEndDateRequest struct {
	StartDate string
	EndDate   string
}

func (a *ClientRouter) getClientHandler(w http.ResponseWriter, r *http.Request) {
	clientIdString := chi.URLParam(r, "clientId")
	if valid.IsNull(clientIdString) {
		api.ErrorJson(w, api.NewFieldError(nil, "No clientId query parameter found", api.InvalidField, "projectId"), http.StatusBadRequest)
		return
	}

	clientId, err := strconv.Atoi(clientIdString)
	if err != nil || clientId <= 0 {
		api.ErrorJson(w, api.NewFieldError(err, "Client id not valid", api.InvalidField, "clientId"), http.StatusBadRequest)
		return
	}

	userProfile, ok := r.Context().Value(config.ProfileContextKey).(*profile.Profile)
	if !ok || userProfile == nil {
		api.ErrorJson(w, api.NewError(nil, "Invalid profile context", api.SystemError), http.StatusInternalServerError)
		return
	}

	clientData, serviceErr := a.clientService.GetClient(clientId, userProfile.AccountId)
	if serviceErr != nil {
		api.ErrorJson(w, serviceErr, http.StatusInternalServerError)
		return
	}

	if clientData == nil {
		api.ErrorJson(w, api.NewError(nil, "No client matching id", api.InvalidClient), http.StatusBadRequest)
		return
	}

	api.Json(w, r, NewClientResponse(clientData))
}

func (a *ClientRouter) getAllClientsHandler(w http.ResponseWriter, r *http.Request) {
	userProfile, ok := r.Context().Value(config.ProfileContextKey).(*profile.Profile)
	if !ok || userProfile == nil {
		api.ErrorJson(w, api.NewError(nil, "Invalid profile context", api.SystemError), http.StatusInternalServerError)
		return
	}

	clients, err := a.clientService.GetAllClients(userProfile.AccountId, true)
	if err != nil {
		api.ErrorJson(w, err, http.StatusInternalServerError)
		return
	}

	api.Json(w, r, NewClientsResponse(clients))
}

func (a *ClientRouter) getArchivedClientsHandler(w http.ResponseWriter, r *http.Request) {
	userProfile, ok := r.Context().Value(config.ProfileContextKey).(*profile.Profile)
	if !ok || userProfile == nil {
		api.ErrorJson(w, api.NewError(nil, "Invalid profile context", api.SystemError), http.StatusInternalServerError)
		return
	}

	clients, err := a.clientService.GetAllClients(userProfile.AccountId, false)
	if err != nil {
		api.ErrorJson(w, err, http.StatusInternalServerError)
		return
	}

	api.Json(w, r, NewClientsResponse(clients))
}

func (a *ClientRouter) createClientHandler(w http.ResponseWriter, r *http.Request) {
	clientRequest, err := getClientRequest(r)
	if err != nil {
		api.ErrorJson(w, err, http.StatusBadRequest)
		return
	}

	if valid.IsNull(clientRequest.Name) {
		api.BadInputs(w, "Missing required client name", api.MissingField, "clientName")
		return
	}

	if !valid.IsLength(clientRequest.Name, ClientNameMinLength, ClientNameMaxLength) {
		api.BadInputs(w, "Client name must be between 1 and 64 characters", api.FieldSize, "clientName")
		return
	}

	userProfile, ok := r.Context().Value(config.ProfileContextKey).(*profile.Profile)
	if !ok || userProfile == nil {
		api.ErrorJson(w, api.NewError(nil, "Invalid profile context", api.SystemError), http.StatusInternalServerError)
		return
	}

	newClient, err := a.clientService.CreateClient(userProfile.AccountId, clientRequest.Name, clientRequest.Address)
	if err != nil {
		api.ErrorJson(w, api.NewError(err, "Failed to create client", api.SystemError), http.StatusInternalServerError)
		return

	}

	api.Json(w, r, NewClientResponse(newClient))
}

func (a *ClientRouter) updateClientHandler(w http.ResponseWriter, r *http.Request) {
	clientRequest, err := getClientRequest(r)
	if err != nil {
		api.ErrorJson(w, err, http.StatusBadRequest)
		return
	}

	if !valid.IsNull(clientRequest.Name) && !valid.IsLength(clientRequest.Name, ClientNameMinLength, ClientNameMaxLength) {
		api.ErrorJson(w, api.NewFieldError(nil, "Client name must be between 1 and 64 characters", api.FieldSize, "clientName"), http.StatusBadRequest)
		return
	}

	if clientRequest.Id <= 0 {
		api.ErrorJson(w, api.NewError(nil, "Missing client id", api.MissingField), http.StatusBadRequest)
		return
	}

	userProfile, ok := r.Context().Value(config.ProfileContextKey).(*profile.Profile)
	if !ok || userProfile == nil {
		api.ErrorJson(w, api.NewError(nil, "Invalid profile context", api.SystemError), http.StatusInternalServerError)
		return
	}

	existingClient, err := a.clientService.GetClient(clientRequest.Id, userProfile.AccountId)
	if err != nil || existingClient == nil {
		api.ErrorJson(w, api.NewError(err, "Failed to get existing client", api.SystemError), http.StatusInternalServerError)
		return
	}

	// Clone the client
	updateClient := *existingClient
	updateClient.ClientId = clientRequest.Id

	if !valid.IsNull(clientRequest.Name) {
		updateClient.ClientName = clientRequest.Name
	}

	if !valid.IsNull(clientRequest.Address) {
		updateClient.Address = valid.ToNullString(clientRequest.Address)
	}

	err = a.clientService.UpdateClient(&updateClient)
	if err != nil {
		api.ErrorJson(w, api.NewError(err, "Failed to update client", api.SystemError), http.StatusInternalServerError)
		return
	}

	api.Json(w, r, nil)
}

func (a *ClientRouter) deleteClientHandler(w http.ResponseWriter, r *http.Request) {
	clientRequest, err := getClientRequest(r)
	if err != nil {
		api.ErrorJson(w, err, http.StatusBadRequest)
		return
	}

	if clientRequest.Id <= 0 {
		api.ErrorJson(w, api.NewError(nil, "Missing client id", api.MissingField), http.StatusBadRequest)
		return
	}

	userProfile, ok := r.Context().Value(config.ProfileContextKey).(*profile.Profile)
	if !ok || userProfile == nil {
		api.ErrorJson(w, api.NewError(nil, "Invalid profile context", api.SystemError), http.StatusInternalServerError)
		return
	}

	err = a.clientService.DeleteClient(clientRequest.Id, userProfile.AccountId)
	if err != nil {
		api.ErrorJson(w, api.NewError(err, "Failed to delete client", api.SystemError), http.StatusInternalServerError)
		return
	}

	api.Json(w, r, nil)
}

func (a *ClientRouter) archiveClientHandler(w http.ResponseWriter, r *http.Request) {
	clientRequest, err := getClientRequest(r)
	if err != nil {
		api.ErrorJson(w, err, http.StatusBadRequest)
		return
	}

	if clientRequest.Id <= 0 {
		api.ErrorJson(w, api.NewError(nil, "Missing client id", api.MissingField), http.StatusBadRequest)
		return
	}

	userProfile, ok := r.Context().Value(config.ProfileContextKey).(*profile.Profile)
	if !ok || userProfile == nil {
		api.ErrorJson(w, api.NewError(nil, "Invalid profile context", api.SystemError), http.StatusInternalServerError)
		return
	}

	err = a.clientService.ArchiveClient(clientRequest.Id, userProfile.AccountId)
	if err != nil {
		api.ErrorJson(w, api.NewError(err, "Failed to archive client", api.SystemError), http.StatusInternalServerError)
		return
	}

	api.Json(w, r, nil)
}

func (a *ClientRouter) restoreClientHandler(w http.ResponseWriter, r *http.Request) {
	clientRequest, err := getClientRequest(r)
	if err != nil {
		api.ErrorJson(w, err, http.StatusBadRequest)
		return
	}

	if clientRequest.Id <= 0 {
		api.ErrorJson(w, api.NewError(nil, "Missing client id", api.MissingField), http.StatusBadRequest)
		return
	}

	userProfile, ok := r.Context().Value(config.ProfileContextKey).(*profile.Profile)
	if !ok || userProfile == nil {
		api.ErrorJson(w, api.NewError(nil, "Invalid profile context", api.SystemError), http.StatusInternalServerError)
		return
	}

	err = a.clientService.RestoreClient(clientRequest.Id, userProfile.AccountId)
	if err != nil {
		api.ErrorJson(w, api.NewError(err, "Failed to archive client", api.SystemError), http.StatusInternalServerError)
		return
	}

	api.Json(w, r, nil)
}

// --- Project

func (a *ClientRouter) getProjectHandler(w http.ResponseWriter, r *http.Request) {
	projectIdString := chi.URLParam(r, "projectId")

	if valid.IsNull(projectIdString) {
		api.ErrorJson(w, api.NewFieldError(nil, "No projectId query parameter found", api.InvalidField, "projectId"), http.StatusBadRequest)
		return
	}

	projectId, err := strconv.Atoi(projectIdString)
	if err != nil || projectId <= 0 {
		api.ErrorJson(w, api.NewFieldError(err, "Project id not a number", api.InvalidField, "projectId"), http.StatusBadRequest)
		return
	}

	userProfile, ok := r.Context().Value(config.ProfileContextKey).(*profile.Profile)
	if !ok || userProfile == nil {
		api.ErrorJson(w, api.NewError(nil, "Invalid profile context", api.SystemError), http.StatusInternalServerError)
		return
	}

	project, serviceErr := a.clientService.GetProject(projectId, userProfile.AccountId)
	if serviceErr != nil {
		api.ErrorJson(w, serviceErr, http.StatusInternalServerError)
		return
	}

	if project == nil {
		api.ErrorJson(w, api.NewError(nil, "No project data found", api.InvalidProject), http.StatusBadRequest)
		return
	}

	api.Json(w, r, NewProjectResponse(project))
}

func (a *ClientRouter) getAllProjectsHandler(w http.ResponseWriter, r *http.Request) {
	userProfile, ok := r.Context().Value(config.ProfileContextKey).(*profile.Profile)
	if !ok || userProfile == nil {
		api.ErrorJson(w, api.NewError(nil, "Invalid profile context", api.SystemError), http.StatusInternalServerError)
		return
	}

	projects, serviceErr := a.clientService.GetAllProjects(userProfile.AccountId, true)
	if serviceErr != nil {
		api.ErrorJson(w, serviceErr, http.StatusInternalServerError)
		return
	}

	api.Json(w, r, NewProjectsResponse(projects))
}

func (a *ClientRouter) getArchivedProjectsHandler(w http.ResponseWriter, r *http.Request) {
	userProfile, ok := r.Context().Value(config.ProfileContextKey).(*profile.Profile)
	if !ok || userProfile == nil {
		api.ErrorJson(w, api.NewError(nil, "Invalid profile context", api.SystemError), http.StatusInternalServerError)
		return
	}

	projects, serviceErr := a.clientService.GetAllProjects(userProfile.AccountId, false)
	if serviceErr != nil {
		api.ErrorJson(w, serviceErr, http.StatusInternalServerError)
		return
	}

	api.Json(w, r, NewProjectsResponse(projects))
}

func (a *ClientRouter) createProjectHandler(w http.ResponseWriter, r *http.Request) {
	projectRequest, err := getProjectRequest(r)
	if err != nil {
		api.ErrorJson(w, err, http.StatusBadRequest)
		return
	}

	if valid.IsNull(projectRequest.Name) {
		api.ErrorJson(w, api.NewFieldError(nil, "Missing required name", api.MissingField, "name"), http.StatusBadRequest)
		return
	}

	if projectRequest.ClientId <= 0 {
		api.ErrorJson(w, api.NewFieldError(nil, "Missing required clientId", api.MissingField, "clientId"), http.StatusBadRequest)
		return
	}

	if !valid.IsLength(projectRequest.Name, ProjectNameMinLength, ProjectNameMaxLength) {
		api.ErrorJson(w, api.NewFieldError(nil, "Project name must be between 1 and 128 characters", api.FieldSize, "name"), http.StatusBadRequest)
		return
	}

	userProfile, ok := r.Context().Value(config.ProfileContextKey).(*profile.Profile)
	if !ok || userProfile == nil {
		api.ErrorJson(w, api.NewError(nil, "Invalid profile context", api.SystemError), http.StatusInternalServerError)
		return
	}

	// Copy request data to Tasks and Project
	projectTasks := createProjectTaskList(projectRequest.Id, projectRequest.Tasks)
	projectData := Project{
		Client: Client{
			ClientId:  projectRequest.ClientId,
			AccountId: userProfile.AccountId,
		},
		ProjectName:   projectRequest.Name,
		ProjectActive: true,
		Tasks:         projectTasks,
	}

	newProject, err := a.clientService.CreateProject(&projectData)
	if err != nil {
		api.ErrorJson(w, api.NewError(err, "Failed to create client", api.SystemError), http.StatusInternalServerError)
		return
	}

	api.Json(w, r, NewProjectResponse(newProject))
}

func (a *ClientRouter) updateProjectHandler(w http.ResponseWriter, r *http.Request) {
	projectRequest, err := getProjectRequest(r)
	if err != nil {
		api.ErrorJson(w, err, http.StatusBadRequest)
		return
	}

	if projectRequest.Id <= 0 {
		api.ErrorJson(w, api.NewFieldError(nil, "Missing required project id", api.MissingField, "id"), http.StatusBadRequest)
		return
	}

	if projectRequest.ClientId <= 0 {
		api.ErrorJson(w, api.NewFieldError(nil, "Missing required clientId", api.MissingField, "clientId"), http.StatusBadRequest)
		return
	}

	if valid.IsNull(projectRequest.Name) || !valid.IsLength(projectRequest.Name, ProjectNameMinLength, ProjectNameMaxLength) {
		api.ErrorJson(w, api.NewFieldError(nil, "Project name must be between 1 and 128 characters", api.FieldSize, "name"), http.StatusBadRequest)
		return
	}

	userProfile, ok := r.Context().Value(config.ProfileContextKey).(*profile.Profile)
	if !ok || userProfile == nil {
		api.ErrorJson(w, api.NewError(nil, "Invalid profile context", api.SystemError), http.StatusInternalServerError)
		return
	}

	projectTasks := createProjectTaskList(projectRequest.Id, projectRequest.Tasks)
	updateProject := Project{
		Client: Client{
			AccountId: userProfile.AccountId,
			ClientId:  projectRequest.ClientId,
		},
		ProjectName:   projectRequest.Name,
		ProjectId:     projectRequest.Id,
		ProjectActive: true,
		Tasks:         projectTasks,
	}

	err = a.clientService.UpdateProject(&updateProject)
	if err != nil {
		api.ErrorJson(w, api.NewError(err, "Failed to update project", api.SystemError), http.StatusInternalServerError)
		return
	}

	api.Json(w, r, nil)

}

func (a *ClientRouter) deleteProjectHandler(w http.ResponseWriter, r *http.Request) {
	request, err := getProjectIdRequest(r)
	if err != nil {
		api.ErrorJson(w, err, http.StatusBadRequest)
		return
	}

	if request.ProjectId <= 0 {
		api.ErrorJson(w, api.NewError(nil, "Missing projectId", api.MissingField), http.StatusBadRequest)
		return
	}

	userProfile, ok := r.Context().Value(config.ProfileContextKey).(*profile.Profile)
	if !ok || userProfile == nil {
		api.ErrorJson(w, api.NewError(nil, "Invalid profile context", api.SystemError), http.StatusInternalServerError)
		return
	}

	err = a.clientService.DeleteProject(request.ProjectId, userProfile.AccountId)
	if err != nil {
		api.ErrorJson(w, api.NewError(err, "Failed to delete project", api.SystemError), http.StatusInternalServerError)
		return
	}

	api.Json(w, r, nil)
}

func (a *ClientRouter) archiveProjectHandler(w http.ResponseWriter, r *http.Request) {
	request, err := getProjectIdRequest(r)
	if err != nil {
		api.ErrorJson(w, err, http.StatusBadRequest)
		return
	}

	if request.ProjectId <= 0 {
		api.ErrorJson(w, api.NewError(nil, "Missing projectId", api.MissingField), http.StatusBadRequest)
		return
	}

	userProfile, ok := r.Context().Value(config.ProfileContextKey).(*profile.Profile)
	if !ok || userProfile == nil {
		api.ErrorJson(w, api.NewError(nil, "Invalid profile context", api.SystemError), http.StatusInternalServerError)
		return
	}

	err = a.clientService.UpdateProjectActive(request.ProjectId, userProfile.AccountId, false)
	if err != nil {
		api.ErrorJson(w, api.NewError(err, "Failed to archive project", api.SystemError), http.StatusInternalServerError)
		return
	}

	api.Json(w, r, nil)
}

func (a *ClientRouter) restoreProjectHandler(w http.ResponseWriter, r *http.Request) {
	request, err := getProjectIdRequest(r)
	if err != nil {
		api.ErrorJson(w, err, http.StatusBadRequest)
		return
	}

	if request.ProjectId <= 0 {
		api.ErrorJson(w, api.NewError(nil, "Missing projectId", api.MissingField), http.StatusBadRequest)
		return
	}

	userProfile, ok := r.Context().Value(config.ProfileContextKey).(*profile.Profile)
	if !ok || userProfile == nil {
		api.ErrorJson(w, api.NewError(nil, "Invalid profile context", api.SystemError), http.StatusInternalServerError)
		return
	}

	err = a.clientService.UpdateProjectActive(request.ProjectId, userProfile.AccountId, true)
	if err != nil {
		api.ErrorJson(w, api.NewError(err, "Failed to archive project", api.SystemError), http.StatusInternalServerError)
		return
	}

	api.Json(w, r, nil)
}

func (a *ClientRouter) copyProjectsFromLastWeek(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		api.ErrorJson(w, api.NewError(nil, "Empty Body", api.InvalidJson), http.StatusBadRequest)
		return
	}

	var dateRequest StartAndEndDateRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&dateRequest); err != nil {
		api.ErrorJson(w, api.NewError(err, "Invalid JSON", api.InvalidJson), http.StatusBadRequest)
		return
	}
	defer api.CloseBody(r.Body)

	// Ensure we got a start and end date
	if valid.IsNull(dateRequest.StartDate) || valid.IsNull(dateRequest.EndDate) {
		api.ErrorJson(w, api.NewError(nil, "Invalid start or end date", api.InvalidField), http.StatusBadRequest)
		return
	}

	// Make sure the start and end date strings are in a valid ISO 8061 format
	start, err := time.Parse(config.ISOShortDateFormat, dateRequest.StartDate)
	if err != nil {
		api.ErrorJson(w, api.NewError(err, "Invalid start date format. Use ISO8061: YYYY-MM-DD", api.InvalidField), http.StatusBadRequest)
		return
	}
	end, err := time.Parse(config.ISOShortDateFormat, dateRequest.EndDate)
	if err != nil {
		api.ErrorJson(w, api.NewError(err, "Invalid end date format. Use ISO8061: YYYY-MM-DD", api.InvalidField), http.StatusBadRequest)
		return
	}

	userProfile, ok := r.Context().Value(config.ProfileContextKey).(*profile.Profile)
	if !ok || userProfile == nil {
		api.ErrorJson(w, api.NewError(nil, "Invalid profile context", api.SystemError), http.StatusInternalServerError)
		return
	}

	// Calculate the prior week based on the start date
	priorWeekStartDate := start.AddDate(0, 0, -7)
	priorWeekEndDate := priorWeekStartDate.AddDate(0, 0, 6)

	timeEntries, serviceErr := a.clientService.CopyProjectsFromDateRanges(
		userProfile.ProfileId,
		userProfile.AccountId,
		priorWeekStartDate,
		priorWeekEndDate,
		start,
		end)

	if serviceErr != nil {
		api.ErrorJson(w, serviceErr, http.StatusInternalServerError)
		return
	}

	if len(timeEntries) == 0 {
		api.Json(w, r, nil)
	} else {
		api.Json(w, r, timesheet.NewTimeRange(timeEntries, start, end))
	}
}

func getClientRequest(r *http.Request) (*ClientRequest, *api.Error) {
	if r.Body == nil {
		return nil, api.NewError(nil, "Empty Body", api.InvalidJson)
	}

	var clientRequest ClientRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&clientRequest); err != nil {
		return nil, api.NewError(err, "Invalid JSON", api.InvalidJson)
	}
	defer api.CloseBody(r.Body)

	return &clientRequest, nil
}

func getProjectRequest(r *http.Request) (*ProjectContainerRequest, *api.Error) {
	if r.Body == nil {
		return nil, api.NewError(nil, "Empty Body", api.InvalidJson)
	}

	var projectRequest ProjectContainerRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&projectRequest); err != nil {
		return nil, api.NewError(err, "Invalid JSON", api.InvalidJson)
	}
	defer api.CloseBody(r.Body)

	return &projectRequest, nil
}
func getProjectIdRequest(r *http.Request) (*ProjectIdRequest, *api.Error) {
	if r.Body == nil {
		return nil, api.NewError(nil, "Empty Body", api.InvalidJson)
	}

	var projectIdRequest ProjectIdRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&projectIdRequest); err != nil {
		return nil, api.NewError(err, "Invalid JSON", api.InvalidJson)
	}
	defer api.CloseBody(r.Body)

	return &projectIdRequest, nil
}

func NewClientResponse(clientData *Client) *ClientResponse {
	if clientData == nil {
		return nil
	}

	return &ClientResponse{
		ClientId: clientData.ClientId,
		Name:     clientData.ClientName,
		Address:  clientData.Address.String,
	}
}

func NewClientsResponse(clients []*Client) []*ClientResponse {
	if clients == nil {
		return nil
	}

	var response []*ClientResponse
	for _, c := range clients {
		response = append(response, NewClientResponse(c))
	}

	return response
}

func NewProjectResponse(project *Project) *ProjectResponse {

	var tasks []ProjectTaskResponse
	for _, t := range project.Tasks {
		tasks = append(tasks, NewProjectTaskResponse(&t))
	}

	return &ProjectResponse{
		ProjectId:     project.ProjectId,
		ProjectName:   project.ProjectName,
		ProjectActive: project.ProjectActive,
		Code:          project.Code.String,
		ClientId:      project.Client.ClientId,
		ClientName:    project.Client.ClientName,
		Tasks:         tasks,
	}
}

func NewProjectTaskResponse(t *task.ProjectTask) ProjectTaskResponse {
	return ProjectTaskResponse{
		TaskId:   t.TaskId,
		Billable: t.Billable,
		Rate:     t.Rate.Float64,
		Name:     t.Name,
		Active:   t.ProjectActive,
	}
}

func NewProjectsResponse(projects []*Project) []*ProjectResponse {
	var response []*ProjectResponse

	if projects != nil {
		for _, p := range projects {
			response = append(response, NewProjectResponse(p))
		}
	}

	return response
}

func createProjectTaskList(projectId int, taskRequests []TaskRequest) []task.ProjectTask {
	var projectTasks []task.ProjectTask
	for _, projectTask := range taskRequests {
		projectTasks = append(projectTasks,
			task.ProjectTask{
				Task: task.Task{
					TaskId: projectTask.Id,
				},
				ProjectId:     projectId,
				Rate:          valid.ToNullFloat64(projectTask.Rate),
				Billable:      projectTask.Billable,
				ProjectActive: true,
			})
	}
	return projectTasks
}
