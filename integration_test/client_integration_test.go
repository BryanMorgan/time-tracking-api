// +build integration

package integration_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/bryanmorgan/time-tracking-api/api"
	_ "github.com/bryanmorgan/time-tracking-api/config"
	"github.com/bryanmorgan/time-tracking-api/valid"
)

func TestCreateClient(t *testing.T) {
	createDefaultUnitTestAccount()
	defer deleteDefaultUnitTestAccount()

	testCases := []struct {
		name          string
		clientName    string
		clientAddress string
		statusCode    int
		errorCode     string
	}{
		{"Create Successful", "Simple Company", "123 Main Dallas, TX 75038", http.StatusOK, ""},
		{"Missing Name", "", "123 Main Dallas, TX 75038", http.StatusBadRequest, api.MissingField},
		{"Missing Address (OK)", "Another Co", "", http.StatusOK, ""},
		{"Company Name Too Long", "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890", "", http.StatusBadRequest, api.FieldSize},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			body := encodeJson(t, &map[string]interface{}{
				"name":    testCase.clientName,
				"address": testCase.clientAddress,
			})

			r, _ := http.NewRequest("POST", "/api/client", body)
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
				type clientData struct {
					Id   int
					Name string
				}

				clientDataJson := clientData{}
				if err := json.Unmarshal(output.Data, &clientDataJson); err != nil {
					t.Fatalf("could not decode to json: %s", err)
				}
				defer deleteTestClient(clientDataJson.Id)

				if clientDataJson.Id <= 0 {
					t.Errorf("Invalid client id value: %d", clientDataJson.Id)
				}
			} else {
				if output.Code != testCase.errorCode {
					t.Errorf("wrong error code: [%s] wanted: [%s]", output.Code, testCase.errorCode)
				}
			}
		})
	}
}

func TestUpdateClient(t *testing.T) {
	_, accountId := createDefaultUnitTestAccount()
	clientId := createTestClient(accountId, TestClientName, TestClientAddress)
	defer deleteTestClient(clientId)
	defer deleteDefaultUnitTestAccount()

	testCases := []struct {
		name          string
		clientId      int
		clientName    string
		clientAddress string
		statusCode    int
		errorCode     string
	}{
		{"Create Successful", clientId, "Simple Company", "123 Main Dallas, TX 75038", http.StatusOK, ""},
		{"Create Successful No Address", clientId, "Simple Company", "", http.StatusOK, ""},
		{"Empty Client Name", clientId, "", "123 Main Dallas, TX 75038", http.StatusOK, ""},
		{"Missing/Invalid", 0, "Simple Company", "123 Main Dallas, TX 75038", http.StatusBadRequest, api.MissingField},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			body := encodeJson(t, &map[string]interface{}{
				"id":      testCase.clientId,
				"name":    testCase.clientName,
				"address": testCase.clientAddress,
			})

			r, _ := http.NewRequest("PUT", "/api/client", body)
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
			}
		})
	}
}

func TestArchiveClient(t *testing.T) {
	_, accountId := createDefaultUnitTestAccount()
	clientId := createTestClient(accountId, TestClientName, TestClientAddress)
	defer deleteTestClient(clientId)
	defer deleteDefaultUnitTestAccount()

	testCases := []struct {
		name       string
		clientId   int
		statusCode int
		errorCode  string
	}{
		{"Invalid Client Id", 0, http.StatusBadRequest, api.MissingField},
		{"Missing Client Id", -1, http.StatusBadRequest, api.MissingField},
		{"Archive Success", clientId, http.StatusOK, ""},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			bodyJson := make(map[string]interface{})
			if testCase.clientId >= 0 {
				bodyJson["id"] = testCase.clientId
			}
			body := encodeJson(t, &bodyJson)

			r, _ := http.NewRequest("PUT", "/api/client/archive", body)
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
			}
		})
	}
}

