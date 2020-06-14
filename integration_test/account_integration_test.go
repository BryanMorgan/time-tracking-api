// +build integration

package integration_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bryanmorgan/time-tracking-api/api"
	_ "github.com/bryanmorgan/time-tracking-api/config"
	"github.com/bryanmorgan/time-tracking-api/profile"
)

func TestCreateAccount(t *testing.T) {
	defer deleteUnitTestProfileByEmail(TestEmail)

	testCases := []struct {
		name       string
		email      string
		password   string
		firstName  string
		lastName   string
		company    string
		statusCode int
	}{
		{"Successful Account", TestEmail, TestPassword, TestFirstName, TestLastName, TestCompany, http.StatusOK},
		{"Missing Email", "", TestPassword, TestFirstName, TestLastName, TestCompany, http.StatusBadRequest},
		{"Missing Password", TestEmail, "", TestFirstName, TestLastName, TestCompany, http.StatusBadRequest},
		{"Missing First Company", TestEmail, TestPassword, "", TestLastName, TestCompany, http.StatusBadRequest},
		{"Missing Last Company", TestEmail, TestPassword, TestFirstName, "", TestCompany, http.StatusBadRequest},
		{"Missing Company", TestEmail, TestPassword, TestFirstName, TestLastName, "", http.StatusBadRequest},
	}

	type accountResult struct {
		FirstName       string
		Token           string
		TokenExpiration time.Time
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			body := encodeJson(t, &map[string]interface{}{
				"firstName": testCase.firstName,
				"lastName":  testCase.lastName,
				"email":     testCase.email,
				"password":  testCase.password,
				"company":   testCase.company,
			})

			r, _ := http.NewRequest("POST", "/api/account", body)
			r.Header.Add("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, r)

			if w.Code != testCase.statusCode {
				t.Fatalf("Wrong status code: [%d] wanted: [%d]. Body: %s", w.Code, testCase.statusCode, w.Body)
			}

			var output jsonResult
			if err := json.NewDecoder(w.Body).Decode(&output); err != nil {
				t.Fatalf("[%s] Could not decode to json: %s", testCase.name, err.Error())
			}

			if testCase.statusCode == http.StatusOK {
				var accountData accountResult
				if err := json.Unmarshal(output.Data, &accountData); err != nil {
					t.Fatalf("Could not decode accountData data to json: %s. Body: %s", err.Error(), w.Body)
				}

				if accountData.FirstName != TestFirstName {
					t.Fatalf("First name does not match: %s wanted: %s", accountData.FirstName, TestFirstName)
				}
			}
		})
	}
}

// Test if an "non-admin" user tries to do an "admin"-level task (add new user) to ensure it fails
func TestInvalidAdminPermissions(t *testing.T) {
	createUnitTestAccount(TestEmail, TestFirstName, TestLastName, TestCompany, TestCompany2, profile.User)
	defer deleteDefaultUnitTestAccount()

	body := encodeJson(t, &map[string]interface{}{
		"firstName": "Jon",
		"lastName":  "Snow",
		"email":     "jon.snow@unit.test.me",
		"role":      string(profile.Admin),
	})

	testCases := []struct {
		name   string
		method string
		path   string
		body   io.Reader
	}{
		{"Add User", "POST", "/api/account/user", body},
		{"Remove User", "DELETE", "/api/account/user", nil},
		{"Get Account", "GET", "/api/account", nil},
		{"Update Account", "PUT", "/api/account", nil},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {

			r, _ := http.NewRequest(testCase.method, testCase.path, testCase.body)
			w := httptest.NewRecorder()
			AddAuthorizationHeaders(r)
			router.ServeHTTP(w, r)

			if w.Code != http.StatusUnauthorized {
				t.Errorf("Wrong status code: [%d] wanted: [%d]", w.Code, http.StatusUnauthorized)
			}

			var output jsonResult
			if err := json.NewDecoder(w.Body).Decode(&output); err != nil {
				t.Fatalf("Could not decode to json: [%s]", err.Error())
			}

			if output.Code != api.NotAuthorized {
				t.Errorf("Wrong error code: [%s] wanted: [%s]. Result: %s", output.Code, api.NotAuthorized, output)
			}
		})
	}
}

