package profile

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bryanmorgan/time-tracking-api/api"
	"github.com/bryanmorgan/time-tracking-api/config"
	"github.com/bryanmorgan/time-tracking-api/logger"
	"github.com/bryanmorgan/time-tracking-api/valid"

	"github.com/spf13/viper"
)

const (
	AuthorizationHeaderName = "Authorization"
	BearerAuthorizationType = "Bearer"
	TokenQueryParameterName = "token"
)

type ProfileRequest struct {
	FirstName string
	LastName  string
	Email     string
	Password  string
	Timezone  string
}

type PasswordChangeRequest struct {
	CurrentPassword string
	Password        string
	ConfirmPassword string
}

type ProfileResponse struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Company   string `json:"company"`
	Email     string `json:"email"`
	Phone     string `json:"phone,omitempty"`
	Timezone  string `json:"timezone,omitempty"`
}

type CompanyResponse struct {
	Id      int    `json:"id"`
	Company string `json:"company"`
}

type AccountResponse struct {
	Company   string `json:"company"`
	WeekStart int    `json:"weekStart"`
	Timezone  string `json:"timezone,omitempty"`
	Created   string `json:"created,omitempty"`
	Updated   string `json:"updated,omitempty"`
}

type EmailRequest struct {
	Email string
}

type LoginRequest struct {
	EmailRequest
	Password string
}

type ForgotPasswordTokenRequest struct {
	ForgotPasswordToken string
}

type SetupNewUserRequest struct {
	Token    string
	Password string
}

