// +build integration

package integration_test

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bryanmorgan/time-tracking-api/api"
	_ "github.com/bryanmorgan/time-tracking-api/config"
)

type timeEntryResponse struct {
	Day         string
	Hours       float64
	ClientName  string
	ProjectName string
	TaskName    string
	ProjectId   int
	TaskId      int
}
type timeDataResponse struct {
	Entries []timeEntryResponse
}

func TestGetTimeEntriesForDayAndWeek(t *testing.T) {
	const entriesStartDate = "2017-11-13"
	profileId, accountId := createDefaultUnitTestAccount()
	clientId := createTestClient(accountId, TestClientName, TestClientAddress)
	projectId := createTestProject(accountId, clientId, TestProjectName)
	taskId := createTestTask(accountId)
	createTestTimeEntries(entriesStartDate, 7, accountId, profileId, projectId, taskId)
	defer deleteDefaultUnitTestAccount()
	defer deleteTestClient(clientId)
	defer deleteTestProject(projectId)
	defer deleteTestTask(taskId, accountId)
	defer deleteTestTimeEntries(accountId, profileId, projectId)

	testCases := []struct {
		name            string
		dateString      string
		firstResultDate string
		entryCount      int
		statusCode      int
		errorCode       string
	}{
		{"Valid Sunday 2017-11-19", "2017-11-19", "2017-11-13", 7, http.StatusOK, ""},
		{"Valid Saturday 2017-11-18", "2017-11-18", "2017-11-13", 7, http.StatusOK, ""},
		{"Valid Monday 2017-11-13", "2017-11-13", "2017-11-13", 7, http.StatusOK, ""},
		{"No Entries Sunday 2017-11-12", "2017-11-12", "", 0, http.StatusOK, ""},
		{"No Entries Sunday 2017-11-30", "2017-11-30", "", 0, http.StatusOK, ""},
		{"Invalid date string", "2017-11-32", "", 0, http.StatusBadRequest, api.InvalidField},
		{"Missing date string", "", "", 0, http.StatusNotFound, api.NotFound},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			r, _ := http.NewRequest("GET", "/api/time/week/"+testCase.dateString, nil)
			w := httptest.NewRecorder()
			AddAuthorizationHeaders(r)
			router.ServeHTTP(w, r)

			if w.Code != testCase.statusCode {
				t.Errorf("status code: [%d] wanted: [%d]", w.Code, testCase.statusCode)
			}

			var output jsonResult
			if err := json.NewDecoder(w.Body).Decode(&output); err != nil {
				t.Fatalf("could not decode to json: [%s]", err)
			}

			if testCase.statusCode == http.StatusOK {
				timeResponseJson := timeDataResponse{}
				if err := json.Unmarshal(output.Data, &timeResponseJson); err != nil {
					t.Fatalf("could not decode to json: %s", err)
				}

				if testCase.entryCount != len(timeResponseJson.Entries) {
					t.Fatalf("Wrong result entry count: [%d] wanted: [%d]", testCase.entryCount, len(timeResponseJson.Entries))
				}

				if testCase.entryCount > 0 {
					if timeResponseJson.Entries[0].Day != testCase.firstResultDate {
						t.Errorf("Invalid entry day: [%s] wanted: [%s]", timeResponseJson.Entries[0].Day, testCase.firstResultDate)
					}

					if timeResponseJson.Entries[0].Hours <= 0 {
						t.Errorf("Invalid entry hours. Must be greater than zero: [%f]", timeResponseJson.Entries[0].Hours)
					}

					if timeResponseJson.Entries[0].ClientName != TestClientName {
						t.Errorf("Invalid entry clientName: [%s] wanted: [%s]", timeResponseJson.Entries[0].ClientName, TestClientName)
					}

					if timeResponseJson.Entries[0].ProjectName != TestProjectName {
						t.Errorf("Invalid entry projectName: [%s] wanted: [%s]", timeResponseJson.Entries[0].ProjectName, TestProjectName)
					}

					if timeResponseJson.Entries[0].ProjectId <= 0 || timeResponseJson.Entries[0].TaskId <= 0 {
						t.Errorf("Invalid entry projectId: [%d] or taskId: [%d]", timeResponseJson.Entries[0].ProjectId, timeResponseJson.Entries[0].TaskId)
					}
				}

			} else {
				if output.Code != testCase.errorCode {
					t.Errorf("wrong error code: [%s] wanted: [%s]", output.Code, testCase.errorCode)
				}
			}
		})
	}
}

