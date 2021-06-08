// +build integration

package integration_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bryanmorgan/time-tracking-api/api"
	_ "github.com/bryanmorgan/time-tracking-api/config"
)

func TestLogin(t *testing.T) {
	createDefaultUnitTestAccount()
	defer deleteDefaultUnitTestAccount()

	cases := []struct {
		name       string
		email      string
		password   string
		statusCode int
	}{
		{"Valid Test User Mixed Case", TestEmail, TestPassword, http.StatusOK},
		{"Valid Test User Lower Case", strings.ToLower(TestEmail), TestPassword, http.StatusOK},
		{"Incorrect Email", TestEmail + "-invalid", TestPassword, http.StatusUnauthorized},
		{"Incorrect Password", TestEmail, TestPassword + "-invalid", http.StatusUnauthorized},
		{"Empty Email", "", TestPassword, http.StatusBadRequest},
		{"Empty Password", TestEmail, "", http.StatusBadRequest},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			body := encodeJson(t, &map[string]interface{}{
				"email":    testCase.email,
				"password": testCase.password,
			})

			r, _ := http.NewRequest("POST", "/api/auth/login", body)
			w := httptest.NewRecorder()
			AddRequestHeaders(r)
			router.ServeHTTP(w, r)

			t.Logf("%s: %s", testCase.name, w.Body)

			if w.Code != testCase.statusCode {
				t.Fatalf("Invalid status code: [%d] wanted: [%d]", w.Code, testCase.statusCode)
			}
		})
	}
}

func TestProfileUpdate(t *testing.T) {
	createDefaultUnitTestAccount()
	defer deleteDefaultUnitTestAccount()

	cases := []struct {
		name       string
		firstName  string
		lastName   string
		email      string
		statusCode int
	}{
		{"Valid Update", "New First Name", "Updated Last Name", TestEmail, http.StatusOK},
		{"Mixed Case Email", TestFirstName, TestLastName, "SomeThing@someWHERE.com", http.StatusOK},
		{"Email Too Short", TestFirstName, TestLastName, "a@a.", http.StatusBadRequest},
		{"Email Missing OK", TestFirstName, TestLastName, "", http.StatusOK},
		{"Short First Company", "X", TestLastName, TestEmail, http.StatusOK},
		{"Short Last Company", TestFirstName, "X", TestEmail, http.StatusOK},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			body := encodeJson(t, &map[string]interface{}{
				"email":     testCase.email,
				"firstName": testCase.firstName,
				"lastName":  testCase.lastName,
			})

			r, _ := http.NewRequest("PUT", "/api/profile", body)
			w := httptest.NewRecorder()
			AddAuthorizationHeaders(r)
			router.ServeHTTP(w, r)

			if w.Code != testCase.statusCode {
				t.Fatalf("Invalid status code: [%d] wanted: [%d]", w.Code, testCase.statusCode)
			}

			var output jsonResult
			if err := json.NewDecoder(w.Body).Decode(&output); err != nil {
				t.Fatalf("[%s] Could not decode to json: %s", testCase.name, err.Error())
			}
		})
	}
}

func TestPasswordUpdate(t *testing.T) {
	createDefaultUnitTestAccount()
	defer deleteDefaultUnitTestAccount()

	newPassword := TestPassword + "updated"

	cases := []struct {
		name            string
		currentPassword string
		password        string
		confirmPassword string
		statusCode      int
		errorCode       string
	}{
		{"Valid Update", TestPassword, newPassword, newPassword, http.StatusOK, ""},
		{"Invalid Current Password", "123456789", newPassword, newPassword, http.StatusBadRequest, api.InvalidPassword},
		{"Invalid Password Size", TestPassword, "123", "123", http.StatusBadRequest, api.FieldSize},
		{"Invalid Confirm Password", TestPassword, newPassword, "12345678", http.StatusBadRequest, api.PasswordMismatch},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			body := encodeJson(t, &map[string]interface{}{
				"currentPassword": testCase.currentPassword,
				"password":        testCase.password,
				"confirmPassword": testCase.confirmPassword,
			})

			r, _ := http.NewRequest("PUT", "/api/profile/password", body)
			w := httptest.NewRecorder()
			AddAuthorizationHeaders(r)
			router.ServeHTTP(w, r)

			if w.Code != testCase.statusCode {
				t.Fatalf("Invalid status code: [%d] wanted: [%d]", w.Code, testCase.statusCode)
			}

			var output jsonResult
			if err := json.NewDecoder(w.Body).Decode(&output); err != nil {
				t.Fatalf("[%s] Could not decode to json: %s", testCase.name, err.Error())
			}

			if output.Code != testCase.errorCode {
				t.Fatalf("Invalid error code: '%s' wanted: '%s'", output.Code, testCase.errorCode)
			}
		})
	}
}

