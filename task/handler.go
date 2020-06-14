package task

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/bryanmorgan/time-tracking-api/api"
	"github.com/bryanmorgan/time-tracking-api/config"
	"github.com/bryanmorgan/time-tracking-api/profile"
	"github.com/bryanmorgan/time-tracking-api/valid"
	"github.com/go-chi/chi"
)

type TaskRequest struct {
	Id              int
	Name            string
	DefaultRate     float64
	DefaultBillable bool
	Common          bool
}

type TaskResponse struct {
	Id              int     `json:"id,omitempty"`
	Name            string  `json:"name"`
	DefaultRate     float64 `json:"defaultRate,omitempty"`
	DefaultBillable bool    `json:"defaultBillable"`
	TaskActive      bool    `json:"taskActive"`
	Common          bool    `json:"common"`
}

func (a *TaskRouter) getTask(w http.ResponseWriter, r *http.Request) {
	taskIdString := chi.URLParam(r, "taskId")

	if valid.IsNull(taskIdString) {
		api.ErrorJson(w, api.NewFieldError(nil, "No taskId parameter", api.InvalidField, "taskId"), http.StatusBadRequest)
		return
	}

	taskId, err := strconv.Atoi(taskIdString)
	if err != nil || taskId <= 0 {
		api.ErrorJson(w, api.NewFieldError(err, "taskId not a number", api.InvalidField, "taskId"), http.StatusBadRequest)
		return
	}

	userProfile, ok := r.Context().Value(config.ProfileContextKey).(*profile.Profile)
	if !ok || userProfile == nil {
		api.ErrorJson(w, api.NewError(nil, "Invalid profile context", api.SystemError), http.StatusInternalServerError)
		return
	}

	task, apperr := a.taskService.GetTask(taskId, userProfile.AccountId)
	if apperr != nil {
		api.ErrorJson(w, apperr, http.StatusInternalServerError)
		return
	}

	if task == nil {
		api.ErrorJson(w, api.NewError(nil, "No task matching id", api.InvalidTask), http.StatusBadRequest)
		return
	}

	api.Json(w, r, NewTaskResponse(task))
}

func (a *TaskRouter) getAllTasks(w http.ResponseWriter, r *http.Request) {
	userProfile, ok := r.Context().Value(config.ProfileContextKey).(*profile.Profile)
	if !ok || userProfile == nil {
		api.ErrorJson(w, api.NewError(nil, "Invalid profile context", api.SystemError), http.StatusInternalServerError)
		return
	}

	tasks, apperr := a.taskService.GetAllTasks(userProfile.AccountId, true)
	if apperr != nil {
		api.ErrorJson(w, apperr, http.StatusInternalServerError)
		return
	}

	api.Json(w, r, NewTasksResponse(tasks))
}

func (a *TaskRouter) getArchivedTasks(w http.ResponseWriter, r *http.Request) {
	userProfile, ok := r.Context().Value(config.ProfileContextKey).(*profile.Profile)
	if !ok || userProfile == nil {
		api.ErrorJson(w, api.NewError(nil, "Invalid profile context", api.SystemError), http.StatusInternalServerError)
		return
	}

	tasks, apperr := a.taskService.GetAllTasks(userProfile.AccountId, false)
	if apperr != nil {
		api.ErrorJson(w, apperr, http.StatusInternalServerError)
		return
	}

	api.Json(w, r, NewTasksResponse(tasks))
}

func (a *TaskRouter) saveTask(w http.ResponseWriter, r *http.Request) {
	taskRequest, err := getTaskRequest(r)
	if err != nil {
		api.ErrorJson(w, err, http.StatusBadRequest)
		return
	}

	userProfile, ok := r.Context().Value(config.ProfileContextKey).(*profile.Profile)
	if !ok || userProfile == nil {
		api.ErrorJson(w, api.NewError(nil, "Invalid profile context", api.SystemError), http.StatusInternalServerError)
		return
	}

	if valid.IsNull(taskRequest.Name) {
		api.ErrorJson(w, api.NewFieldError(nil, "Missing required name", api.MissingField, "taskName"), http.StatusBadRequest)
		return
	}

	if taskRequest.DefaultRate <= 0.0 {
		taskRequest.DefaultRate = 0.0
	}

	task, err := a.taskService.SaveTask(userProfile.AccountId, taskRequest.Name, taskRequest.Common, taskRequest.DefaultRate, taskRequest.DefaultBillable)
	if err != nil {
		api.ErrorJson(w, err, http.StatusInternalServerError)
		return
	}

	api.Json(w, r, NewTaskResponse(task))
}