func TestSaveTimeEntries(t *testing.T) {
	profileId, accountId := createDefaultUnitTestAccount()
	clientId := createTestClient(accountId, TestClientName, TestClientAddress)
	projectId := createTestProject(accountId, clientId, TestProjectName)
	taskId := createTestTask(accountId)
	defer deleteDefaultUnitTestAccount()
	defer deleteTestClient(clientId)
	defer deleteTestProject(projectId)
	defer deleteTestTask(taskId, accountId)

	testCases := []struct {
		name       string
		day1       string
		day2       string
		hours1     float64
		taskId     int
		projectId  int
		entries    int
		statusCode int
		errorCode  string
	}{
		{"2017-11-25 and 2017-11-26", "2017-11-25", "2017-11-26", 7.8, taskId, projectId, 2, http.StatusOK, ""},
		{"Invalid Date", "2017-11-32", "2017-11-26", 7.8, taskId, projectId, 0, http.StatusBadRequest, api.InvalidField},
		{"Missing Date", "", "2017-11-26", 7.8, taskId, projectId, 0, http.StatusBadRequest, api.InvalidField},
		{"Only 1 result (27th on Monday)", "2017-11-26", "2017-11-27", 7.8, taskId, projectId, 1, http.StatusOK, ""},
		{"Hours set to 0", "2017-11-26", "2017-11-27", 0, taskId, projectId, 1, http.StatusOK, ""},
		{"Invalid Task Id", "2017-11-26", "2017-11-27", 0, 0, projectId, 1, http.StatusBadRequest, api.InvalidField},
		{"Invalid Project Id", "2017-11-26", "2017-11-27", 0, taskId, 0, 1, http.StatusBadRequest, api.InvalidField},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			// Clean-up new time entries after they are created
			defer deleteTestTimeEntries(accountId, profileId, projectId)

			entryList := map[string]interface{}{
				"entries": [2]interface{}{
					map[string]interface{}{
						"day":       testCase.day1,
						"hours":     testCase.hours1,
						"taskId":    testCase.taskId,
						"projectId": testCase.projectId,
					},
					map[string]interface{}{
						"day":       testCase.day2,
						"hours":     rand.Float64()*7 + 1.0,
						"taskId":    testCase.taskId,
						"projectId": testCase.projectId,
					},
				},
			}
			body := encodeJson(t, &entryList)

			r, _ := http.NewRequest("PUT", "/api/time", body)
			w := httptest.NewRecorder()
			AddAuthorizationHeaders(r)
			router.ServeHTTP(w, r)

			if w.Code != testCase.statusCode {
				t.Errorf("status code: [%d] wanted: [%d]", w.Code, testCase.statusCode)
			}

			var output jsonResult
			if err := json.NewDecoder(w.Body).Decode(&output); err != nil {
				t.Fatalf("could not decode to json: [%s]", err)
			}

			if testCase.statusCode != http.StatusOK {
				if output.Code != testCase.errorCode {
					t.Errorf("wrong error code: [%s] wanted: [%s]", output.Code, testCase.errorCode)
				}
			} else {
				// Now fetch data and ensure we get the correct values back
				r, _ = http.NewRequest("GET", "/api/time/week/"+testCase.day1, nil)
				w = httptest.NewRecorder()
				AddAuthorizationHeaders(r)
				router.ServeHTTP(w, r)

				if w.Code != testCase.statusCode {
					t.Errorf("status code: [%d] wanted: [%d]", w.Code, testCase.statusCode)
				}

				var weekOutput jsonResult
				if err := json.NewDecoder(w.Body).Decode(&weekOutput); err != nil {
					t.Fatalf("could not decode to json: [%s]", err)
				}

				if testCase.statusCode == http.StatusOK {
					timeResponseJson := timeDataResponse{}
					if err := json.Unmarshal(weekOutput.Data, &timeResponseJson); err != nil {
						t.Fatalf("could not decode to json: %s", err)
					}

					if len(timeResponseJson.Entries) < testCase.entries {
						t.Errorf("Expected [%d] entries got: [%d]", testCase.entries, len(timeResponseJson.Entries))
					} else {
						if timeResponseJson.Entries[0].Day != testCase.day1 {
							t.Errorf("Invalid entry day: [%s] wanted: [%s]", timeResponseJson.Entries[0].Day, testCase.day1)
						}

						if timeResponseJson.Entries[0].Hours != testCase.hours1 {
							t.Errorf("Invalid entry day 1 hours: [%g] wanted: [%g]", timeResponseJson.Entries[0].Hours, testCase.hours1)
						}

						if timeResponseJson.Entries[0].ProjectId != testCase.projectId {
							t.Errorf("Invalid projectId: [%d] wanted: [%d]", timeResponseJson.Entries[0].ProjectId, testCase.projectId)
						}

						if timeResponseJson.Entries[0].TaskId != testCase.taskId {
							t.Errorf("Invalid tasjId: [%d] wanted: [%d]", timeResponseJson.Entries[0].TaskId, testCase.taskId)
						}
					}
				}
			}
		})
	}
}