func TestDeleteClient(t *testing.T) {
	_, accountId := createDefaultUnitTestAccount()
	clientId := createTestClient(accountId, TestClientName, TestClientAddress)
	defer deleteTestClient(clientId)
	defer deleteDefaultUnitTestAccount()

	testCases := []struct {
		name       string
		clientId   int
		statusCode int
		errorCode  string
	}{
		{"Invalid Client Id", 0, http.StatusBadRequest, api.MissingField},
		{"Missing Client Id", -1, http.StatusBadRequest, api.MissingField},
		{"Delete Success", clientId, http.StatusOK, ""},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			bodyJson := make(map[string]interface{})
			if testCase.clientId >= 0 {
				bodyJson["id"] = testCase.clientId
			}
			body := encodeJson(t, &bodyJson)

			r, _ := http.NewRequest("DELETE", "/api/client", body)
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
			}
		})
	}
}

func TestGetClient(t *testing.T) {
	_, accountId := createDefaultUnitTestAccount()
	clientId := createTestClient(accountId, TestClientName, TestClientAddress)
	defer deleteDefaultUnitTestAccount()
	defer deleteTestClient(clientId)

	testCases := []struct {
		name       string
		clientId   int
		statusCode int
		errorCode  string
	}{
		{"Valid Client", clientId, http.StatusOK, ""},
		{"Missing/Invalid Client", 0, http.StatusBadRequest, api.InvalidField},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			r, _ := http.NewRequest("GET", "/api/client/"+strconv.Itoa(testCase.clientId), nil)
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
				type clientData struct {
					Id   int
					Name string
				}

				clientDataJson := clientData{}
				if err := json.Unmarshal(output.Data, &clientDataJson); err != nil {
					t.Fatalf("could not decode to json: %s", err)
				}

				if clientDataJson.Id <= 0 {
					t.Errorf("Invalid client id value: %d", clientDataJson.Id)
				}
			} else {
				if output.Code != testCase.errorCode {
					t.Errorf("wrong error code: [%s] wanted: [%s]", output.Code, testCase.errorCode)
				}
			}
		})
	}
}

func TestGetAllClient(t *testing.T) {
	_, accountId := createDefaultUnitTestAccount()
	clientId := createTestClient(accountId, TestClientName, TestClientAddress)
	defer deleteDefaultUnitTestAccount()
	defer deleteTestClient(clientId)

	testCases := []struct {
		name       string
		accountId  int
		statusCode int
		errorCode  string
	}{
		{api.SuccessStatus, accountId, http.StatusOK, ""},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			r, _ := http.NewRequest("GET", "/api/client/all", nil)
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
				type clientData struct {
					Id   int
					Name string
				}
				var clientsData []clientData

				if err := json.Unmarshal(output.Data, &clientsData); err != nil {
					t.Fatalf("could not decode to json: %s", err)
				}

				if len(clientsData) <= 0 {
					t.Errorf("Service should return at least 1 row: %d", len(clientsData))
				}

				if clientsData[0].Id <= 0 {
					t.Errorf("Invalid client id value: %d", clientsData[0].Id)
				}
			} else {
				if output.Code != testCase.errorCode {
					t.Errorf("wrong error code: [%s] wanted: [%s]", output.Code, testCase.errorCode)
				}
			}
		})
	}
}

// --- Project

