package timesheet

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/bryanmorgan/time-tracking-api/api"
	"github.com/bryanmorgan/time-tracking-api/config"
	"github.com/bryanmorgan/time-tracking-api/profile"
	"github.com/bryanmorgan/time-tracking-api/valid"

	"github.com/go-chi/chi"
)

type TimeEntryRequest struct {
	Day       string
	Hours     float64
	ProjectId int
	TaskId    int
}

type TimeEntryRangeRequest struct {
	Entries []TimeEntryRequest
}

type ProjectWeekRequest struct {
	StartDate string
	EndDate   string
	ProjectId int
	TaskId    int
}

type ProjectDeleteRequest struct {
	StartDate string
	EndDate   string
	ProjectId int
	TaskId    int
}

type TimeEntryResponse struct {
	Day         string  `json:"day"`
	Hours       float64 `json:"hours"`
	ProjectId   int     `json:"projectId"`
	TaskId      int     `json:"taskId"`
	ClientName  string  `json:"clientName"`
	ProjectName string  `json:"projectName"`
	TaskName    string  `json:"taskName"`
}

type TimeRangeResponse struct {
	Start       string               `json:"start"`
	End         string               `json:"end"`
	TimeEntries []*TimeEntryResponse `json:"entries"`
}

const (
	startDatePathParameter = "startDate"
)

func (a *TimeRouter) getTimeEntriesForWeek(w http.ResponseWriter, r *http.Request) {
	userProfile, ok := r.Context().Value(config.ProfileContextKey).(*profile.Profile)
	if !ok || userProfile == nil {
		api.ErrorJson(w, api.NewError(nil, "Invalid profile context", api.SystemError), http.StatusInternalServerError)
		return
	}

	// Get the start weekday from the account
	weekday := getWeekdayStart(userProfile.WeekStart)

	var start, end time.Time
	var err error

	startDateString := chi.URLParam(r, startDatePathParameter)
	if !valid.IsNull(startDateString) {
		// Make sure the date string is in a valid ISO 8061 format
		_, err = time.Parse(config.ISOShortDateFormat, startDateString)
		if err != nil {
			api.ErrorJson(w, api.NewFieldError(err, "Invalid format. Use ISO8061: YYYY-MM-DD", api.InvalidField, startDatePathParameter), http.StatusBadRequest)
			return
		}

		start, end, err = getWeekRangeFromDate(startDateString, weekday)
		if err != nil {
			api.ErrorJson(w, api.NewError(err, "Could not get week range", api.SystemError), http.StatusInternalServerError)
			return
		}
	} else {
		start, end, err = getCurrentWeekRange(userProfile.Timezone, weekday)
		if err != nil {
			api.ErrorJson(w, api.NewError(err, "Could not get current week range", api.SystemError), http.StatusInternalServerError)
			return
		}
	}

	timeEntries, serviceErr := a.timeService.GetTimeEntriesForRange(userProfile.ProfileId, userProfile.AccountId, start, end)
	if serviceErr != nil {
		api.ErrorJson(w, serviceErr, http.StatusInternalServerError)
		return
	}

	api.Json(w, r, NewTimeRange(timeEntries, start, end))
}

func (a *TimeRouter) saveTimeEntries(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		api.ErrorJson(w, api.NewError(nil, "Empty Body", api.InvalidJson), http.StatusBadRequest)
		return
	}

	var entryRequest TimeEntryRangeRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&entryRequest); err != nil {
		api.ErrorJson(w, api.NewError(err, "Invalid JSON", api.InvalidJson), http.StatusBadRequest)
		return
	}
	defer api.CloseBody(r.Body)

	// If there are no entries do nothing
	if entryRequest.Entries == nil {
		api.ErrorJson(w, api.NewError(nil, "Invalid JSON. No time entries array.", api.InvalidJson), http.StatusBadRequest)
		return
	}

	userProfile, ok := r.Context().Value(config.ProfileContextKey).(*profile.Profile)
	if !ok || userProfile == nil {
		api.ErrorJson(w, api.NewError(nil, "Invalid profile context", api.SystemError), http.StatusInternalServerError)
		return
	}

	// Validate profile and account data and build TimeEntry objects
	var entryData []*TimeEntry
	for _, entry := range entryRequest.Entries {
		entryDate, err := time.Parse(config.ISOShortDateFormat, entry.Day)
		if err != nil {
			api.ErrorJson(w, api.NewError(err, "Invalid entry date", api.InvalidField), http.StatusBadRequest)
			return
		}

		if entry.ProjectId <= 0 || entry.TaskId <= 0 {
			api.ErrorJson(w, api.NewError(err, "Missing or invalid project id or task id", api.InvalidField), http.StatusBadRequest)
			return
		}
		entryData = append(entryData, &TimeEntry{
			Day:       entryDate,
			Hours:     entry.Hours,
			ProfileId: userProfile.ProfileId,
			AccountId: userProfile.AccountId,
			ProjectId: entry.ProjectId,
			TaskId:    entry.TaskId,
		})
	}

	err := a.timeService.SaveOrUpdateTimeEntries(entryData)
	if err != nil {
		api.ErrorJson(w, err, http.StatusInternalServerError)
		return
	}

	api.Json(w, r, nil)
}