func TestDeleteProjectTimeEntries(t *testing.T) {
	profileId, accountId := createDefaultUnitTestAccount()
	clientId := createTestClient(accountId, TestClientName, TestClientAddress)
	projectId := createTestProject(accountId, clientId, TestProjectName)
	taskId := createTestTask(accountId)
	const entriesStartDate = "2017-11-20"
	const entriesEndDate = "2017-11-27"
	createTestTimeEntries(entriesStartDate, 7, accountId, profileId, projectId, taskId)
	defer deleteDefaultUnitTestAccount()
	defer deleteTestClient(clientId)
	defer deleteTestProject(projectId)
	defer deleteTestTask(taskId, accountId)
	defer deleteTestTimeEntries(accountId, profileId, projectId)

	testCases := []struct {
		name       string
		day        string
		taskId     int
		projectId  int
		statusCode int
		errorCode  string
	}{
		{"Success for week: 2017-11-20", entriesStartDate, taskId, projectId, http.StatusOK, ""},
		{"Failure for week (already deleted)", entriesStartDate, taskId, projectId, http.StatusBadRequest, api.InvalidField},
		{"Failure for 2017-11-01 (no rows)", "2017-11-01", taskId, projectId, http.StatusBadRequest, api.InvalidField},
		{"Failure missing day", "", taskId, projectId, http.StatusBadRequest, api.InvalidJson},
		{"Invalid Task Id", entriesStartDate, 0, projectId, http.StatusBadRequest, api.InvalidJson},
		{"Invalid Project Id", entriesStartDate, taskId, 0, http.StatusBadRequest, api.InvalidJson},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			deleteProjectJson := map[string]interface{}{
				"startDate":  testCase.day,
				"endDate":  entriesEndDate,
				"projectId":  testCase.projectId,
				"taskId":     testCase.taskId,
			}
			body := encodeJson(t, &deleteProjectJson)

			r, _ := http.NewRequest("DELETE", "/api/time/project/week", body)
			w := httptest.NewRecorder()
			AddAuthorizationHeaders(r)
			router.ServeHTTP(w, r)

			if w.Code != testCase.statusCode {
				t.Errorf("status code: [%d] wanted: [%d]", w.Code, testCase.statusCode)
			}

			var output jsonResult
			if err := json.NewDecoder(w.Body).Decode(&output); err != nil {
				t.Fatalf("could not decode to json: [%s]", err)
			}

			if testCase.statusCode == http.StatusOK {
				if output.Code != testCase.errorCode {
					t.Errorf("wrong error code: [%s] wanted: [%s]", output.Code, testCase.errorCode)
				}
			}
		})
	}
}
