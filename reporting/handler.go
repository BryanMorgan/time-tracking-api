package reporting

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/bryanmorgan/time-tracking-api/api"
	"github.com/bryanmorgan/time-tracking-api/config"
	"github.com/bryanmorgan/time-tracking-api/logger"
	"github.com/bryanmorgan/time-tracking-api/profile"
	"github.com/bryanmorgan/time-tracking-api/valid"
)

type ClientReportResponse struct {
	ClientId         int     `json:"clientId"`
	ClientName       string  `json:"clientName"`
	NonBillableHours float64 `json:"nonBillableHours"`
	BillableHours    float64 `json:"billableHours"`
	BillableTotal    float64 `json:"billableTotal"`
}

type ProjectReportResponse struct {
	ProjectId        int     `json:"projectId"`
	ProjectName      string  `json:"projectName"`
	ClientName       string  `json:"clientName"`
	NonBillableHours float64 `json:"nonBillableHours"`
	BillableHours    float64 `json:"billableHours"`
	BillableTotal    float64 `json:"billableTotal"`
}

type TaskReportResponse struct {
	TaskId           int     `json:"taskId"`
	ClientId         int     `json:"clientId"`
	TaskName         string  `json:"taskName"`
	ClientName       string  `json:"clientName"`
	NonBillableHours float64 `json:"nonBillableHours"`
	BillableHours    float64 `json:"billableHours"`
	BillableTotal    float64 `json:"billableTotal"`
}

type PersonReportResponse struct {
	ProfileId        int     `json:"profileId"`
	FirstName        string  `json:"firstName"`
	LastName         string  `json:"lastName"`
	NonBillableHours float64 `json:"nonBillableHours"`
	BillableHours    float64 `json:"billableHours"`
	BillableTotal    float64 `json:"billableTotal"`
}

func (a *ReportingRouter) getTimeByClient(w http.ResponseWriter, r *http.Request) {
	fromDateString := r.URL.Query().Get("from")
	toDateString := r.URL.Query().Get("to")
	offsetString := r.URL.Query().Get("page")

	if valid.IsNull(fromDateString) {
		api.ErrorJson(w, api.NewFieldError(nil, "No from parameter", api.InvalidField, "from"), http.StatusBadRequest)
		return
	}

	var toDate time.Time
	var fromDate time.Time
	fromDate, err := time.Parse(config.ISOShortDateFormat, fromDateString)
	if err != nil {
		api.ErrorJson(w, api.NewFieldError(err, "Invalid format. Use ISO8061: YYYY-MM-DD", api.InvalidField, fromDateString), http.StatusBadRequest)
		return
	}

	if !valid.IsNull(toDateString) {
		// Make sure the date string is in a valid ISO 8061 format
		toDate, err = time.Parse(config.ISOShortDateFormat, toDateString)
		if err != nil {
			api.ErrorJson(w, api.NewFieldError(err, "Invalid format. Use ISO8061: YYYY-MM-DD", api.InvalidField, toDateString), http.StatusBadRequest)
			return
		}
	} else {
		toDate = time.Now()
	}

	offset := 0
	if !valid.IsNull(offsetString) {
		offset, err = strconv.Atoi(offsetString)
		if err != nil {
			api.ErrorJson(w, api.NewFieldError(err, "Invalid page offset", api.InvalidField, offsetString), http.StatusInternalServerError)
			return
		}
	}

	userProfile, ok := r.Context().Value(config.ProfileContextKey).(*profile.Profile)
	if !ok || userProfile == nil {
		api.ErrorJson(w, api.NewError(nil, "Invalid profile context", api.SystemError), http.StatusInternalServerError)
		return
	}

	fromDate, toDate = AdjustForWeekStart(fromDate, toDate, userProfile.WeekStart, time.Now())

	reportRows, apperr := a.ReportingService.GetTimeByClient(userProfile.AccountId, fromDate, toDate, offset)
	if apperr != nil {
		api.ErrorJson(w, apperr, http.StatusInternalServerError)
		return
	}

	api.Json(w, r, NewClientReportsResponse(reportRows))
}