func (a *TimeRouter) updateTimeEntries(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		api.ErrorJson(w, api.NewError(nil, "Empty Body", api.InvalidJson), http.StatusBadRequest)
		return
	}

	var entryRequest TimeEntryRangeRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&entryRequest); err != nil {
		api.ErrorJson(w, api.NewError(err, "Invalid JSON", api.InvalidJson), http.StatusBadRequest)
		return
	}
	defer api.CloseBody(r.Body)

	// If there are no entries do nothing
	if entryRequest.Entries == nil {
		api.ErrorJson(w, api.NewError(nil, "Invalid JSON. No time entries.", api.InvalidJson), http.StatusBadRequest)
		return
	}

	userProfile, ok := r.Context().Value(config.ProfileContextKey).(*profile.Profile)
	if !ok || userProfile == nil {
		api.ErrorJson(w, api.NewError(nil, "Invalid profile context", api.SystemError), http.StatusInternalServerError)
		return
	}

	// Build TimeEntry objects
	var entryData []*TimeEntry
	for _, entry := range entryRequest.Entries {
		entryDate, err := time.Parse(config.ISOShortDateFormat, entry.Day)
		if err != nil {
			api.ErrorJson(w, api.NewError(err, "Invalid entry date", api.InvalidField), http.StatusBadRequest)
			return
		}

		if entry.ProjectId <= 0 || entry.TaskId <= 0 {
			api.ErrorJson(w, api.NewError(err, "Missing or invalid project id or task id", api.InvalidField), http.StatusBadRequest)
			return
		}

		entryData = append(entryData, &TimeEntry{
			AccountId: userProfile.AccountId,
			ProfileId: userProfile.ProfileId,
			ProjectId: entry.ProjectId,
			TaskId:    entry.TaskId,
			Day:       entryDate,
			Hours:     entry.Hours,
		})
	}

	err := a.timeService.UpdateTimeEntries(entryData)
	if err != nil {
		api.ErrorJson(w, err, http.StatusInternalServerError)
		return
	}

	api.Json(w, r, nil)
}

func (a *TimeRouter) addProjectToWeek(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		api.ErrorJson(w, api.NewError(nil, "Empty Body", api.InvalidJson), http.StatusBadRequest)
		return
	}

	var projectWeekRequest ProjectWeekRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&projectWeekRequest); err != nil {
		api.ErrorJson(w, api.NewError(err, "Invalid JSON", api.InvalidJson), http.StatusBadRequest)
		return
	}
	defer api.CloseBody(r.Body)

	// Ensure we got a start and end date
	if valid.IsNull(projectWeekRequest.StartDate) || valid.IsNull(projectWeekRequest.EndDate) {
		api.ErrorJson(w, api.NewError(nil, "Invalid start or end date", api.InvalidField), http.StatusBadRequest)
		return
	}

	if projectWeekRequest.ProjectId <= 0 {
		api.ErrorJson(w, api.NewError(nil, "Missing projectId", api.InvalidProject), http.StatusBadRequest)
		return
	}

	if projectWeekRequest.TaskId <= 0 {
		api.ErrorJson(w, api.NewError(nil, "Missing taskId", api.InvalidTask), http.StatusBadRequest)
		return
	}

	// Make sure the start and end date strings are in a valid ISO 8061 format
	start, err := time.Parse(config.ISOShortDateFormat, projectWeekRequest.StartDate)
	if err != nil {
		api.ErrorJson(w, api.NewError(err, "Invalid start date format. Use ISO8061: YYYY-MM-DD", api.InvalidField), http.StatusBadRequest)
		return
	}

	end, err := time.Parse(config.ISOShortDateFormat, projectWeekRequest.EndDate)
	if err != nil {
		api.ErrorJson(w, api.NewError(err, "Invalid end date format. Use ISO8061: YYYY-MM-DD", api.InvalidField), http.StatusBadRequest)
		return
	}

	userProfile, ok := r.Context().Value(config.ProfileContextKey).(*profile.Profile)
	if !ok || userProfile == nil {
		api.ErrorJson(w, api.NewError(nil, "Invalid profile context", api.SystemError), http.StatusInternalServerError)
		return
	}

	appErr := a.timeService.AddInitialProjectTimeEntries(userProfile.ProfileId, userProfile.AccountId, start, end, projectWeekRequest.ProjectId, projectWeekRequest.TaskId)
	if appErr != nil {
		api.ErrorJson(w, appErr, http.StatusInternalServerError)
		return
	}

	api.Json(w, r, nil)
}