func TestLogout(t *testing.T) {
	createDefaultUnitTestAccount()
	defer deleteDefaultUnitTestAccount()

	// Make sure logout worked
	r, _ := http.NewRequest("POST", "/api/auth/logout", nil)
	AddAuthorizationHeaders(r)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("Invalid status code: [%d] wanted: [%d]", w.Code, http.StatusOK)
	}

	var output jsonResult
	if err := json.NewDecoder(w.Body).Decode(&output); err != nil {
		t.Fatalf("Could not decode to json: %s", err.Error())
	}

	// Verify the token is deleted from the session table by calling the token API
	w = httptest.NewRecorder()
	r, _ = http.NewRequest("POST", "/api/auth/token", nil)
	AddAuthorizationHeaders(r)
	router.ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("Invalid status code: expected %d. Got: %d. Session should be deleted and token API should fail.",
			http.StatusUnauthorized, w.Code)
	}
}

func TestTokens(t *testing.T) {
	createDefaultUnitTestAccount()
	defer deleteDefaultUnitTestAccount()

	testCases := []struct {
		name        string
		headerName  string
		headerValue string
		statusCode  int
	}{
		{"Valid Token", "Authorization", "Bearer " + TestToken, http.StatusOK},
		{"Invalid or Expired Token", "Authorization", "Bearer 1234567890", http.StatusUnauthorized},
		{"Missing Token Value", "Authorization", "Bearer ", http.StatusUnauthorized},
		{"Invalid Token Type", "Authorization", "Auth 1234567890", http.StatusUnauthorized},
		{"Invalid Header", "", "", http.StatusUnauthorized},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			r, _ := http.NewRequest("POST", "/api/auth/token", nil)
			AddRequestHeaders(r)
			r.Header.Add(testCase.headerName, testCase.headerValue)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, r)

			if w.Code != testCase.statusCode {
				t.Fatalf("[%s] invalid status code. Expected %d. Got: %d", testCase.name, testCase.statusCode, w.Code)
			}

			var output jsonResult
			if err := json.NewDecoder(w.Body).Decode(&output); err != nil {
				t.Fatalf("[%s] Could not decode to json: %s", testCase.name, err.Error())
			}
		})
	}
}