func (a *ReportingRouter) getTimeByProject(w http.ResponseWriter, r *http.Request) {
	fromDateString := r.URL.Query().Get("from")
	toDateString := r.URL.Query().Get("to")
	offsetString := r.URL.Query().Get("page")

	if valid.IsNull(fromDateString) {
		api.ErrorJson(w, api.NewFieldError(nil, "No from parameter", api.InvalidField, "from"), http.StatusBadRequest)
		return
	}

	var toDate time.Time
	var fromDate time.Time
	fromDate, err := time.Parse(config.ISOShortDateFormat, fromDateString)
	if err != nil {
		api.ErrorJson(w, api.NewFieldError(err, "Invalid format. Use ISO8061: YYYY-MM-DD", api.InvalidField, fromDateString), http.StatusBadRequest)
		return
	}

	if !valid.IsNull(toDateString) {
		// Make sure the date string is in a valid ISO 8061 format
		toDate, err = time.Parse(config.ISOShortDateFormat, toDateString)
		if err != nil {
			api.ErrorJson(w, api.NewFieldError(err, "Invalid format. Use ISO8061: YYYY-MM-DD", api.InvalidField, toDateString), http.StatusBadRequest)
			return
		}
	} else {
		toDate = time.Now()
	}

	offset := 0
	if !valid.IsNull(offsetString) {
		offset, err = strconv.Atoi(offsetString)
		if err != nil {
			api.ErrorJson(w, api.NewFieldError(err, "Invalid page offset", api.InvalidField, offsetString), http.StatusInternalServerError)
			return
		}
	}

	userProfile, ok := r.Context().Value(config.ProfileContextKey).(*profile.Profile)
	if !ok || userProfile == nil {
		api.ErrorJson(w, api.NewError(nil, "Invalid profile context", api.SystemError), http.StatusInternalServerError)
		return
	}

	reportRows, apperr := a.ReportingService.GetTimeByProject(userProfile.AccountId, fromDate, toDate, offset)
	if apperr != nil {
		api.ErrorJson(w, apperr, http.StatusInternalServerError)
		return
	}

	api.Json(w, r, NewProjectReportsResponse(reportRows))
}

func (a *ReportingRouter) getTimeByTask(w http.ResponseWriter, r *http.Request) {
	fromDateString := r.URL.Query().Get("from")
	toDateString := r.URL.Query().Get("to")
	offsetString := r.URL.Query().Get("page")

	if valid.IsNull(fromDateString) {
		api.ErrorJson(w, api.NewFieldError(nil, "No from parameter", api.InvalidField, "from"), http.StatusBadRequest)
		return
	}

	var toDate time.Time
	var fromDate time.Time
	fromDate, err := time.Parse(config.ISOShortDateFormat, fromDateString)
	if err != nil {
		api.ErrorJson(w, api.NewFieldError(err, "Invalid format. Use ISO8061: YYYY-MM-DD", api.InvalidField, fromDateString), http.StatusBadRequest)
		return
	}

	if !valid.IsNull(toDateString) {
		// Make sure the date string is in a valid ISO 8061 format
		toDate, err = time.Parse(config.ISOShortDateFormat, toDateString)
		if err != nil {
			api.ErrorJson(w, api.NewFieldError(err, "Invalid format. Use ISO8061: YYYY-MM-DD", api.InvalidField, toDateString), http.StatusBadRequest)
			return
		}
	} else {
		toDate = time.Now()
	}

	offset := 0
	if !valid.IsNull(offsetString) {
		offset, err = strconv.Atoi(offsetString)
		if err != nil {
			api.ErrorJson(w, api.NewFieldError(err, "Invalid page offset", api.InvalidField, offsetString), http.StatusInternalServerError)
			return
		}
	}

	userProfile, ok := r.Context().Value(config.ProfileContextKey).(*profile.Profile)
	if !ok || userProfile == nil {
		api.ErrorJson(w, api.NewError(nil, "Invalid profile context", api.SystemError), http.StatusInternalServerError)
		return
	}

	reportRows, apperr := a.ReportingService.GetTimeByTask(userProfile.AccountId, fromDate, toDate, offset)
	if apperr != nil {
		api.ErrorJson(w, apperr, http.StatusInternalServerError)
		return
	}

	api.Json(w, r, NewTaskReportsResponse(reportRows))
}