type AuthResponse struct {
	Id        int    `json:"id,omitempty"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Company   string `json:"company"`
	WeekStart int    `json:"weekStart"`
}

type AccountIdRequest struct {
	AccountId int
}

func (pr *ProfileRouter) loginHandler(w http.ResponseWriter, r *http.Request) {
	var loginRequest LoginRequest

	if r.Body == nil {
		api.ErrorJson(w, api.NewError(nil, "Empty Body", api.InvalidJson), http.StatusBadRequest)
		return
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&loginRequest); err != nil {
		api.ErrorJson(w, api.NewError(err, "Invalid JSON", api.InvalidJson), http.StatusBadRequest)
		return
	}
	defer api.CloseBody(r.Body)

	// Validate fields
	if !valid.IsEmail(loginRequest.Email) {
		api.BadInputs(w, "Invalid email", api.InvalidEmail, "email")
		return
	}

	if !valid.IsLength(loginRequest.Email, EmailMinLength, EmailMaxLength) {
		api.BadInputs(w, "Invalid email length", api.FieldSize, "email")
		return
	}

	if !valid.IsLength(loginRequest.Password, PasswordMinLength, PasswordMaxLength) {
		api.BadInputs(w, "Invalid password length", api.FieldSize, "password")
		return
	}

	userProfile, appErr := pr.profileService.Login(loginRequest.Email, loginRequest.Password, r.RemoteAddr)
	if appErr != nil {
		api.ErrorJson(w, appErr, http.StatusUnauthorized)
		return
	}

	AddSessionCookie(w, userProfile.Token.String, viper.GetString("application.applicationDomain"))
	api.Json(w, r, NewAuthResponse(userProfile))
}

func (pr *ProfileRouter) logoutHandler(w http.ResponseWriter, r *http.Request) {
	token, ok := r.Context().Value(config.TokenContextKey).(string)
	if !ok {
		api.ErrorJson(w, api.NewError(nil, "Invalid token context", api.SystemError), http.StatusUnauthorized)
		return
	}

	if appErr := pr.profileService.Logout(token); appErr != nil {
		api.ErrorJson(w, appErr, http.StatusBadRequest)
		return
	}

	ClearSessionCookie(w, viper.GetString("application.applicationDomain"))
	api.Json(w, r, nil)
}

func (pr *ProfileRouter) forgotPasswordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		api.ErrorJson(w, api.NewError(nil, "Empty Body", api.InvalidJson), http.StatusBadRequest)
		return
	}

	var emailRequest EmailRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&emailRequest); err != nil {
		api.ErrorJson(w, api.NewError(err, "Invalid JSON", api.InvalidJson), http.StatusBadRequest)
		return
	}
	defer api.CloseBody(r.Body)

	logger.Log.Info("Forgot password request", logger.String("email", emailRequest.Email))

	// Validate fields
	if !valid.IsEmail(emailRequest.Email) {
		api.BadInputs(w, "Invalid email", api.InvalidEmail, "email")
		return
	}

	if appErr := pr.profileService.ForgotPassword(emailRequest.Email); appErr != nil {
		api.ErrorJson(w, appErr, http.StatusBadRequest)
		return
	}

	api.Json(w, r, nil)
}

func (pr *ProfileRouter) validateForgotTokenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		api.ErrorJson(w, api.NewError(nil, "Empty Body", api.InvalidJson), http.StatusBadRequest)
		return
	}

	var tokenRequest ForgotPasswordTokenRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&tokenRequest); err != nil {
		api.ErrorJson(w, api.NewError(err, "Invalid JSON", api.InvalidJson), http.StatusBadRequest)
		return
	}
	defer api.CloseBody(r.Body)

	// Validate fields
	if valid.IsNull(tokenRequest.ForgotPasswordToken) {
		api.BadInputs(w, "Invalid forgot password token", api.InvalidForgotToken, "forgotPasswordToken")
		return
	}

	if appErr := pr.profileService.ValidateForgotPasswordToken(tokenRequest.ForgotPasswordToken); appErr != nil {
		api.ErrorJson(w, appErr, http.StatusBadRequest)
		return
	}

	api.Json(w, r, nil)
}

func (pr *ProfileRouter) validateTokenHandler(w http.ResponseWriter, r *http.Request) {
	token, ok := r.Context().Value(config.TokenContextKey).(string)
	if !ok {
		api.ErrorJson(w, api.NewError(nil, "Invalid token context", api.SystemError), http.StatusUnauthorized)
		return
	}

	authProfile, err := pr.profileService.GetProfile(token)
	if err != nil {
		api.ErrorJson(w, api.NewError(nil, "Failed to get profile", api.InvalidToken), http.StatusUnauthorized)
		return
	}

	if authProfile == nil {
		api.ErrorJson(w, api.NewError(nil, "No profile for token", api.InvalidToken), http.StatusUnauthorized)
		return
	}

	api.Json(w, r, NewAuthResponse(authProfile))
}

func (pr *ProfileRouter) setupNewUserAccountHandler(w http.ResponseWriter, r *http.Request) {
	var setupNewUserRequest SetupNewUserRequest

	if r.Body == nil {
		api.ErrorJson(w, api.NewError(nil, "Empty Body", api.InvalidJson), http.StatusBadRequest)
		return
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&setupNewUserRequest); err != nil {
		api.ErrorJson(w, api.NewError(err, "Invalid JSON", api.InvalidJson), http.StatusBadRequest)
		return
	}
	defer api.CloseBody(r.Body)

	// Validate fields
	if valid.IsNull(setupNewUserRequest.Token) {
		api.BadInputs(w, "Missing token", api.InvalidToken, "token")
		return
	}

	tokenLength := viper.GetInt("session.tokenLength")
	if valid.IsLength(setupNewUserRequest.Token, 0, tokenLength) {
		api.BadInputs(w, "Invalid token length", api.InvalidToken, "token")
		return
	}

	if !valid.IsLength(setupNewUserRequest.Password, PasswordMinLength, PasswordMaxLength) {
		api.BadInputs(w, "Invalid password length", api.FieldSize, "password")
		return
	}

	appErr := pr.profileService.SetupNewUser(setupNewUserRequest.Token, setupNewUserRequest.Password)
	if appErr != nil {
		api.ErrorJson(w, appErr, http.StatusUnauthorized)
		return
	}

	api.Json(w, r, nil)
}

func TokenHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := GetTokenFromRequest(r)
		if err != nil {
			api.ErrorJson(w, err, http.StatusUnauthorized)
			return
		}

		// Add/UpdateAccount the session cookie
		if !strings.Contains(r.URL.Path, "/api/auth/logout") {
			AddSessionCookie(w, token, viper.GetString("application.applicationDomain"))
		}

		ctx := context.WithValue(r.Context(), config.TokenContextKey, token)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func NewAuthResponse(values *Profile) *AuthResponse {
	return &AuthResponse{
		FirstName: values.FirstName,
		LastName:  values.LastName,
		Company:   values.Company,
		Id:        values.ProfileId,
		WeekStart: values.WeekStart,
	}
}

func GetTokenFromRequest(r *http.Request) (string, *api.Error) {
	var token string

	// First check for an Authorization: Bearer {token} header
	authHeader := r.Header.Get(AuthorizationHeaderName)
	if authHeader != "" {
		bearerToken := strings.SplitN(authHeader, " ", 2)
		if len(bearerToken) != 2 {
			return "", api.NewError(nil, "Missing authorization token", api.InvalidToken)
		}

		if bearerToken[0] != BearerAuthorizationType {
			return "", api.NewError(nil, "Token not Bearer type: "+bearerToken[0], api.InvalidToken)
		}

		token = bearerToken[1]
	}

	// Next check the query parameter
	if valid.IsNull(token) {
		token = r.URL.Query().Get(TokenQueryParameterName)
	}

	// If not found, check for a session cookie
	if valid.IsNull(token) {
		cookie := GetSessionCookieFromRequest(r)
		if cookie != nil {
			token = cookie.Value
		}
	}

	if valid.IsNull(token) {
		return "", api.NewError(nil, "Missing token", api.MissingToken)
	}

	return token, nil
}

// -----------------------------------------

func (pr *ProfileRouter) updateProfileHandler(w http.ResponseWriter, r *http.Request) {
	var profileRequest ProfileRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&profileRequest); err != nil {
		api.ErrorJson(w, api.NewError(err, "Invalid JSON", api.InvalidJson), http.StatusBadRequest)
		return
	}
	defer api.CloseBody(r.Body)

	existingUserProfile, ok := r.Context().Value(config.ProfileContextKey).(*Profile)
	if !ok || existingUserProfile == nil {
		api.ErrorJson(w, api.NewError(nil, "Invalid profile context", api.SystemError), http.StatusInternalServerError)
		return
	}

	// Copy this profile to update new values
	updatedProfile := *existingUserProfile

	if !valid.IsNull(profileRequest.Email) {
		emailLowerCase := strings.ToLower(profileRequest.Email)
		if !valid.IsEmail(emailLowerCase) {
			api.BadInputs(w, "Invalid email", api.InvalidEmail, "email")
			return
		}
		updatedProfile.Email = emailLowerCase
	}

	if !valid.IsNull(profileRequest.FirstName) {
		if !valid.IsLength(profileRequest.FirstName, NameMinLength, NameMaxLength) {
			api.BadInputs(w, "First name must be between 1 and 64 characters", api.FieldSize, "firstName")
			return
		}
		updatedProfile.FirstName = profileRequest.FirstName
	}

	if !valid.IsNull(profileRequest.LastName) {
		if !valid.IsLength(profileRequest.LastName, NameMinLength, NameMaxLength) {
			api.BadInputs(w, "Last name must be between 1 and 64 characters", api.FieldSize, "lastName")
			return
		}
		updatedProfile.LastName = profileRequest.LastName
	}

	if !valid.IsNull(profileRequest.Timezone) {
		if !valid.IsTimezone(profileRequest.Timezone) {
			api.BadInputs(w, "Invalid timezone", api.InvalidTimezone, "timezone")
			return
		}
		updatedProfile.Timezone = profileRequest.Timezone
	}

	appErr := pr.profileService.UpdateProfile(&updatedProfile, existingUserProfile)
	if appErr != nil {
		api.ErrorJson(w, appErr, http.StatusBadRequest)
		return
	}

	api.Json(w, r, NewProfileResponse(&updatedProfile))
}

func (pr *ProfileRouter) updatePasswordHandler(w http.ResponseWriter, r *http.Request) {
	var request PasswordChangeRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&request); err != nil {
		api.ErrorJson(w, api.NewError(err, "Invalid password change JSON", api.InvalidJson), http.StatusBadRequest)
		return
	}
	defer api.CloseBody(r.Body)

	userProfile, ok := r.Context().Value(config.ProfileContextKey).(*Profile)
	if !ok || userProfile == nil {
		api.ErrorJson(w, api.NewError(nil, "Invalid profile context", api.SystemError), http.StatusInternalServerError)
		return
	}

	if valid.IsNull(request.CurrentPassword) || !valid.IsLength(request.CurrentPassword, PasswordMinLength, PasswordMaxLength) {
		api.BadInputs(w, "Invalid current password", api.FieldSize, "currentPassword")
		return
	}

	if valid.IsNull(request.Password) || !valid.IsLength(request.Password, PasswordMinLength, PasswordMaxLength) {
		api.BadInputs(w, "Invalid password field size", api.FieldSize, "password")
		return
	}

	if valid.IsNull(request.ConfirmPassword) || !valid.IsLength(request.ConfirmPassword, PasswordMinLength, PasswordMaxLength) {
		api.BadInputs(w, "Invalid confirm password field size", api.FieldSize, "confirmPassword")
		return
	}

	err := pr.profileService.UpdatePassword(userProfile.ProfileId, request.CurrentPassword, request.Password, request.ConfirmPassword)
	if err != nil {
		api.ErrorJson(w, err, http.StatusBadRequest)
		return
	}

	api.Json(w, r, nil)
}

func (pr *ProfileRouter) ValidateSessionHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userProfile, ok := r.Context().Value(config.ProfileContextKey).(*Profile)
		if !ok || userProfile == nil {
			api.ErrorJson(w, api.NewError(nil, "No profile context found", api.ProfileNotFound), http.StatusUnauthorized)
			return
		}

		// Validate the profile
		if valid.IsNullString(userProfile.Token) {
			api.ErrorJson(w, api.NewError(nil, "Not authenticated", api.InvalidToken), http.StatusUnauthorized)
			return
		}

		// Make sure the profile is in pr valid status
		if !IsProfileStatusValid(userProfile.ProfileStatus) {
			api.ErrorJson(w, api.NewError(nil, "Profile not active", api.ProfileInactive, api.NewErrorDetail("status", string(userProfile.ProfileStatus))), http.StatusUnauthorized)
			return
		}

		tokenExpiration := userProfile.TokenExpiration.Time
		if !userProfile.TokenExpiration.Valid || tokenExpiration.Before(time.Now()) {
			api.ErrorJson(w, api.NewError(nil, "Session expired", api.TokenExpired), http.StatusUnauthorized)
			return
		}

		// If we're within 50% of the time to expiration, update the token expiration
		durationUntilExpiration := time.Until(tokenExpiration)
		if durationUntilExpiration.Minutes() < float64(GetCookieExpirationMinutes())/2 {
			start := time.Now()
			err := pr.profileService.UpdateTokenExpiration(userProfile.Token.String, GetSessionExpiration())
			if err != nil {
				logger.Log.Error("Failed to update session expiration: " + err.String())
			}
			logger.Log.Debug("UpdateAccount session took: " + fmt.Sprintf("%v\n", time.Since(start)))
		}

		next.ServeHTTP(w, r)
	})
}

func (pr *ProfileRouter) createAccountHandler(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		api.ErrorJson(w, api.NewError(nil, "Empty Body", api.InvalidJson), http.StatusBadRequest)
		return
	}

	var accountRequest AccountRequest
	if err := json.NewDecoder(r.Body).Decode(&accountRequest); err != nil {
		api.ErrorJson(w, api.NewError(err, "Invalid account JSON", api.InvalidJson), http.StatusBadRequest)
		return
	}
	defer api.CloseBody(r.Body)

	if !valid.IsEmail(accountRequest.Email) {
		api.BadInputs(w, "Invalid email", api.InvalidEmail, "email")
		return
	}

	if !valid.IsLength(accountRequest.Password, PasswordMinLength, PasswordMaxLength) {
		api.BadInputs(w, "Password must be between 6 and 64 characters", api.FieldSize, "password")
		return
	}

	if !valid.IsLength(accountRequest.FirstName, NameMinLength, NameMaxLength) {
		api.BadInputs(w, "First name must be between 1 and 64 characters", api.FieldSize, "firstName")
		return
	}

	if !valid.IsLength(accountRequest.LastName, NameMinLength, NameMaxLength) {
		api.BadInputs(w, "Last name must be between 1 and 64 characters", api.FieldSize, "lastName")
		return
	}

	if !valid.IsLength(accountRequest.Company, CompanyNameMinLength, CompanyNameMaxLength) {
		api.BadInputs(w, "Company name must be between 1 and 64 characters", api.FieldSize, "company")
		return
	}

	if valid.IsNull(accountRequest.Timezone) {
		accountRequest.Timezone = "America/New_York"
	}

	_, newUser, appError := pr.profileService.Create(&accountRequest)
	if appError != nil {
		api.ErrorJson(w, appError, http.StatusBadRequest)
		return
	}

	AddSessionCookie(w, newUser.Token.String, viper.GetString("application.applicationDomain"))
	api.Json(w, r, NewProfileResponse(newUser))
}

func (pr *ProfileRouter) updateAccountHandler(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		api.ErrorJson(w, api.NewError(nil, "Empty Body", api.InvalidJson), http.StatusBadRequest)
		return
	}

	var accountRequest AccountUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&accountRequest); err != nil {
		api.ErrorJson(w, api.NewError(err, "Invalid update JSON", api.InvalidJson), http.StatusBadRequest)
		return
	}
	defer api.CloseBody(r.Body)

	if !valid.IsLength(accountRequest.Company, CompanyNameMinLength, CompanyNameMaxLength) {
		api.BadInputs(w, "Company name must be between 1 and 64 characters", api.FieldSize, "company")
		return
	}

	if !IsWeekStart(accountRequest.WeekStart) {
		api.BadInputs(w, "Invalid week start", api.InvalidWeekStart, "weekStart")
		return
	}

	if !valid.IsTimezone(accountRequest.Timezone) {
		api.BadInputs(w, "Invalid timezone", api.InvalidTimezone, "timezone")
		return
	}

	accountProfile, ok := r.Context().Value(config.ProfileContextKey).(*Profile)
	if !ok || accountProfile == nil {
		api.ErrorJson(w, api.NewError(nil, "Invalid session accountProfile", api.SystemError), http.StatusUnauthorized)
		return
	}

	accountData, appError := pr.profileService.UpdateAccount(accountProfile.AccountId, &accountRequest)
	if appError != nil {
		api.ErrorJson(w, appError, http.StatusBadRequest)
		return
	}

	if accountData == nil {
		api.ErrorJson(w, api.NewError(nil, "No account updated", api.SystemError), http.StatusInternalServerError)
		return
	}

	api.Json(w, r, NewAccountResponse(accountData))
}

func (pr *ProfileRouter) closeAccountHandler(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		api.ErrorJson(w, api.NewError(nil, "Empty Body", api.InvalidJson), http.StatusBadRequest)
		return
	}

	type closeAccountRequest struct {
		Reason string
	}
	var closeRequest closeAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&closeRequest); err != nil {
		api.ErrorJson(w, api.NewError(err, "Invalid close account JSON", api.InvalidJson), http.StatusBadRequest)
		return
	}
	defer api.CloseBody(r.Body)

	if valid.IsNull(closeRequest.Reason) {
		api.BadInputs(w, "Must include pr valid reason for closing account", api.MissingField, "reason")
		return
	}

	Profile, ok := r.Context().Value(config.ProfileContextKey).(*Profile)
	if !ok || Profile == nil {
		api.ErrorJson(w, api.NewError(nil, "Invalid session profile account", api.SystemError), http.StatusUnauthorized)
		return
	}

	appError := pr.profileService.CloseAccount(Profile.AccountId, closeRequest.Reason)
	if appError != nil {
		api.ErrorJson(w, appError, http.StatusBadRequest)
		return
	}

	api.Json(w, r, nil)
}

func (pr *ProfileRouter) getAccountHandler(w http.ResponseWriter, r *http.Request) {
	accountProfile, ok := r.Context().Value(config.ProfileContextKey).(*Profile)
	if !ok || accountProfile == nil {
		api.ErrorJson(w, api.NewError(nil, "Invalid session profile", api.SystemError), http.StatusUnauthorized)
		return
	}

	accountData, appErr := pr.profileService.GetAccount(accountProfile.AccountId)
	if appErr != nil {
		api.ErrorJson(w, appErr, http.StatusBadRequest)
		return
	}

	if accountData == nil {
		api.ErrorJson(w, api.NewError(nil, "No account found", api.AccountInactive), http.StatusBadRequest)
		return
	}

	api.Json(w, r, NewAccountResponse(accountData))
}

func (pr *ProfileRouter) addUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		api.ErrorJson(w, api.NewError(nil, "Empty Body", api.InvalidJson), http.StatusBadRequest)
		return
	}

	accountProfile, ok := r.Context().Value(config.ProfileContextKey).(*Profile)
	if !ok || accountProfile == nil {
		api.ErrorJson(w, api.NewError(nil, "Invalid session profile", api.SystemError), http.StatusUnauthorized)
		return
	}

	var newUserRequest AddUserRequest
	if err := json.NewDecoder(r.Body).Decode(&newUserRequest); err != nil {
		api.ErrorJson(w, api.NewError(err, "Invalid add user JSON", api.InvalidJson), http.StatusBadRequest)
		return
	}

	defer api.CloseBody(r.Body)

	if !valid.IsEmail(newUserRequest.Email) {
		api.BadInputs(w, "Invalid email", api.InvalidEmail, "email")
		return
	}

	if !valid.IsLength(newUserRequest.FirstName, NameMinLength, NameMaxLength) {
		api.BadInputs(w, "First name must be between 1 and 64 characters", api.FieldSize, "firstName")
		return
	}

	if !valid.IsLength(newUserRequest.LastName, NameMinLength, NameMaxLength) {
		api.BadInputs(w, "Last name must be between 1 and 64 characters", api.FieldSize, "lastName")
		return
	}

	if !IsRole(newUserRequest.Role) {
		newUserRequest.Role = "user"
	}

	newUser, appError := pr.profileService.AddUser(&newUserRequest, &accountProfile.Account)
	if appError != nil {
		if appError.Code == EmailExistsInAccount {
			api.ErrorJson(w, appError, http.StatusOK)
		} else {
			api.ErrorJson(w, appError, http.StatusBadRequest)
		}
		return
	}

	api.Json(w, r, NewProfileResponse(newUser))
}

func (pr *ProfileRouter) getUsersHandler(w http.ResponseWriter, r *http.Request) {
	accountProfile, ok := r.Context().Value(config.ProfileContextKey).(*Profile)
	if !ok || accountProfile == nil {
		api.ErrorJson(w, api.NewError(nil, "Invalid session profile", api.SystemError), http.StatusUnauthorized)
		return
	}

	profiles, appError := pr.profileService.GetAllProfiles(accountProfile.Account.AccountId)
	if appError != nil {
		api.ErrorJson(w, appError, http.StatusBadRequest)
		return
	}

	api.Json(w, r, NewProfileListResponse(profiles))
}

func (pr *ProfileRouter) getProfileHandler(w http.ResponseWriter, r *http.Request) {
	userProfile, ok := r.Context().Value(config.ProfileContextKey).(*Profile)
	if !ok || userProfile == nil {
		api.ErrorJson(w, api.NewError(nil, "Invalid session profile", api.SystemError), http.StatusUnauthorized)
		return
	}

	api.Json(w, r, NewProfileResponse(userProfile))
}

func (pr *ProfileRouter) removeUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		api.ErrorJson(w, api.NewError(nil, "Empty Body", api.InvalidJson), http.StatusBadRequest)
		return
	}

	accountProfile, ok := r.Context().Value(config.ProfileContextKey).(*Profile)
	if !ok || accountProfile == nil {
		api.ErrorJson(w, api.NewError(nil, "Invalid session profile", api.SystemError), http.StatusUnauthorized)
		return
	}

	var removeUserRequest RemoveUserRequest
	if err := json.NewDecoder(r.Body).Decode(&removeUserRequest); err != nil {
		api.ErrorJson(w, api.NewError(err, "Invalid remove user JSON", api.InvalidJson), http.StatusBadRequest)
		return
	}
	defer api.CloseBody(r.Body)

	if valid.IsNull(removeUserRequest.Email) {
		api.BadInputs(w, "Invalid email", api.InvalidField, "email")
		return
	}

	appError := pr.profileService.RemoveUser(removeUserRequest.Email, &accountProfile.Account)
	if appError != nil {
		api.ErrorJson(w, appError, http.StatusBadRequest)
		return
	}

	api.Json(w, r, nil)
}

func (pr *ProfileRouter) AdminPermissionHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accountProfile, ok := r.Context().Value(config.ProfileContextKey).(*Profile)
		if !ok || accountProfile == nil {
			api.ErrorJson(w, api.NewError(nil, "Invalid profile account", api.SystemError), http.StatusUnauthorized)
			return
		}

		if !IsAdmin(accountProfile.Role) {
			api.ErrorJson(w, api.NewError(nil, "Not permitted", api.NotAuthorized), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (pr *ProfileRouter) ValidateProfileHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := r.Context().Value(config.TokenContextKey).(string)
		if !ok || token == "" {
			api.ErrorJson(w, api.NewError(nil, "No token found", api.MissingToken), http.StatusUnauthorized)
			return
		}

		// Get the user profile for this token
		userProfile, err := pr.profileService.GetProfileByToken(token)

		if err != nil {
			api.ErrorJson(w, err, http.StatusUnauthorized)
			return
		}

		if userProfile == nil {
			api.ErrorJson(w, api.NewError("", "Missing profile", api.ProfileNotFound), http.StatusUnauthorized)
			return
		}

		// Make sure the profile is in pr valid status
		if !IsProfileStatusValid(userProfile.ProfileStatus) {
			api.ErrorJson(w, api.NewError(nil, "Profile account not valid", api.ProfileInactive, api.NewErrorDetail("status", string(userProfile.ProfileStatus))), http.StatusUnauthorized)
			return
		}

		// Add current profile account and profile to the request context
		ctx := context.WithValue(r.Context(), config.ProfileContextKey, userProfile)

		next.ServeHTTP(w, r.WithContext(ctx))

	})
}

// Require the profile's Account to be in a valid status
func (pr *ProfileRouter) requireValidAccount(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Profile, ok := r.Context().Value(config.ProfileContextKey).(*Profile)
		if !ok || Profile == nil {
			api.ErrorJson(w, api.NewError(nil, "Invalid session profile account", api.SystemError), http.StatusUnauthorized)
			return
		}

		if !IsAccountStatusValid(Profile.Account.AccountStatus) {
			api.ErrorJson(w, api.NewError(nil, "Account not active", api.AccountInactive, api.NewErrorDetail("status", string(Profile.Account.AccountStatus))), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func NewAccountCompaniesResponse(accounts []*Account) []*CompanyResponse {
	var accountCompaniesResponse []*CompanyResponse
	for _, companyAccount := range accounts {
		accountCompaniesResponse = append(accountCompaniesResponse,
			&CompanyResponse{
				Id:      companyAccount.AccountId,
				Company: companyAccount.Company,
			})
	}
	return accountCompaniesResponse
}

func NewAccountResponse(accountData *Account) *AccountResponse {
	return &AccountResponse{
		Company:   accountData.Company,
		Timezone:  accountData.AccountTimezone,
		WeekStart: accountData.WeekStart,
		Created:   accountData.Created.Format(time.RFC3339),
		Updated:   accountData.Updated.Format(time.RFC3339),
	}
}

func NewProfileListResponse(profiles []*Profile) []ProfileResponse {
	var userList []ProfileResponse
	for _, user := range profiles {
		userList = append(userList, *NewProfileResponse(user))
	}
	return userList
}

func NewProfileResponse(user *Profile) *ProfileResponse {
	return &ProfileResponse{
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Company:   user.Company,
		Email:     user.Email,
		Phone:     user.Phone.String,
		Timezone:  user.Timezone,
	}
}