func TestGetProject(t *testing.T) {
	_, accountId := createDefaultUnitTestAccount()
	clientId := createTestClient(accountId, TestClientName, TestClientAddress)
	projectId := createTestProject(accountId, clientId, TestProjectName)
	defer deleteDefaultUnitTestAccount()
	defer deleteTestClient(clientId)
	defer deleteTestProject(projectId)

	testCases := []struct {
		name       string
		projectId  int
		statusCode int
		errorCode  string
	}{
		{"Valid Project", projectId, http.StatusOK, ""},
		{"Missing/Invalid Project", 0, http.StatusBadRequest, api.InvalidField},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			r, _ := http.NewRequest("GET", "/api/client/project/"+strconv.Itoa(testCase.projectId), nil)
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
				type projectJson struct {
					Id         int
					Name       string
					ClientId   int
					ClientName string
				}

				projectDataJson := projectJson{}
				if err := json.Unmarshal(output.Data, &projectDataJson); err != nil {
					t.Fatalf("could not decode to json: %s", err)
				}

				if projectDataJson.Id <= 0 {
					t.Errorf("Invalid project id value: %d", projectDataJson.Id)
				}

				if projectDataJson.ClientId <= 0 {
					t.Errorf("Invalid client id value: %d", projectDataJson.ClientId)
				}

				if valid.IsNull(projectDataJson.ClientName) {
					t.Errorf("Invalid  name: %s", projectDataJson.ClientName)
				}
			} else {
				if output.Code != testCase.errorCode {
					t.Errorf("wrong error code: [%s] wanted: [%s]", output.Code, testCase.errorCode)
				}
			}
		})
	}
}

func TestGetAllProjects(t *testing.T) {
	_, accountId := createDefaultUnitTestAccount()
	clientId := createTestClient(accountId, TestClientName, TestClientAddress)
	projectId := createTestProject(accountId, clientId, TestProjectName)
	defer deleteDefaultUnitTestAccount()
	defer deleteTestClient(clientId)
	defer deleteTestProject(projectId)

	r, _ := http.NewRequest("GET", "/api/client/project/all", nil)
	w := httptest.NewRecorder()
	AddAuthorizationHeaders(r)
	router.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status code: [%d] wanted: [%d]", w.Code, http.StatusOK)
	}

	var output jsonResult
	if err := json.NewDecoder(w.Body).Decode(&output); err != nil {
		t.Fatalf("could not decode to json: [%s]", err)
	}

	type projectData struct {
		Id         int
		Name       string
		ClientId   int
		ClientName string
	}
	var projectsData []projectData

	if err := json.Unmarshal(output.Data, &projectsData); err != nil {
		t.Fatalf("could not decode to json: %s", err)
	}

	if len(projectsData) <= 0 {
		t.Errorf("Service should return at least 1 row: %d", len(projectsData))
	}

	if projectsData[0].Id <= 0 {
		t.Errorf("Invalid project id value: %d", projectsData[0].Id)
	}

	if projectsData[0].ClientId <= 0 {
		t.Errorf("Invalid client id value: %d", projectsData[0].ClientId)
	}

	if valid.IsNull(projectsData[0].ClientName) {
		t.Errorf("Invalid client name: %s", projectsData[0].ClientName)
	}
}

func TestCreateProject(t *testing.T) {
	_, accountId := createDefaultUnitTestAccount()
	clientId := createTestClient(accountId, TestClientName, TestClientAddress)
	defer deleteDefaultUnitTestAccount()
	defer deleteTestClient(clientId)

	testCases := []struct {
		name        string
		projectName string
		clientId    int
		code        string
		statusCode  int
		errorCode   string
	}{
		{"Create Successful", "Banking and Stuff", clientId, "100-102", http.StatusOK, ""},
		{"Successful Short Name", "A", clientId, "1", http.StatusOK, ""},
		{"Successful No Code", "Banking and Stuff", clientId, "", http.StatusOK, ""},
		{"Missing Name", "", clientId, "100-102", http.StatusBadRequest, api.MissingField},
		{"Invalid Client Id", "A", 0, "100-102", http.StatusBadRequest, api.MissingField},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			body := encodeJson(t, &map[string]interface{}{
				"name":     testCase.projectName,
				"clientId": testCase.clientId,
				"code":     testCase.code,
			})

			r, _ := http.NewRequest("POST", "/api/client/project", body)
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
				type projectData struct {
					Id     int
					Name   string
					Active bool
				}

				projectDataJson := projectData{}
				if err := json.Unmarshal(output.Data, &projectDataJson); err != nil {
					t.Fatalf("could not decode to json: %s", err)
				}
				defer deleteTestProject(projectDataJson.Id)

				if projectDataJson.Id <= 0 {
					t.Errorf("Invalid client id value: %d", projectDataJson.Id)
				}

				if projectDataJson.Active != true {
					t.Errorf("Invalid active value: %v", projectDataJson.Active)
				}
			} else {
				if output.Code != testCase.errorCode {
					t.Errorf("wrong error code: [%s] wanted: [%s]", output.Code, testCase.errorCode)
				}
			}
		})
	}
}