func (a *ReportingRouter) getTimeByPerson(w http.ResponseWriter, r *http.Request) {
	fromDateString := r.URL.Query().Get("from")
	toDateString := r.URL.Query().Get("to")
	offsetString := r.URL.Query().Get("page")

	if valid.IsNull(fromDateString) {
		api.ErrorJson(w, api.NewFieldError(nil, "No from parameter", api.InvalidField, "from"), http.StatusBadRequest)
		return
	}

	var toDate time.Time
	var fromDate time.Time
	fromDate, err := time.Parse(config.ISOShortDateFormat, fromDateString)
	if err != nil {
		api.ErrorJson(w, api.NewFieldError(err, "Invalid format. Use ISO8061: YYYY-MM-DD", api.InvalidField, fromDateString), http.StatusBadRequest)
		return
	}

	if !valid.IsNull(toDateString) {
		// Make sure the date string is in a valid ISO 8061 format
		toDate, err = time.Parse(config.ISOShortDateFormat, toDateString)
		if err != nil {
			api.ErrorJson(w, api.NewFieldError(err, "Invalid format. Use ISO8061: YYYY-MM-DD", api.InvalidField, toDateString), http.StatusBadRequest)
			return
		}
	} else {
		toDate = time.Now()
	}

	offset := 0
	if !valid.IsNull(offsetString) {
		offset, err = strconv.Atoi(offsetString)
		if err != nil {
			api.ErrorJson(w, api.NewFieldError(err, "Invalid page offset", api.InvalidField, offsetString), http.StatusInternalServerError)
			return
		}
	}

	userProfile, ok := r.Context().Value(config.ProfileContextKey).(*profile.Profile)
	if !ok || userProfile == nil {
		api.ErrorJson(w, api.NewError(nil, "Invalid profile context", api.SystemError), http.StatusInternalServerError)
		return
	}

	reportRows, apperr := a.ReportingService.GetTimeByPerson(userProfile.AccountId, fromDate, toDate, offset)
	if apperr != nil {
		api.ErrorJson(w, apperr, http.StatusInternalServerError)
		return
	}

	api.Json(w, r, NewPersonReportsResponse(reportRows))
}

func (a *ReportingRouter) exportTimeByClient(w http.ResponseWriter, r *http.Request) {
	fromDateString := r.URL.Query().Get("from")
	toDateString := r.URL.Query().Get("to")

	if valid.IsNull(fromDateString) {
		api.ErrorJson(w, api.NewFieldError(nil, "No from parameter", api.InvalidField, "from"), http.StatusBadRequest)
		return
	}

	var toDate time.Time
	var fromDate time.Time
	fromDate, err := time.Parse(config.ISOShortDateFormat, fromDateString)
	if err != nil {
		api.ErrorJson(w, api.NewFieldError(err, "Invalid format. Use ISO8061: YYYY-MM-DD", api.InvalidField, fromDateString), http.StatusBadRequest)
		return
	}

	if !valid.IsNull(toDateString) {
		// Make sure the date string is in a valid ISO 8061 format
		toDate, err = time.Parse(config.ISOShortDateFormat, toDateString)
		if err != nil {
			api.ErrorJson(w, api.NewFieldError(err, "Invalid format. Use ISO8061: YYYY-MM-DD", api.InvalidField, toDateString), http.StatusBadRequest)
			return
		}
	} else {
		toDate = time.Now()
	}

	userProfile, ok := r.Context().Value(config.ProfileContextKey).(*profile.Profile)
	if !ok || userProfile == nil {
		api.ErrorJson(w, api.NewError(nil, "Invalid profile context", api.SystemError), http.StatusInternalServerError)
		return
	}

	reportRows, apperr := a.ReportingService.GetTimeByClient(userProfile.AccountId, fromDate, toDate, 0)
	if apperr != nil {
		api.ErrorJson(w, apperr, http.StatusInternalServerError)
		return
	}

	WriteExportClientReportsResponse(w, userProfile.Account.Company, fromDate, toDate, reportRows)
}