func (a *TimeRouter) deleteProjectForWeek(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		api.ErrorJson(w, api.NewError(nil, "Empty body", api.InvalidJson), http.StatusBadRequest)
		return
	}

	var request ProjectDeleteRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&request); err != nil {
		api.ErrorJson(w, api.NewError(err, "Invalid JSON", api.InvalidJson), http.StatusBadRequest)
		return
	}
	defer api.CloseBody(r.Body)

	// Validate required fields
	if request.ProjectId <= 0 {
		api.ErrorJson(w, api.NewFieldError(nil, "Invalid or missing projectId", api.InvalidField, "projectId"), http.StatusBadRequest)
		return
	}

	if request.TaskId <= 0 {
		api.ErrorJson(w, api.NewFieldError(nil, "Invalid or missing taskId", api.InvalidField, "taskId"), http.StatusBadRequest)
		return
	}

	if request.StartDate == "" {
		api.ErrorJson(w, api.NewFieldError(nil, "Missing startDate", api.InvalidField, "startDate"), http.StatusBadRequest)
		return
	}

	if request.EndDate == "" {
		api.ErrorJson(w, api.NewFieldError(nil, "Missing endDate", api.InvalidField, "endDate"), http.StatusBadRequest)
		return
	}

	start, err := time.Parse(config.ISOShortDateFormat, request.StartDate)
	if err != nil {
		api.ErrorJson(w, api.NewFieldError(err, "Invalid startDate field. Use ISO8601 format: YYYY-MM-DD", api.InvalidField, "startDate"), http.StatusBadRequest)
		return
	}

	end, err := time.Parse(config.ISOShortDateFormat, request.EndDate)
	if err != nil {
		api.ErrorJson(w, api.NewFieldError(err, "Invalid endDate field. Use ISO8601 format: YYYY-MM-DD", api.InvalidField, "endDate"), http.StatusBadRequest)
		return
	}

	if end.After(start.AddDate(0, 0, 7)) {
		api.ErrorJson(w, api.NewError(nil, "Can only delete 1 week of data", api.InvalidField), http.StatusBadRequest)
		return
	}

	userProfile, ok := r.Context().Value(config.ProfileContextKey).(*profile.Profile)
	if !ok || userProfile == nil {
		api.ErrorJson(w, api.NewError(nil, "Invalid profile context", api.SystemError), http.StatusInternalServerError)
		return
	}

	serr := a.timeService.DeleteProjectForDates(userProfile.ProfileId, userProfile.AccountId, request.ProjectId, request.TaskId, start, end)
	if serr != nil {
		if serr.Code == api.SystemError {
			api.ErrorJson(w, serr, http.StatusInternalServerError)
		} else {
			api.ErrorJson(w, serr, http.StatusBadRequest)
		}
		return
	}

	api.Json(w, r, nil)
}

func NewTimeRange(timeEntries []*TimeEntry, start time.Time, end time.Time) *TimeRangeResponse {
	var response TimeRangeResponse
	response.Start = start.Format(config.ISOShortDateFormat)
	response.End = end.Format(config.ISOShortDateFormat)

	if timeEntries == nil {
		response.TimeEntries = []*TimeEntryResponse{}
		return &response
	}

	var timeRange []*TimeEntryResponse
	for _, t := range timeEntries {
		timeRange = append(timeRange, NewTimeEntryResponse(t))
	}

	response.TimeEntries = timeRange
	return &response
}

func NewTimeEntryResponse(entry *TimeEntry) *TimeEntryResponse {
	return &TimeEntryResponse{
		Day:         entry.Day.Format(config.ISOShortDateFormat),
		Hours:       entry.Hours,
		ClientName:  entry.ClientName,
		ProjectName: entry.ProjectName,
		TaskName:    entry.TaskName,
		ProjectId:   entry.ProjectId,
		TaskId:      entry.TaskId,
	}
}
