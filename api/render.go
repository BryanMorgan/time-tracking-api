package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap/zapcore"

	"github.com/bryanmorgan/time-tracking-api/logger"
)

const (
	SuccessStatus = "success"
	ErrorStatus   = "error"
)

type successResponse struct {
	Data interface{} `json:"data,omitempty"`
}

// Writes a successful JSON response and marshals the value type as the data element
func Json(w http.ResponseWriter, r *http.Request, data interface{}) {
	JsonWithStatus(w, r, data, http.StatusOK)
}

// Writes a JSON response with the given header status code and marshals the value type as the data element
func JsonWithStatus(w http.ResponseWriter, r *http.Request, data interface{}, statusCode int) {
	e := json.NewEncoder(w)

	if viper.GetBool("logging.pretty") {
		e.SetIndent("", "    ")
	}

	success := successResponse{
		Data: data,
	}

	setJsonHeaders(w)
	w.WriteHeader(statusCode)

	if err := e.Encode(&success); err != nil {
		ErrorJson(w, NewError(err, "JSON encoding failure", JsonEncodeFailed), http.StatusInternalServerError)
		return
	}
}

func BadInputs(w http.ResponseWriter, message string, code string, field string) {
	ErrorJson(w, NewError(errors.New("invalid input"), message, code, NewErrorDetail("field", field)), http.StatusBadRequest)
}

// Writes an error JSON response
func ErrorJson(w http.ResponseWriter, e *Error, httpStatusCode int) {
	writeErrorJson(w, e, httpStatusCode)
	logger.Log.Error(e.Message, createFields(e)...)
}

// Writes an error JSON response
func WarnJson(w http.ResponseWriter, e *Error, httpStatusCode int) {
	writeErrorJson(w, e, httpStatusCode)
	logger.Log.Warn(e.Message, createFields(e)...)
}

func CloseBody(body io.ReadCloser) {
	if err := body.Close(); err != nil {
		logger.Log.Error("Failed to close body: " + err.Error())
	}
}

func createFields(appErr *Error) []zapcore.Field {
	var fields []zapcore.Field
	fields = append(fields, logger.Error(appErr.Err))
	fields = append(fields, logger.String("code", appErr.Code))
	if appErr.Detail != nil {
		for key, value := range appErr.Detail {
			fields = append(fields, logger.Any(key, value))
		}
	}
	return fields
}

func writeErrorJson(w http.ResponseWriter, e *Error, httpStatusCode int) {
	encoder := json.NewEncoder(w)

	if viper.GetBool("logging.pretty") {
		encoder.SetIndent("", "    ")
	}

	setJsonHeaders(w)
	w.WriteHeader(httpStatusCode)

	if err := encoder.Encode(e); err != nil {
		logger.Log.Error("Failed to encode error JSON", logger.Error(err))
		if err = json.NewEncoder(w).Encode([]byte("{\"error\": \"" + e.Message + "\"}")); err != nil {
			logger.Log.Error("Failed to encode static error", logger.Error(err))
		}

		return
	}
}

// Set JSON header response values with a focus on security
func setJsonHeaders(w http.ResponseWriter) {
	header := w.Header()
	header.Set("Content-Type", "application/json")

	// Add security headers
	header.Add("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
	header.Add("X-Frame-Options", "DENY")
	header.Add("X-Content-Type-Options", "nosniff")
	header.Add("X-XSS-Protection", "1; mode=block")

	// TODO: Review content-security-policy
	header.Add("Content-Security-Policy", "default-src 'self' ; img-src 'self'; script-src 'unsafe-inline' ; style-src 'self' ; report-uri /_csp_report; ")

	// Add expires and cache control headers
	header.Add("Expires", "-1")
	header.Add("Cache-Control", "private, max-age=0")
}

// Standard Time JSON format
func TimeJson(t time.Time) string {
	return t.Format(time.RFC3339)
}