func (a *ReportingRouter) exportTimeByProject(w http.ResponseWriter, r *http.Request) {
	fromDateString := r.URL.Query().Get("from")
	toDateString := r.URL.Query().Get("to")

	if valid.IsNull(fromDateString) {
		api.ErrorJson(w, api.NewFieldError(nil, "No from parameter", api.InvalidField, "from"), http.StatusBadRequest)
		return
	}

	var toDate time.Time
	var fromDate time.Time
	fromDate, err := time.Parse(config.ISOShortDateFormat, fromDateString)
	if err != nil {
		api.ErrorJson(w, api.NewFieldError(err, "Invalid format. Use ISO8061: YYYY-MM-DD", api.InvalidField, fromDateString), http.StatusBadRequest)
		return
	}

	if !valid.IsNull(toDateString) {
		// Make sure the date string is in a valid ISO 8061 format
		toDate, err = time.Parse(config.ISOShortDateFormat, toDateString)
		if err != nil {
			api.ErrorJson(w, api.NewFieldError(err, "Invalid format. Use ISO8061: YYYY-MM-DD", api.InvalidField, toDateString), http.StatusBadRequest)
			return
		}
	} else {
		toDate = time.Now()
	}

	userProfile, ok := r.Context().Value(config.ProfileContextKey).(*profile.Profile)
	if !ok || userProfile == nil {
		api.ErrorJson(w, api.NewError(nil, "Invalid profile context", api.SystemError), http.StatusInternalServerError)
		return
	}

	reportRows, apperr := a.ReportingService.GetTimeByProject(userProfile.AccountId, fromDate, toDate, 0)
	if apperr != nil {
		api.ErrorJson(w, apperr, http.StatusInternalServerError)
		return
	}

	WriteExportProjectReportsResponse(w, userProfile.Account.Company, fromDate, toDate, reportRows)
}

func (a *ReportingRouter) exportTimeByTask(w http.ResponseWriter, r *http.Request) {
	fromDateString := r.URL.Query().Get("from")
	toDateString := r.URL.Query().Get("to")

	if valid.IsNull(fromDateString) {
		api.ErrorJson(w, api.NewFieldError(nil, "No from parameter", api.InvalidField, "from"), http.StatusBadRequest)
		return
	}

	var toDate time.Time
	var fromDate time.Time
	fromDate, err := time.Parse(config.ISOShortDateFormat, fromDateString)
	if err != nil {
		api.ErrorJson(w, api.NewFieldError(err, "Invalid format. Use ISO8061: YYYY-MM-DD", api.InvalidField, fromDateString), http.StatusBadRequest)
		return
	}

	if !valid.IsNull(toDateString) {
		// Make sure the date string is in a valid ISO 8061 format
		toDate, err = time.Parse(config.ISOShortDateFormat, toDateString)
		if err != nil {
			api.ErrorJson(w, api.NewFieldError(err, "Invalid format. Use ISO8061: YYYY-MM-DD", api.InvalidField, toDateString), http.StatusBadRequest)
			return
		}
	} else {
		toDate = time.Now()
	}

	userProfile, ok := r.Context().Value(config.ProfileContextKey).(*profile.Profile)
	if !ok || userProfile == nil {
		api.ErrorJson(w, api.NewError(nil, "Invalid profile context", api.SystemError), http.StatusInternalServerError)
		return
	}

	reportRows, apperr := a.ReportingService.GetTimeByTask(userProfile.AccountId, fromDate, toDate, 0)
	if apperr != nil {
		api.ErrorJson(w, apperr, http.StatusInternalServerError)
		return
	}

	WriteExportTaskReportsResponse(w, userProfile.Account.Company, fromDate, toDate, reportRows)
}

