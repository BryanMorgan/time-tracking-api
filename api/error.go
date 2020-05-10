package api

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Error Codes
const (
	PANIC            = "Panic"
	NotFound         = "NotFound"
	MethodNotAllowed = "MethodNotAllowed"

	InvalidJson        = "InvalidJson"
	InvalidContentType = "InvalidContentType"
	JsonEncodeFailed   = "JsonEncodeFailed"
	SystemError        = "SystemError"
	PaymentError       = "PaymentError"

	EncryptionFailed    = "EncryptionFailed"
	TokenCreationFailed = "TokenCreationFailed"

	AccountExists        = "AccountExists"
	SubscriptionExists   = "SubscriptionExists"
	UpdateFailed         = "UpdateFailed"
	AccountCreateFailed  = "AccountCreateFailed"
	ProfileCreateFailed  = "ProfileCreateFailed"
	EmailExistsInAccount = "EmailExistsInAccount"
	ProfileNotFound      = "ProfileNotFound"
	InvalidEmail         = "InvalidEmail"
	InvalidForgotToken   = "InvalidForgotToken"
	InvalidAccountId     = "InvalidAccountId"
	InvalidPassword      = "InvalidPassword"
	PasswordMismatch     = "PasswordMismatch"
	InvalidField         = "InvalidField"
	ProfileLocked        = "ProfileLocked"
	NotAuthorized        = "NotAuthorized"

	IncorrectPassword        = "IncorrectPassword"
	InvalidToken             = "InvalidToken"
	TokenExpired             = "TokenExpired"
	MissingToken             = "MissingToken"
	InvalidAccountForProfile = "InvalidAccountForProfile"
	AccountSwitchFailed      = "AccountSwitchFailed"

	ProfileInactive = "ProfileInactive"
	AccountInactive = "AccountInactive"

	MissingField = "MissingField"
	FieldSize    = "FieldSize"

	InvalidWeekStart = "InvalidWeekStart"
	InvalidRole      = "InvalidRole"
	InvalidTimezone  = "InvalidTimezone"

	InvalidClient  = "InvalidClient"
	InvalidTask    = "InvalidTask"
	InvalidProject = "InvalidProject"
)

type Error struct {
	Err     error
	Message string
	Code    string
	Detail  map[string]interface{}
}

type ErrorDetail struct {
	Key   string
	Value interface{}
}

func NewFieldError(errorDetail interface{}, message string, code string, field string) *Error {
	return NewError(errorDetail, message, code, ErrorDetail{"field", field})
}

func NewErrorDetail(key string, value interface{}) ErrorDetail {
	return ErrorDetail{
		Key:   key,
		Value: value,
	}
}

// Standardize the creation and reporting of errors using a generic Err type
func NewError(errorDetail interface{}, message string, code string, details ...ErrorDetail) *Error {
	var err error
	if errorDetail != nil {
		switch errorDetail.(type) {
		case error:
			err = errorDetail.(error)
		case string:
			err = errors.New(errorDetail.(string))
		default:
			err = fmt.Errorf("%v", errorDetail)
		}
	}

	errorMessage := Error{
		Err:     err,
		Message: message,
		Code:    code}

	if details != nil {
		errorMessage.Detail = make(map[string]interface{})

		for _, detail := range details {
			if detail.Value != "" {
				errorMessage.Detail[detail.Key] = detail.Value
			}
		}
	}

	return &errorMessage
}

func (e *Error) MarshalJSON() ([]byte, error) {
	var errorString string
	if e.Err != nil {
		errorString = e.Err.Error()
	}

	return json.Marshal(&struct {
		Status  string                 `json:"status,omitempty"`
		Error   string                 `json:"error,omitempty"`
		Message string                 `json:"message,omitempty"`
		Code    string                 `json:"code,omitempty"`
		Detail  map[string]interface{} `json:"detail,omitempty"`
	}{
		Status:  ErrorStatus,
		Error:   errorString,
		Message: e.Message,
		Code:    e.Code,
		Detail:  e.Detail,
	})
}

func (e *Error) String() string {
	return fmt.Sprintf("[%s] %s: %s", e.Code, e.Message, e.Err)
}

func (e *Error) Error() string {
	return e.String()
}