func TestAddUserToAccount(t *testing.T) {
	createDefaultUnitTestAccount()
	defer deleteDefaultUnitTestAccount()

	testCases := []struct {
		name       string
		email      string
		firstName  string
		lastName   string
		role       string
		statusCode int
		errorCode  string
	}{
		{"Successful", "new-" + TestEmail, TestFirstName, TestLastName, string(profile.User), http.StatusOK, ""},
		{"Missing Role", "missing-role-" + TestEmail, TestFirstName, TestLastName, "", http.StatusOK, api.InvalidRole},
		{"Invalid Email", "invalid-email", TestFirstName, TestLastName, string(profile.Admin), http.StatusBadRequest, api.InvalidEmail},
		{"Missing First Name", "missing-first-name-" + TestEmail, "", TestLastName, string(profile.Admin), http.StatusBadRequest, api.FieldSize},
		{"Missing LastName", "missing-last-name-" + TestEmail, "A", "", string(profile.Admin), http.StatusBadRequest, api.FieldSize},
	}

	type addUserResult struct {
		FirstName string
		LastName  string
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {

			// Clean up after adding this new profile
			defer deleteUnitTestProfileByEmail(testCase.email)

			body := encodeJson(t, &map[string]interface{}{
				"firstName": testCase.firstName,
				"lastName":  testCase.lastName,
				"email":     testCase.email,
				"role":      testCase.role,
			})

			r, _ := http.NewRequest("POST", "/api/account/user", body)
			w := httptest.NewRecorder()
			AddAuthorizationHeaders(r)
			router.ServeHTTP(w, r)

			if want, have := testCase.statusCode, w.Code; have != want {
				t.Errorf("Wrong status code: [%d] wanted: [%d]", have, want)
			}

			var output jsonResult
			if err := json.NewDecoder(w.Body).Decode(&output); err != nil {
				t.Fatalf("Could not decode to json: %s", err.Error())
			}

			if testCase.statusCode != http.StatusOK {
				if output.Code != testCase.errorCode {
					t.Errorf("Err code does not match: [%s] wanted: [%s]", output.Code, testCase.errorCode)
				}
			} else {
				var result addUserResult
				if err := json.Unmarshal(output.Data, &result); err != nil {
					t.Fatalf("Could not decode add user json result: %s", err.Error())
				}
				if result.FirstName != TestFirstName {
					t.Errorf("First name does not match: [%s] wanted: [%s]", result.FirstName, TestFirstName)
				}
			}
		})
	}
}

func TestRemoveUserFromAccount(t *testing.T) {
	createDefaultUnitTestAccount()
	defer deleteDefaultUnitTestAccount()

	testCases := []struct {
		name       string
		email      string
		firstName  string
		lastName   string
		role       string
		statusCode int
		errorCode  string
	}{
		{"Successful", "new-" + TestEmail, TestFirstName, TestLastName, string(profile.User), http.StatusOK, ""},
	}

	type addUserResult struct {
		Id int
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {

			// Clean up after adding this new profile
			defer deleteUnitTestProfileByEmail(testCase.email)

			body := encodeJson(t, &map[string]interface{}{
				"firstName": testCase.firstName,
				"lastName":  testCase.lastName,
				"email":     testCase.email,
				"role":      testCase.role,
			})

			// First create a new user account
			r, _ := http.NewRequest("POST", "/api/account/user", body)
			w := httptest.NewRecorder()
			AddAuthorizationHeaders(r)
			router.ServeHTTP(w, r)

			if w.Code != testCase.statusCode {
				t.Errorf("Could not create user to remove. Status code: [%d] wanted: [%d]", w.Code, testCase.statusCode)
			}

			var output jsonResult
			if err := json.NewDecoder(w.Body).Decode(&output); err != nil {
				t.Fatalf("Could not decode to json: %s", err.Error())
			}

			var result addUserResult
			if err := json.Unmarshal(output.Data, &result); err != nil {
				t.Fatalf("Could not decode add user json result: %s", err.Error())
			}

			removeUserJson := map[string]string{
				"email": testCase.email,
			}

			buf := new(bytes.Buffer)
			if err := json.NewEncoder(buf).Encode(&removeUserJson); err != nil {
				t.Fatalf("Could not encode remove user data to JSON: %s", err.Error())
			}

			// Now call DELETE to remove the user
			r, _ = http.NewRequest("DELETE", "/api/account/user", buf)
			AddAuthorizationHeaders(r)
			w = httptest.NewRecorder()

			router.ServeHTTP(w, r)

			if w.Code != testCase.statusCode {
				t.Errorf("Remove user status code: [%d] wanted: [%d]. Body: %s", w.Code, testCase.statusCode, w.Body)
			}

			if err := json.NewDecoder(w.Body).Decode(&output); err != nil {
				t.Fatalf("Could not decode to json: %s", err.Error())
			}

		})
	}
}