func (a *ReportingRouter) exportTimeByPerson(w http.ResponseWriter, r *http.Request) {
	fromDateString := r.URL.Query().Get("from")
	toDateString := r.URL.Query().Get("to")

	if valid.IsNull(fromDateString) {
		api.ErrorJson(w, api.NewFieldError(nil, "No from parameter", api.InvalidField, "from"), http.StatusBadRequest)
		return
	}

	var toDate time.Time
	var fromDate time.Time
	fromDate, err := time.Parse(config.ISOShortDateFormat, fromDateString)
	if err != nil {
		api.ErrorJson(w, api.NewFieldError(err, "Invalid format. Use ISO8061: YYYY-MM-DD", api.InvalidField, fromDateString), http.StatusBadRequest)
		return
	}

	if !valid.IsNull(toDateString) {
		// Make sure the date string is in a valid ISO 8061 format
		toDate, err = time.Parse(config.ISOShortDateFormat, toDateString)
		if err != nil {
			api.ErrorJson(w, api.NewFieldError(err, "Invalid format. Use ISO8061: YYYY-MM-DD", api.InvalidField, toDateString), http.StatusBadRequest)
			return
		}
	} else {
		toDate = time.Now()
	}

	userProfile, ok := r.Context().Value(config.ProfileContextKey).(*profile.Profile)
	if !ok || userProfile == nil {
		api.ErrorJson(w, api.NewError(nil, "Invalid profile context", api.SystemError), http.StatusInternalServerError)
		return
	}

	reportRows, apperr := a.ReportingService.GetTimeByPerson(userProfile.AccountId, fromDate, toDate, 0)
	if apperr != nil {
		api.ErrorJson(w, apperr, http.StatusInternalServerError)
		return
	}

	WriteExportPersonReportsResponse(w, userProfile.Account.Company, fromDate, toDate, reportRows)
}

/*
 * Adjust the from/to dates by the weekStart from the profile
 * All dates assume 0 = Sunday and 6 = Saturday
 */
func AdjustForWeekStart(fromDate time.Time, toDate time.Time, weekStart int, today time.Time) (time.Time, time.Time) {
	todayWeekday := today.Weekday()
	if weekStart <= int(todayWeekday) {
		fromDate = fromDate.AddDate(0, 0, weekStart)
		toDate = toDate.AddDate(0, 0, weekStart)
	} else {
		fromDate = fromDate.AddDate(0, 0, -7+weekStart)
		toDate = toDate.AddDate(0, 0, -7+weekStart)
	}

	return fromDate, toDate
}

func NewClientReportResponse(report *ClientReport) *ClientReportResponse {
	if report == nil {
		return nil
	}

	return &ClientReportResponse{
		ClientId:         report.ClientId,
		ClientName:       report.ClientName,
		NonBillableHours: report.NonBillableHours.Float64,
		BillableHours:    report.BillableHours.Float64,
		BillableTotal:    report.BillableTotal.Float64,
	}
}

func ExportClientReportResponse(report *ClientReport) []string {
	if report == nil {
		return []string{}
	}

	result := make([]string, 4)
	result[0] = report.ClientName
	result[1] = fmt.Sprintf(" %0.2f", report.NonBillableHours.Float64)
	result[2] = fmt.Sprintf(" %0.2f", report.BillableHours.Float64)
	result[3] = fmt.Sprintf(" %0.2f", report.BillableTotal.Float64)

	return result
}