func TestLoginLockAccount(t *testing.T) {
	createDefaultUnitTestAccount()
	defer deleteDefaultUnitTestAccount()

	// Login with an invalid password 6 times
	for i := 0; i < 6; i++ {
		body := encodeJson(t, &map[string]interface{}{
			"email":    TestEmail,
			"password": "invalid!",
		})

		r, _ := http.NewRequest("POST", "/api/auth/login", body)
		w := httptest.NewRecorder()
		AddRequestHeaders(r)
		router.ServeHTTP(w, r)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("invalid status code in loop [%d]: [%d] wanted: [%d]", i, w.Code, http.StatusUnauthorized)
		}

		var output jsonResult
		if err := json.NewDecoder(w.Body).Decode(&output); err != nil {
			t.Fatalf("could not decode to json: %s", err.Error())
		}

		if output.Code != api.IncorrectPassword {
			t.Fatalf("invalid error code in loop [%d]: [%s] wanted: [%s]", i, output.Code, api.IncorrectPassword)
		}
	}

	// Account should be locked, try 3 more times with a valid password and check lock status
	for i := 0; i < 3; i++ {
		body := encodeJson(t, &map[string]interface{}{
			"email":    TestEmail,
			"password": TestPassword,
		})

		r, _ := http.NewRequest("POST", "/api/auth/login", body)
		w := httptest.NewRecorder()
		AddRequestHeaders(r)
		router.ServeHTTP(w, r)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("Account not locked in loop [%d]: status code: [%d] wanted: [%d]", i, w.Code, http.StatusUnauthorized)
		}

		var output jsonResult
		if err := json.NewDecoder(w.Body).Decode(&output); err != nil {
			t.Fatalf("could not decode to json: %s", err.Error())
		}

		if output.Code != api.ProfileLocked {
			t.Fatalf("account not locked in loop [%d]: wrong error code: [%s] wanted: [%s]", i, output.Code, api.ProfileLocked)
		}
	}
}

func TestForgotPassword(t *testing.T) {
	createDefaultUnitTestAccount()
	defer deleteDefaultUnitTestAccount()

	cases := []struct {
		name       string
		email      string
		statusCode int
		errorCode  string
	}{
		{"Valid Email", TestEmail, http.StatusOK, ""},
		{"Invalid Email", "a@", http.StatusBadRequest, api.InvalidEmail},
		{"Missing Email", "", http.StatusBadRequest, api.InvalidEmail},
		{"Not Found Email", "noone@acme.test.me", http.StatusBadRequest, api.ProfileNotFound},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			body := encodeJson(t, &map[string]interface{}{
				"email": testCase.email,
			})

			r, _ := http.NewRequest("POST", "/api/auth/forgot", body)
			w := httptest.NewRecorder()
			AddRequestHeaders(r)
			router.ServeHTTP(w, r)

			//t.Logf("%s: %s", testCase.name, loginResponse.Body)
			if w.Code != testCase.statusCode {
				t.Fatalf("Invalid status code: [%d] wanted: [%d]", w.Code, testCase.statusCode)
			}

			var output jsonResult
			if err := json.NewDecoder(w.Body).Decode(&output); err != nil {
				t.Fatalf("could not decode to json: %s", err)
			}

			if output.Code != testCase.errorCode {
				t.Errorf("wrong error code: [%s] wanted: [%s]", output.Code, testCase.errorCode)
			}
		})
	}
}

func TestForgotPasswordTokenValidation(t *testing.T) {
	createDefaultUnitTestAccount()
	defer deleteDefaultUnitTestAccount()

	cases := []struct {
		name       string
		token      string
		statusCode int
		errorCode  string
	}{
		{"Valid Token", TestForgotPasswordToken, http.StatusOK, ""},
		{"Invalid Token", "123", http.StatusBadRequest, api.InvalidForgotToken},
		{"Missing Token", "", http.StatusBadRequest, api.InvalidForgotToken},
		{"Not Found Token", "123321", http.StatusBadRequest, api.InvalidForgotToken},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			body := encodeJson(t, &map[string]interface{}{
				"forgotPasswordToken": testCase.token,
			})

			r, _ := http.NewRequest("POST", "/api/auth/forgot/validate", body)
			w := httptest.NewRecorder()
			AddRequestHeaders(r)
			router.ServeHTTP(w, r)

			if w.Code != testCase.statusCode {
				t.Fatalf("Invalid status code: [%d] wanted: [%d]", w.Code, testCase.statusCode)
			}

			var output jsonResult
			if err := json.NewDecoder(w.Body).Decode(&output); err != nil {
				t.Fatalf("could not decode to json: %s", err)
			}

			if output.Code != testCase.errorCode {
				t.Errorf("wrong error code: [%s] wanted: [%s]", output.Code, testCase.errorCode)
			}
		})
	}
}
