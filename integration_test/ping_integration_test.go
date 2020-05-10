// +build integration

package integration_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPing(t *testing.T) {
	testCases := []struct {
		name       string
		method     string
		statusCode int
	}{
		{"Ping successful", "GET", http.StatusOK},
		{"Invalid method", "POST", http.StatusMethodNotAllowed},
		{"Invalid method", "PUT", http.StatusMethodNotAllowed},
		{"Invalid method", "HEAD", http.StatusMethodNotAllowed},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			r, _ := http.NewRequest(testCase.method, "/_ping", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, r)

			if w.Code != testCase.statusCode {
				t.Errorf("status code: [%d] wanted: [%d]", w.Code, testCase.statusCode)
			}
		})
	}
}
