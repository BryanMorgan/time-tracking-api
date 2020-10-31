package profile

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/bryanmorgan/time-tracking-api/emails"
	"strconv"
	"strings"
	"time"

	"github.com/bryanmorgan/time-tracking-api/config"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"

	"github.com/bryanmorgan/time-tracking-api/api"
	"github.com/bryanmorgan/time-tracking-api/logger"
	"github.com/bryanmorgan/time-tracking-api/valid"
)

// Compile Only: ensure interface is implemented
var _ ProfileService = &ProfileResource{}

func NewProfileService(store ProfileStore) ProfileService {
	return &ProfileResource{store: store}
}


type AccountUpdateRequest struct {
	Company   string
	Phone     string
	Timezone  string
	WeekStart int
}

type AddUserRequest struct {
	FirstName string
	LastName  string
	Email     string
	Role      string
}

type RemoveUserRequest struct {
	Email string
}
type ProfileResource struct {
	store ProfileStore
}

type ProfileService interface {
	Login(email string, password string, ipAddress string) (*Profile, *api.Error)
	Logout(token string) *api.Error
	GetProfile(token string) (*Profile, *api.Error)
	ForgotPassword(email string) *api.Error
	ValidateForgotPasswordToken(token string) *api.Error
	SetupNewUser(token string, password string) *api.Error
	GetAccount(int) (*Account, *api.Error)
	GetAllProfiles(accountId int) ([]*Profile, *api.Error)
	GetProfileByToken(token string) (*Profile, *api.Error)
	Create(*AccountRequest) (*Account, *Profile, *api.Error)
	AddUser(*AddUserRequest, *Account) (*Profile, *api.Error)
	RemoveUser(email string, account *Account) *api.Error
	UpdateProfile(updatedProfile *Profile, existingProfile *Profile) *api.Error
	UpdateTokenExpiration(token string, expiration time.Time) *api.Error
	UpdatePassword(profileId int, currentPassword string, password string, confirmPassword string) *api.Error
	UpdateAccount(accountId int, request *AccountUpdateRequest) (*Account, *api.Error)
	CloseAccount(accountId int, reason string) *api.Error
}

func (pr *ProfileResource) Login(email string, password string, ipAddress string) (*Profile, *api.Error) {
	// Convert email to all lower case
	email = strings.ToLower(email)

	var err error
	user, err := pr.store.GetByEmail(email)
	if err != nil {
		return nil, api.NewError(err, "Failed to find profile by email", api.SystemError, api.NewErrorDetail("field", "email"))
	}

	if user == nil {
		return nil, api.NewError(nil, "No user found for email: "+email, api.ProfileNotFound, api.NewErrorDetail("field", "email"))
	}

	// See if this profile is locked
	if user.LockedUntil.Valid && user.LockedUntil.Time.After(time.Now()) {
		return nil, api.NewError(nil, "Profile locked ", api.ProfileLocked, api.NewErrorDetail("until", user.LockedUntil.Time.String()))
	}

	// Make sure the profile is in a valid state
	if !IsProfileStatusValid(user.ProfileStatus) {
		return nil, api.NewError(nil, "Profile not valid", api.ProfileInactive, api.NewErrorDetail("status", user.ProfileStatus))
	}

	// Comparing the password with the hash
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		// Passwords do not match. Only log if there was a real error
		if err != bcrypt.ErrMismatchedHashAndPassword {
			logger.Log.Error("Failed to compare passwords: ", logger.Error(err))
		}

		// Add a row for the failed password attempt
		err = pr.store.AddFailedLoginAttempt(email, ipAddress)
		if err != nil {
			logger.Log.Error("Could not add failed login attempt: ", logger.Error(err))
		}

		count, err := pr.store.GetFailedLoginCount(email, time.Now())
		if err != nil {
			logger.Log.Error("Could not get failed login count: ", logger.Error(err))
		}

		maxLoginAttempts := viper.GetInt("session.maxFailedLoginAttempts")
		logger.Log.Debug("Failed login attempt: " + strconv.Itoa(count) + " of " + strconv.Itoa(maxLoginAttempts))

		if count >= maxLoginAttempts {
			err = pr.store.SetProfileLocked(email)
			if err != nil {
				logger.Log.Error("Could not update profile to set as locked: ", logger.Error(err))
			}
		}

		return nil, api.NewError(nil, "Incorrect password", api.IncorrectPassword, api.NewErrorDetail("field", "password"))
	}

	// UpdateAccount token
	appErr := pr.GenerateNewTokenForProfile(user)
	if appErr != nil {
		return nil, appErr
	}

	return user, nil
}