func TestUpdateProject(t *testing.T) {
	_, accountId := createDefaultUnitTestAccount()
	clientId := createTestClient(accountId, TestClientName, TestClientAddress)
	projectId := createTestProject(accountId, clientId, TestProjectName)
	defer deleteDefaultUnitTestAccount()
	defer deleteTestClient(clientId)
	defer deleteTestProject(projectId)

	testCases := []struct {
		name        string
		projectId   int
		projectName string
		code        string
		active      bool
		statusCode  int
		errorCode   string
	}{
		{"Create Successful", projectId, "Design UI for Web", "123-909-123", true, http.StatusOK, ""},
		{"Invalid Project Name", projectId, "", "", false, http.StatusBadRequest, api.FieldSize},
		{"Missing/Invalid Project Id", 0, "Simple Company", "123 Main Dallas, TX 75038", false, http.StatusBadRequest, api.MissingField},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			body := encodeJson(t, &map[string]interface{}{
				"id":       testCase.projectId,
				"name":     testCase.projectName,
				"code":     testCase.code,
				"active":   testCase.active,
				"clientId": clientId,
			})

			r, _ := http.NewRequest("PUT", "/api/client/project", body)
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
			}
		})
	}
}

func TestDeleteProject(t *testing.T) {
	_, accountId := createDefaultUnitTestAccount()
	clientId := createTestClient(accountId, TestClientName, TestClientAddress)
	projectId := createTestProject(accountId, clientId, TestProjectName)
	defer deleteDefaultUnitTestAccount()
	defer deleteTestClient(clientId)
	defer deleteTestProject(projectId)

	testCases := []struct {
		name       string
		projectId  int
		statusCode int
		errorCode  string
	}{
		{"Invalid project id", 0, http.StatusBadRequest, api.MissingField},
		{"Missing project id", -1, http.StatusBadRequest, api.MissingField},
		{"Delete Successful", projectId, http.StatusOK, ""},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			bodyJson := make(map[string]interface{})
			if testCase.projectId >= 0 {
				bodyJson["projectId"] = testCase.projectId
			}
			body := encodeJson(t, &bodyJson)

			r, _ := http.NewRequest("DELETE", "/api/client/project", body)
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
			}
		})
	}
}

func TestUpdateProjectActive(t *testing.T) {
	_, accountId := createDefaultUnitTestAccount()
	clientId := createTestClient(accountId, TestClientName, TestClientAddress)
	projectId := createTestProject(accountId, clientId, TestProjectName)
	defer deleteDefaultUnitTestAccount()
	defer deleteTestClient(clientId)
	defer deleteTestProject(projectId)

	testCases := []struct {
		name       string
		projectId  int
		statusCode int
		errorCode  string
	}{
		{"Invalid project id", 0, http.StatusBadRequest, api.MissingField},
		{"Missing project id", -1, http.StatusBadRequest, api.MissingField},
		{"Archive Successful", projectId, http.StatusOK, ""},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			bodyJson := make(map[string]interface{})
			if testCase.projectId >= 0 {
				bodyJson["projectId"] = testCase.projectId
			}
			body := encodeJson(t, &bodyJson)

			r, _ := http.NewRequest("PUT", "/api/client/project/archive", body)
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
			}
		})
	}
}