func (a *TaskRouter) updateTask(w http.ResponseWriter, r *http.Request) {
	taskRequest, err := getTaskRequest(r)
	if err != nil {
		api.ErrorJson(w, err, http.StatusBadRequest)
		return
	}

	if taskRequest.Id <= 0 {
		api.ErrorJson(w, api.NewFieldError(nil, "Missing id", api.MissingField, "id"), http.StatusBadRequest)
		return
	}

	userProfile, ok := r.Context().Value(config.ProfileContextKey).(*profile.Profile)
	if !ok || userProfile == nil {
		api.ErrorJson(w, api.NewError(nil, "Invalid profile context", api.SystemError), http.StatusInternalServerError)
		return
	}

	if valid.IsNull(taskRequest.Name) {
		api.ErrorJson(w, api.NewFieldError(nil, "Missing required name", api.MissingField, "taskName"), http.StatusBadRequest)
		return
	}

	if taskRequest.DefaultRate <= 0.0 {
		taskRequest.DefaultRate = 0.0
	}

	existingTask, err := a.taskService.GetTask(taskRequest.Id, userProfile.AccountId)
	if err != nil || existingTask == nil {
		api.ErrorJson(w, api.NewError(err, "Could not find task", api.SystemError), http.StatusInternalServerError)
		return
	}

	updateTask := *existingTask
	if !valid.IsNull(taskRequest.Name) {
		updateTask.Name = taskRequest.Name
	}

	updateTask.DefaultRate = valid.ToNullFloat64(taskRequest.DefaultRate)
	updateTask.DefaultBillable = taskRequest.DefaultBillable
	updateTask.Common = taskRequest.Common

	err = a.taskService.UpdateTask(&updateTask)
	if err != nil {
		api.ErrorJson(w, err, http.StatusInternalServerError)
		return
	}

	api.Json(w, r, nil)
}

func (a *TaskRouter) archiveTaskHandler(w http.ResponseWriter, r *http.Request) {
	request, err := getTaskRequest(r)
	if err != nil {
		api.ErrorJson(w, err, http.StatusBadRequest)
		return
	}

	if request.Id <= 0 {
		api.ErrorJson(w, api.NewError(nil, "Missing id", api.MissingField), http.StatusBadRequest)
		return
	}

	userProfile, ok := r.Context().Value(config.ProfileContextKey).(*profile.Profile)
	if !ok || userProfile == nil {
		api.ErrorJson(w, api.NewError(nil, "Invalid profile context", api.SystemError), http.StatusInternalServerError)
		return
	}

	err = a.taskService.ArchiveTask(request.Id, userProfile.AccountId)
	if err != nil {
		api.ErrorJson(w, api.NewError(err, "Failed to archive task", api.SystemError), http.StatusInternalServerError)
		return
	}

	api.Json(w, r, nil)
}

func (a *TaskRouter) restoreTaskHandler(w http.ResponseWriter, r *http.Request) {
	request, err := getTaskRequest(r)
	if err != nil {
		api.ErrorJson(w, err, http.StatusBadRequest)
		return
	}

	if request.Id <= 0 {
		api.ErrorJson(w, api.NewError(nil, "Missing id", api.MissingField), http.StatusBadRequest)
		return
	}

	userProfile, ok := r.Context().Value(config.ProfileContextKey).(*profile.Profile)
	if !ok || userProfile == nil {
		api.ErrorJson(w, api.NewError(nil, "Invalid profile context", api.SystemError), http.StatusInternalServerError)
		return
	}

	err = a.taskService.RestoreTask(request.Id, userProfile.AccountId)
	if err != nil {
		api.ErrorJson(w, api.NewError(err, "Failed to archive task", api.SystemError), http.StatusInternalServerError)
		return
	}

	api.Json(w, r, nil)
}

func (a *TaskRouter) deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	request, err := getTaskRequest(r)
	if err != nil {
		api.ErrorJson(w, err, http.StatusBadRequest)
		return
	}

	if request.Id <= 0 {
		api.ErrorJson(w, api.NewError(nil, "Missing id", api.MissingField), http.StatusBadRequest)
		return
	}

	userProfile, ok := r.Context().Value(config.ProfileContextKey).(*profile.Profile)
	if !ok || userProfile == nil {
		api.ErrorJson(w, api.NewError(nil, "Invalid profile context", api.SystemError), http.StatusInternalServerError)
		return
	}

	err = a.taskService.DeleteTask(request.Id, userProfile.AccountId)
	if err != nil {
		api.ErrorJson(w, api.NewError(err, "Failed to delete task", api.SystemError), http.StatusInternalServerError)
		return
	}

	api.Json(w, r, nil)
}

func getTaskRequest(r *http.Request) (*TaskRequest, *api.Error) {
	if r.Body == nil {
		return nil, api.NewError(nil, "Empty Body", api.InvalidJson)
	}

	var taskRequest TaskRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&taskRequest); err != nil {
		return nil, api.NewError(err, "Invalid JSON", api.InvalidJson)
	}
	defer api.CloseBody(r.Body)

	return &taskRequest, nil
}

func NewTaskResponse(task *Task) *TaskResponse {
	if task == nil {
		return nil
	}

	return &TaskResponse{
		Id:              task.TaskId,
		Name:            task.Name,
		DefaultRate:     task.DefaultRate.Float64,
		DefaultBillable: task.DefaultBillable,
		TaskActive:      task.TaskActive,
		Common:          task.Common,
	}
}

func NewTasksResponse(tasks []*Task) []*TaskResponse {
	if tasks == nil {
		return nil
	}

	var response []*TaskResponse
	for _, t := range tasks {
		response = append(response, NewTaskResponse(t))
	}

	return response
}