func (pr *ProfileResource) ForgotPassword(email string) *api.Error {
	// Convert email to all lower case
	email = strings.ToLower(email)

	authProfile, err := pr.store.GetByEmail(email)
	if err != nil {
		return api.NewError(err, "No profile", api.SystemError)
	}

	if authProfile == nil {
		return api.NewError(nil, "No profile", api.ProfileNotFound)
	}

	forgotPasswordToken, err := generateForgotPasswordToken()
	if err != nil {
		return api.NewError(err, "Failed to generate forgot password token", api.SystemError)
	}

	forgotPasswordExpirationMinutes := viper.GetInt("session.forgotPasswordExpirationInMinutes")
	err = pr.store.UpdateForgotPassword(authProfile.ProfileId, forgotPasswordToken, forgotPasswordExpirationMinutes)
	if err != nil {
		return api.NewError(err, "Failed to update forgot password details", api.SystemError)
	}

	// Build reset URL with token
	resetUrl := config.CreateUrl("/reset-password", "verify-token="+forgotPasswordToken)

	err = emails.SendForgotPasswordEmail(authProfile.FirstName, authProfile.Email, resetUrl)
	if err != nil {
		return api.NewError(err, "Failed to send email", api.SystemError)
	}

	return nil
}

func (pr *ProfileResource) ValidateForgotPasswordToken(token string) *api.Error {
	forgotPassword, err := pr.store.GetForgotPasswordToken(token)
	if err != nil {
		return api.NewError(err, "Failed to get forgot password token", api.SystemError)
	}

	if forgotPassword == nil {
		return api.NewError(nil, "Invalid forgot password token", api.InvalidForgotToken)
	}

	if !forgotPassword.ForgotPasswordExpiration.Valid || forgotPassword.ForgotPasswordExpiration.Time.Before(time.Now()) {
		return api.NewError(err, "Expired or invalid forgot password token", api.InvalidForgotToken)
	}

	// Successful, now clear out the values so it cannot be used again
	clearToken := viper.GetBool("session.clearForgotPasswordOnValidate")
	if clearToken {
		err = pr.store.UpdateForgotPassword(forgotPassword.ProfileId, "", -1)
		if err != nil {
			return api.NewError(err, "Failed to clear forgot password", api.SystemError)
		}
	}

	return nil
}

func (pr *ProfileResource) SetupNewUser(token string, password string) *api.Error {
	newUserTokenData, err := pr.store.GetForgotPasswordToken(token)
	if err != nil {
		return api.NewError(err, "Failed to get new user token", api.SystemError)
	}

	if newUserTokenData == nil {
		return api.NewError(nil, "Missing new user token", api.InvalidToken)
	}

	if !newUserTokenData.ForgotPasswordExpiration.Valid || newUserTokenData.ForgotPasswordExpiration.Time.Before(time.Now()) {
		return api.NewError(err, "Expired or invalid new user token", api.TokenExpired)
	}

	// Successful, now clear out the values so it cannot be used again
	err = pr.store.UpdateForgotPassword(newUserTokenData.ProfileId, "", -1)
	if err != nil {
		return api.NewError(err, "Failed to new user token", api.SystemError)
	}

	// Update profile with new password
	encryptedPassword, err := EncryptPassword(password)
	if err != nil {
		return api.NewError(err, "Failed to encrypt password", api.EncryptionFailed)
	}

	err = pr.store.UpdatePassword(newUserTokenData.ProfileId, encryptedPassword)
	if err != nil {
		return api.NewError(nil, "Failed to update password", api.SystemError)
	}

	err = pr.store.UpdateProfileState(newUserTokenData.ProfileId, ProfileValid)
	if err != nil {
		return api.NewError(nil, "Failed to update profile state", api.SystemError)
	}

	return nil
}