func ExportProjectReportResponse(report *ProjectReport) []string {
	if report == nil {
		return []string{}
	}

	result := make([]string, 5)
	result[0] = report.ClientName
	result[1] = report.ProjectName
	result[2] = fmt.Sprintf(" %0.2f", report.NonBillableHours.Float64)
	result[3] = fmt.Sprintf(" %0.2f", report.BillableHours.Float64)
	result[4] = fmt.Sprintf(" %0.2f", report.BillableTotal.Float64)

	return result
}

func ExportTaskReportResponse(report *TaskReport) []string {
	if report == nil {
		return []string{}
	}

	result := make([]string, 5)
	result[0] = report.TaskName
	result[1] = fmt.Sprintf(" %0.2f", report.NonBillableHours.Float64)
	result[2] = fmt.Sprintf(" %0.2f", report.BillableHours.Float64)
	result[3] = fmt.Sprintf(" %0.2f", report.BillableTotal.Float64)

	return result
}

func ExportPersonReportResponse(report *PersonReport) []string {
	if report == nil {
		return []string{}
	}

	result := make([]string, 5)
	result[0] = report.LastName
	result[1] = report.FirstName
	result[2] = fmt.Sprintf(" %0.2f", report.NonBillableHours.Float64)
	result[3] = fmt.Sprintf(" %0.2f", report.BillableHours.Float64)
	result[4] = fmt.Sprintf(" %0.2f", report.BillableTotal.Float64)

	return result
}

func NewClientReportsResponse(clientReportRows []*ClientReport) []*ClientReportResponse {
	if clientReportRows == nil {
		return []*ClientReportResponse{}
	}

	var response []*ClientReportResponse
	for _, t := range clientReportRows {
		response = append(response, NewClientReportResponse(t))
	}

	return response
}

func NewProjectReportResponse(report *ProjectReport) *ProjectReportResponse {
	if report == nil {
		return nil
	}

	return &ProjectReportResponse{
		ProjectId:        report.ProjectId,
		ProjectName:      report.ProjectName,
		ClientName:       report.ClientName,
		NonBillableHours: report.NonBillableHours.Float64,
		BillableHours:    report.BillableHours.Float64,
		BillableTotal:    report.BillableTotal.Float64,
	}
}

func NewProjectReportsResponse(reportRows []*ProjectReport) []*ProjectReportResponse {
	if reportRows == nil {
		return []*ProjectReportResponse{}
	}

	var response []*ProjectReportResponse
	for _, t := range reportRows {
		response = append(response, NewProjectReportResponse(t))
	}

	return response
}

func writeExportCsvHeader(w http.ResponseWriter, fromDate time.Time, toDate time.Time, companyName string) {
	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	var cleanCompanyName string
	if err != nil {
		logger.Log.Error("Failed to compile regular expression: " + err.Error())
		cleanCompanyName = companyName
	} else {
		cleanCompanyName = reg.ReplaceAllString(companyName, "-")
	}

	header := w.Header()
	header.Set("Content-Type", "text/csv")
	header.Set("Content-Disposition",
		fmt.Sprintf("attachment;filename=export_%s_%s_to_%s.csv",
			cleanCompanyName,
			fromDate.Format(config.ISOShortDateFormat),
			toDate.Format(config.ISOShortDateFormat)))

}

func WriteExportClientReportsResponse(w http.ResponseWriter, companyName string, fromDate time.Time, toDate time.Time, clientReportRows []*ClientReport) {
	writeExportCsvHeader(w, fromDate, toDate, companyName)

	wr := csv.NewWriter(w)
	if clientReportRows != nil {
		header := []string{
			"Client Name",
			"Non-Billable Hours",
			"Billable Hours",
			"Billable Total",
		}

		err := wr.Write(header)
		if err != nil {
			logger.Log.Error("Failed to write row: " + err.Error())
		}

		for _, row := range clientReportRows {
			err := wr.Write(ExportClientReportResponse(row))
			if err != nil {
				logger.Log.Error("Failed to write row: " + err.Error())
			}
		}
		wr.Flush()
	}
}