func EncryptPassword(password string) (string, error) {
	start := time.Now()
	bytePass := []byte(password)
	hashedPassword, err := bcrypt.GenerateFromPassword(bytePass, 12)
	if err != nil {
		logger.Log.Error("encryptPassword failed", logger.NamedError(api.ErrorStatus, err))
		return "", err
	}

	elapsed := time.Since(start)
	logger.Log.Debug("Encrypt password", logger.Duration("duration", elapsed))
	return string(hashedPassword), err
}

func GenerateToken() (string, error) {
	tokenLength := viper.GetInt("session.tokenLength")
	if tokenLength <= 0 {
		tokenLength = 128
	}

	b, err := generateRandomBytes(tokenLength)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func generateForgotPasswordToken() (string, error) {
	tokenLength := viper.GetInt("session.forgotPasswordTokenLength")
	if tokenLength <= 0 {
		tokenLength = 256
	}

	b, err := generateRandomBytes(tokenLength)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func GenerateTokenExpiration() time.Time {
	tokenExpirationMinutes := viper.GetInt("session.tokenExpirationMinutes")
	if tokenExpirationMinutes <= 0 {
		tokenExpirationMinutes = 60 * 24 * 30 // 30 days
	}

	return time.Now().Add(time.Minute * time.Duration(tokenExpirationMinutes))
}

func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// ---------------------------------------------

func (pr *ProfileResource) UpdateTokenExpiration(token string, expiration time.Time) *api.Error {
	if valid.IsNull(token) {
		return api.NewError(nil, "Empty token", api.MissingToken)
	}

	if err := pr.store.UpdateTokenExpiration(token, expiration); err != nil {
		return api.NewError(err, "Failed to update session expiration", api.SystemError)
	}

	return nil
}

func (pr *ProfileResource) UpdatePassword(profileId int, currentPassword string, password string, confirmPassword string) *api.Error {
	if valid.IsNull(password) {
		return api.NewError(nil, "Empty password", api.InvalidPassword)
	}

	if password != confirmPassword {
		return api.NewError(nil, "Confirm password does not match", api.PasswordMismatch)
	}

	oldPassword, err := pr.store.GetPasswordById(profileId)
	if err != nil {
		return api.NewError(err, "Failed to get password for profile", api.SystemError)
	}

	err = bcrypt.CompareHashAndPassword([]byte(oldPassword), []byte(currentPassword))
	if err != nil {
		// Passwords do not match. Only log if there was a real error
		if err != bcrypt.ErrMismatchedHashAndPassword {
			logger.Log.Error("Failed to compare passwords: ", logger.Error(err))
			return api.NewError(nil, "Passwords failed to match", api.InvalidPassword)
		}
		return api.NewError(nil, "Passwords do not match", api.InvalidPassword)
	}

	encryptedPassword, err := EncryptPassword(password)
	if err != nil {
		return api.NewError(err, "Failed to encrypt password", api.EncryptionFailed)
	}

	err = pr.store.UpdatePassword(profileId, encryptedPassword)
	if err != nil {
		return api.NewError(nil, "Failed to update password", api.SystemError)
	}

	return nil
}

func (pr *ProfileResource) UpdateProfile(profile *Profile, existingProfile *Profile) *api.Error {
	if profile.Email != existingProfile.Email {
		existingUser, err := pr.store.GetByEmail(profile.Email)
		if err != nil {
			return api.NewError(err, "Lookup by email failed for: "+profile.Email, api.SystemError)
		}

		if existingUser != nil {
			return api.NewError(nil, "Profile exists for email: "+profile.Email, api.AccountExists)
		}
	}

	err := pr.store.UpdateProfile(profile)
	if err != nil {
		return api.NewError(err, "Failed to update profile", api.SystemError)
	}

	return nil
}

func (pr *ProfileResource) Logout(token string) *api.Error {
	err := pr.store.DeleteSessionByToken(token)
	if err != nil {
		return api.NewError(err, "Failed to logout", api.SystemError)
	}

	return nil
}

func (pr *ProfileResource) GetProfile(token string) (*Profile, *api.Error) {
	profile, err := pr.store.GetProfileByToken(token)
	if err != nil {
		return nil, api.NewError(err, "Failed to get profile", api.SystemError)
	}

	if profile == nil {
		return nil, nil
	}

	return profile, nil
}

func (pr *ProfileResource) Create(accountRequest *AccountRequest) (*Account, *Profile, *api.Error) {
	accountRequest.Email = strings.ToLower(accountRequest.Email)

	user, err := pr.store.GetByEmail(accountRequest.Email)
	if err != nil {
		return nil, nil, api.NewFieldError(err, "Lookup by email failed for: "+accountRequest.Email, api.SystemError, "email")
	}

	encryptedPassword, err := EncryptPassword(accountRequest.Password)
	if err != nil {
		return nil, nil, api.NewFieldError(err, "Failed to encrypt password", api.EncryptionFailed, "")
	}

	token, err := GenerateToken()
	if err != nil {
		return nil, nil, api.NewFieldError(nil, "Failed to generate token", api.TokenCreationFailed, "")
	}

	newAccount := Account{
		Company:         accountRequest.Company,
		AccountStatus:   AccountValid,
		AccountTimezone: accountRequest.Timezone,
	}

	accountId, err := pr.store.CreateAccount(&newAccount)
	if err != nil {
		return nil, nil, api.NewError(err, "Failed to create account", api.AccountCreateFailed)
	}

	if user == nil {
		user = &Profile{
			Email:     accountRequest.Email,
			Password:  encryptedPassword,
			FirstName: accountRequest.FirstName,
			Account: Account{
				AccountId: accountId,
				Company:   accountRequest.Company,
			},
			LastName: accountRequest.LastName,
			Timezone: accountRequest.Timezone,
			Session: Session{
				Token:           valid.ToNullString(token),
				TokenExpiration: valid.ToNullTime(GenerateTokenExpiration()),
			},
		}

		userId, err := pr.store.CreateProfile(user, accountId, true, ProfileValid)
		if err != nil {
			return nil, nil, api.NewError(err, "Failed to create profile", api.ProfileCreateFailed)
		}

		user.ProfileId = userId
	}

	// Associate user profile to account with an initial role and status
	err = pr.store.AddUser(accountId, user.ProfileId, user.Email, Owner, ProfileAccountValid)
	if err != nil {
		return nil, nil, api.NewError(err, "Failed to add user to account", api.ProfileCreateFailed)
	}

	return &newAccount, user, nil
}

func (pr *ProfileResource) AddUser(request *AddUserRequest, account *Account) (*Profile, *api.Error) {
	// Convert email to all lower case
	request.Email = strings.ToLower(request.Email)

	var err error
	user, err := pr.store.GetByEmail(request.Email)
	if err != nil {
		return nil, api.NewFieldError(err, "Failed to check for existing email: "+request.Email, api.SystemError, "email")
	}

	if user == nil {
		// Generate placeholder token for password - never used by user
		tempPassword, err := GenerateToken()
		if err != nil {
			return nil, api.NewError(err, "Could not generate temp password: ", api.SystemError)

		}
		user = &Profile{
			Email:     request.Email,
			Password:  tempPassword,
			FirstName: request.FirstName,
			LastName:  request.LastName,
		}

		// Create new user profile
		profileId, err := pr.store.CreateProfile(user, account.AccountId, false, ProfileNotVerified)
		if err != nil {
			return nil, api.NewError(err, "Failed to add user", api.ProfileCreateFailed)
		}
		user.ProfileId = profileId

		forgotPasswordToken, err := generateForgotPasswordToken()
		if err != nil {
			return nil, api.NewError(err, "Failed to generate add user login token", api.SystemError)
		}

		forgotPasswordExpirationMinutes := viper.GetInt("session.addUserTokenExpirationInMinutes")
		err = pr.store.UpdateForgotPassword(profileId, forgotPasswordToken, forgotPasswordExpirationMinutes)
		if err != nil {
			return nil, api.NewError(err, "Failed to update add user login expiration", api.SystemError)
		}

		// TODO: send user user email and create page for '/new-user' experience, where we validate email,
		// token and ask for password, then update the state to ProfileValid
		//firstNameEncoded := base64.RawURLEncoding.EncodeToString([]byte(user.FirstName))
		//resetUrl := config.CreateUrl("/new-user/"+firstNameEncoded+"/"+forgotPasswordToken, "")
		//
		//logger.Log.Debug(fmt.Sprintf("TODO: Send email for user %d URL: %s", profileId, resetUrl))
	}

	// Associate user profile to account with an initial role and status
	err = pr.store.AddUser(account.AccountId, user.ProfileId, user.Email, AuthorizationRole(request.Role), ProfileAccountValid)
	if err != nil {
		if err.Error() == EmailExistsInAccount {
			return nil, api.NewError(nil, "Email exists in account", api.EmailExistsInAccount)
		} else {
			return nil, api.NewError(err, "Failed to add new user to account", api.ProfileCreateFailed)
		}
	}

	return user, nil
}

func (pr *ProfileResource) UpdateAccount(accountId int, request *AccountUpdateRequest) (*Account, *api.Error) {
	currentAccount, err := pr.store.GetAccount(accountId)
	if err != nil {
		return nil, api.NewError(nil, "Failed to get current account", api.SystemError)
	}

	updatedAccountData := *currentAccount

	if !valid.IsNull(request.Company) {
		updatedAccountData.Company = request.Company
	}
	if !valid.IsNull(request.Timezone) {
		updatedAccountData.AccountTimezone = request.Timezone
	}
	if request.WeekStart >= 0 {
		updatedAccountData.WeekStart = request.WeekStart
	}

	err = pr.store.UpdateAccount(&updatedAccountData)
	if err != nil {
		return nil, api.NewError(err, "Failed to update account", api.UpdateFailed)
	}

	return &updatedAccountData, nil
}

func (pr *ProfileResource) CloseAccount(accountId int, reason string) *api.Error {
	err := pr.store.CloseAccount(accountId, reason)
	if err != nil {
		return api.NewError(err, "Failed to close account", api.UpdateFailed)
	}

	return nil
}

func (pr *ProfileResource) GetAccount(accountId int) (*Account, *api.Error) {
	accountData, err := pr.store.GetAccount(accountId)
	if err != nil {
		return nil, api.NewError(err, "Failed to get account", api.SystemError)
	}

	return accountData, nil
}

func (pr *ProfileResource) GetAllProfiles(accountId int) ([]*Profile, *api.Error) {
	profiles, err := pr.store.GetProfiles(accountId)
	if err != nil {
		return nil, api.NewError(err, "Failed to get profiles for account", api.SystemError)
	}

	return profiles, nil
}

func (pr *ProfileResource) GetProfileByToken(token string) (*Profile, *api.Error) {
	profileAccount, err := pr.store.GetProfileByToken(token)

	if err != nil {
		return nil, api.NewError(err, "Failed to get profile and account", api.SystemError)
	}

	if profileAccount == nil {
		return nil, api.NewError(err, "No matching profile and account", api.ProfileNotFound)
	}

	return profileAccount, nil
}

func (pr *ProfileResource) RemoveUser(email string, account *Account) *api.Error {
	email = strings.ToLower(email)
	user, err := pr.store.GetByEmail(email)
	if err != nil {
		return api.NewError(err, "Failed to find user in account", api.SystemError)
	}

	if user == nil {
		return api.NewError(err, "Failed to find user in account", api.ProfileNotFound)
	}

	if user.AccountId != account.AccountId {
		return api.NewError(err, "Failed to find user in account", api.ProfileNotFound)
	}

	err = pr.store.RemoveUser(account.AccountId, user.ProfileId)
	if err != nil {
		return api.NewError(err, "Failed to remove user from account", api.SystemError)
	}

	return nil
}

func (pr *ProfileResource) GenerateNewTokenForProfile(profile *Profile) *api.Error {
	if profile == nil {
		return api.NewError(nil, "Invalid profile", api.ProfileNotFound)
	}

	token, err := GenerateToken()
	if err != nil {
		return api.NewError(err, "Failed to generate token", api.TokenCreationFailed)
	}

	tokenExpirationMinutes := GenerateTokenExpiration()

	profile.Token = valid.ToNullString(token)
	profile.TokenExpiration = valid.ToNullTime(tokenExpirationMinutes)

	err = pr.store.AddToken(profile.ProfileId, profile.AccountId, token, tokenExpirationMinutes)
	if err != nil {
		return api.NewError(err, "Failed to add new token", api.TokenCreationFailed)
	}

	return nil
}