func WriteExportProjectReportsResponse(w http.ResponseWriter, companyName string, fromDate time.Time, toDate time.Time, projectReportRows []*ProjectReport) {
	writeExportCsvHeader(w, fromDate, toDate, companyName)

	wr := csv.NewWriter(w)
	if projectReportRows != nil {
		header := []string{
			"Client Name",
			"Project Name",
			"Non-Billable Hours",
			"Billable Hours",
			"Billable Total",
		}

		err := wr.Write(header)
		if err != nil {
			logger.Log.Error("Failed to write row: " + err.Error())
		}

		for _, row := range projectReportRows {
			err := wr.Write(ExportProjectReportResponse(row))
			if err != nil {
				logger.Log.Error("Failed to write row: " + err.Error())
			}
		}
		wr.Flush()
	}
}

func WriteExportTaskReportsResponse(w http.ResponseWriter, companyName string, fromDate time.Time, toDate time.Time, taskReportRows []*TaskReport) {
	writeExportCsvHeader(w, fromDate, toDate, companyName)

	wr := csv.NewWriter(w)
	if taskReportRows != nil {
		header := []string{
			"Task Name",
			"Non-Billable Hours",
			"Billable Hours",
			"Billable Total",
		}

		err := wr.Write(header)
		if err != nil {
			logger.Log.Error("Failed to write row: " + err.Error())
		}

		for _, row := range taskReportRows {
			err := wr.Write(ExportTaskReportResponse(row))
			if err != nil {
				logger.Log.Error("Failed to write row: " + err.Error())
			}
		}
		wr.Flush()
	}
}

func WriteExportPersonReportsResponse(w http.ResponseWriter, companyName string, fromDate time.Time, toDate time.Time, personReportRows []*PersonReport) {
	writeExportCsvHeader(w, fromDate, toDate, companyName)

	wr := csv.NewWriter(w)
	if personReportRows != nil {
		header := []string{
			"Last Name",
			"First Name",
			"Non-Billable Hours",
			"Billable Hours",
			"Billable Total",
		}

		err := wr.Write(header)
		if err != nil {
			logger.Log.Error("Failed to write row: " + err.Error())
		}
		for _, row := range personReportRows {
			err := wr.Write(ExportPersonReportResponse(row))
			if err != nil {
				logger.Log.Error("Failed to write row: " + err.Error())
			}
		}
		wr.Flush()
	}
}

func NewTaskReportResponse(report *TaskReport) *TaskReportResponse {
	if report == nil {
		return nil
	}

	return &TaskReportResponse{
		TaskId:           report.TaskId,
		ClientId:         report.ClientId,
		TaskName:         report.TaskName,
		NonBillableHours: report.NonBillableHours.Float64,
		BillableHours:    report.BillableHours.Float64,
		BillableTotal:    report.BillableTotal.Float64,
	}
}

func NewTaskReportsResponse(reportRows []*TaskReport) []*TaskReportResponse {
	if reportRows == nil {
		return []*TaskReportResponse{}
	}

	var response []*TaskReportResponse
	for _, t := range reportRows {
		response = append(response, NewTaskReportResponse(t))
	}

	return response
}

func NewPersonReportResponse(report *PersonReport) *PersonReportResponse {
	if report == nil {
		return nil
	}

	return &PersonReportResponse{
		ProfileId:        report.ProfileId,
		FirstName:        report.FirstName,
		LastName:         report.LastName,
		NonBillableHours: report.NonBillableHours.Float64,
		BillableHours:    report.BillableHours.Float64,
		BillableTotal:    report.BillableTotal.Float64,
	}
}

func NewPersonReportsResponse(reportRows []*PersonReport) []*PersonReportResponse {
	if reportRows == nil {
		return []*PersonReportResponse{}
	}

	var response []*PersonReportResponse
	for _, t := range reportRows {
		response = append(response, NewPersonReportResponse(t))
	}

	return response
}
