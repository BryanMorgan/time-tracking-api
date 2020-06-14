package profile

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/bryanmorgan/time-tracking-api/database"
	"github.com/bryanmorgan/time-tracking-api/logger"
	"github.com/bryanmorgan/time-tracking-api/valid"

	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
)

const (
	EmailExistsInAccount = "EmailExistsInAccount"
)

// Compile Only: ensure interface is implemented
var _ ProfileStore = &ProfileData{}

type ProfileStore interface {
	// Profile
	CreateProfile(*Profile, int, bool, ProfileStatus) (int, error)
	GetByEmail(email string) (*Profile, error)
	GetFailedLoginCount(email string, currentTime time.Time) (int, error)
	GetPasswordById(profileId int) (string, error)
	UpdateProfile(*Profile) error
	UpdateTokenExpiration(string, time.Time) error
	UpdatePassword(profileId int, password string) error
	UpdateProfileState(profileId int, profileStatus ProfileStatus) error
	SetProfileLocked(email string) error
	DeleteSessionByToken(token string) error
	AddToken(profileId int, accountId int, token string, expiration time.Time) error
	UpdateForgotPassword(profileId int, forgotPasswordToken string, forgotPasswordExpirationMinutes int) error
	AddFailedLoginAttempt(email string, ipAddress string) error
	GetForgotPasswordToken(token string) (*ForgotPassword, error)

	// Account
	CreateAccount(account *Account) (int, error)
	GetAccount(accountId int) (*Account, error)
	GetProfiles(accountId int) ([]*Profile, error)
	GetProfileByToken(token string) (*Profile, error)
	UpdateAccount(account *Account) error
	CloseAccount(accountId int, reason string) error
	AddUser(accountId int, userId int, email string, role AuthorizationRole, status ProfileAccountStatus) error
	RemoveUser(accountId int, userId int) error
}

// ProfileData implements database operations for user profiles
type ProfileData struct {
	db *sqlx.DB
}

func NewProfileAccountStore(db *sqlx.DB) ProfileStore {
	return &ProfileData{
		db: db,
	}
}

func (pa *ProfileData) GetForgotPasswordToken(token string) (*ForgotPassword, error) {
	query := `select profile_id, forgot_password_expiration from profile where forgot_password_token = $1`

	var forgotPassword ForgotPassword
	err := pa.db.Get(&forgotPassword, query, token)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		logger.Log.Info("Failed to get forgot password details", logger.Error(err))
		return nil, err
	}

	return &forgotPassword, nil
}

func (pa *ProfileData) AddToken(profileId int, accountId int, token string, expiration time.Time) error {

	upsertSql := `
		INSERT INTO session (token, token_expiration, profile_id, account_id, type) VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (token)
		DO UPDATE SET token_expiration=$2
		WHERE session.token=$1`
	_, err := pa.db.Exec(upsertSql, token, expiration, profileId, accountId, "web")
	if err != nil {
		logger.Log.Error("Failed to upsert token into session", logger.Error(err))
		return err
	}

	updateSql := `UPDATE profile_account SET last_used=CURRENT_TIMESTAMP where profile_id = $1 and account_id = $2`
	_, err = pa.db.Exec(updateSql, profileId, accountId)
	if err != nil {
		logger.Log.Error("Failed to update profile account with last_used value", logger.Error(err))
		return err
	}
	return err
}

func (pa *ProfileData) UpdateForgotPassword(profileId int, forgotPasswordToken string, forgotPasswordExpirationMinutes int) error {
	tokenExpiration := time.Now().Add(time.Minute * time.Duration(forgotPasswordExpirationMinutes))

	updateSql := `UPDATE profile SET forgot_password_token=$1, forgot_password_expiration=$2 WHERE profile_id=$3`
	result, err := pa.db.Exec(updateSql, forgotPasswordToken, tokenExpiration, profileId)

	if err != nil {
		return err
	}

	n, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if n == 0 {
		return errors.New("profile forgot password not updated")
	}

	return nil
}

func (pa *ProfileData) GetByEmail(email string) (*Profile, error) {
	if email == "" {
		return nil, errors.New("missing email")
	}

	user := Profile{}
	query := `
		SELECT p.*, a.*
		FROM profile p,
	         profile_account pa,
	         account a
		WHERE p.email = $1
		  AND p.profile_id = pa.profile_id
		  AND pa.account_id = a.account_id
          ORDER BY pa.last_used DESC  
		  LIMIT 1`

	err := pa.db.Get(&user, query, email)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		logger.Log.Info("Failed to get user by email: " + err.Error())
		return nil, err
	}

	return &user, nil
}

func (pa *ProfileData) GetPasswordById(profileId int) (string, error) {
	var password string
	query := `SELECT password FROM profile WHERE profile_id = $1`
	err := pa.db.Get(&password, query, profileId)
	if err == sql.ErrNoRows {
		return "", nil
	}

	if err != nil {
		return "", err
	}

	return password, nil
}

func (pa *ProfileData) CreateProfile(user *Profile, accountId int, newSession bool, status ProfileStatus) (int, error) {
	tx, err := pa.db.Begin()
	if err != nil {
		return 0, err
	}

	sqlStatement := `
		INSERT INTO profile (email, password, first_name, last_name, phone, profile_status, timezone)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING profile_id`

	var userId int
	err = tx.QueryRow(sqlStatement,
		user.Email,
		user.Password,
		user.FirstName,
		user.LastName,
		user.Phone,
		status,
		user.Timezone).Scan(&userId)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			logger.Log.Error("Failed to rollback transaction: " + rollbackErr.Error())
		}
		return 0, err
	}

	if newSession {
		sqlStatement = `
		INSERT INTO session (profile_id, account_id, token, token_expiration, type)
		VALUES ($1, $2, $3, $4, $5)`

		_, err = tx.Exec(sqlStatement, userId, accountId, user.Token, user.TokenExpiration, "web")
		if err != nil {
			logger.Log.Warn(err.Error())
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				logger.Log.Error("Failed to rollback transaction: " + rollbackErr.Error())
			}
			return 0, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return 0, err
	}

	return userId, nil
}

func (pa *ProfileData) SetProfileLocked(email string) error {
	lockDuration := viper.GetInt("session.profileLockDurationMinutes")
	if lockDuration <= 0 {
		lockDuration = 5
	}
	lockExpiration := time.Now().Add(time.Minute * time.Duration(lockDuration))

	updateSql := `UPDATE profile SET locked_until=$1 WHERE email=$2`
	result, err := pa.db.Exec(updateSql, lockExpiration, email)

	if err != nil {
		return err
	}

	n, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if n == 0 {
		return errors.New("profile lock duration not updated")
	}

	return nil
}

func (pa *ProfileData) UpdateProfile(profile *Profile) error {
	updateSql := `UPDATE profile
				  SET email=$1, first_name=$2, last_name=$3, phone=$4, timezone=$5, updated=CURRENT_TIMESTAMP
				  WHERE profile_id=$6`

	result, err := pa.db.Exec(updateSql,
		profile.Email,
		profile.FirstName,
		profile.LastName,
		profile.Phone,
		profile.Timezone,
		profile.ProfileId)

	if err != nil {
		return err
	}

	n, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if n == 0 {
		return errors.New("no values in profile were updated")
	}

	return nil
}

func (pa *ProfileData) UpdatePassword(profileId int, encryptedPassword string) error {
	updateSql := `UPDATE profile SET password=$1, updated=CURRENT_TIMESTAMP WHERE profile_id=$2`
	result, err := pa.db.Exec(updateSql, encryptedPassword, profileId)

	if err != nil {
		return err
	}

	n, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if n == 0 {
		return errors.New("no values in profile were updated")
	}

	return nil
}

func (pa *ProfileData) UpdateProfileState(profileId int, profileStatus ProfileStatus) error {
	updateSql := `UPDATE profile SET profile_status=$1, updated=CURRENT_TIMESTAMP WHERE profile_id=$2`
	result, err := pa.db.Exec(updateSql, profileStatus, profileId)

	if err != nil {
		return err
	}

	n, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if n == 0 {
		return errors.New("no values in profile status were updated")
	}

	return nil
}

func (pa *ProfileData) UpdateTokenExpiration(token string, expiration time.Time) error {
	if valid.IsNull(token) {
		return errors.New("token is empty")
	}

	updateSql := `UPDATE session SET token_expiration = $1 WHERE token = $2`

	result, err := pa.db.Exec(updateSql, expiration, token)
	if err != nil {
		return err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if count <= 0 {
		return errors.New("session expiration not updated")
	}

	return nil
}

func (pa *ProfileData) AddFailedLoginAttempt(email string, ipAddress string) error {
	if valid.IsNull(ipAddress) {
		ipAddress = MissingIpAddress
	}
	insertSql := `INSERT INTO login_attempts (email, ip_address) VALUES ($1, $2)`

	rows, err := pa.db.Exec(insertSql, email, ipAddress)
	if err != nil {
		return err
	}

	count, err := rows.RowsAffected()
	if err != nil {
		return err
	}

	if count == 0 {
		return errors.New("no rows inserted into login_attempts")
	}

	return nil
}

func (pa *ProfileData) GetFailedLoginCount(email string, currentTime time.Time) (int, error) {
	var count int
	window := viper.GetInt("session.loginFailureWindowMinutes")
	query := `select count(*) from login_attempts where email=$1 and login_attempt_time > $2`
	err := pa.db.Get(&count, query, email, currentTime.Add(time.Minute*-time.Duration(window)))
	if err == sql.ErrNoRows {
		return 0, nil
	}

	if err != nil {
		logger.Log.Info("Failed to get login count", logger.Error(err))
		return 0, err
	}

	return count, nil
}

func (pa *ProfileData) DeleteSessionByToken(token string) error {
	result, err := pa.db.Exec("DELETE FROM session WHERE token = $1", token)
	if err != nil {
		return err
	}

	if rows, _ := result.RowsAffected(); rows == 0 {
		logger.Log.Warn(fmt.Sprintf("No rows deleted using token [%s]", token))
	}

	return nil
}

func (pa *ProfileData) GetAccount(accountId int) (*Account, error) {
	var err error
	account := Account{}

	accountRow := pa.db.QueryRowx("SELECT * FROM account WHERE account_id = $1 AND account_status != $2", accountId, AccountArchived)
	if accountRow.Err() != nil {
		logger.Log.Error("Failed to find account by id: " + accountRow.Err().Error())
		return nil, accountRow.Err()
	}

	err = accountRow.StructScan(&account)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		logger.Log.Info("Failed to find account by id: " + err.Error())
		return nil, err
	}

	return &account, nil
}

func (pa *ProfileData) GetProfiles(accountId int) ([]*Profile, error) {
	var err error
	var profiles []*Profile

	query := `
		SELECT p.profile_id, first_name, last_name, email
		FROM profile p, profile_account pa
		WHERE pa.account_id = $1
		AND pa.profile_id = p.profile_id
	`
	rows, err := pa.db.Queryx(query, accountId)
	if err != nil {
		return nil, err
	}
	defer database.CloseRows(rows)

	for rows.Next() {
		var userProfile Profile
		err := rows.StructScan(&userProfile)
		if err != nil {
			return nil, err
		}
		profiles = append(profiles, &userProfile)
	}

	return profiles, nil
}

func (pa *ProfileData) GetProfileByToken(token string) (*Profile, error) {
	userProfile := Profile{}
	query := `	
		SELECT a.*, p.*, pa.profile_account_status, pa.role, s.token, s.token_expiration, s.type
		FROM profile p,
     		session s,
     		account a,
     		profile_account pa
		WHERE s.token = $1
  		AND p.profile_id = s.profile_id
  		AND pa.profile_id = p.profile_id
  		AND pa.account_id = a.account_id
  		AND s.account_id = a.account_id
	`
	err := pa.db.Get(&userProfile, query, token)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		logger.Log.Info("Failed to get profile and account by token", logger.Error(err))
		return nil, err
	}

	return &userProfile, nil
}

func (pa *ProfileData) CreateAccount(newAccount *Account) (int, error) {
	tx, err := pa.db.Begin()
	if err != nil {
		return 0, err
	}

	var accountId int
	accountInsertSql := `INSERT INTO account (company, account_status, account_timezone) VALUES ($1, $2, $3) RETURNING account_id`
	err = tx.QueryRow(accountInsertSql, newAccount.Company, newAccount.AccountStatus, newAccount.AccountTimezone).Scan(&accountId)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			logger.Log.Error("Failed to rollback transaction: " + rollbackErr.Error())
		}
		return 0, err
	}

	err = tx.Commit()
	if err != nil {
		return 0, err
	}

	return accountId, nil
}

func (pa *ProfileData) UpdateAccount(updateAccount *Account) error {
	updateSql := `UPDATE account
				  SET company=$1,
				  	  week_start=$2,
				  	  account_timezone=$3,
				  	  updated=CURRENT_TIMESTAMP
				  WHERE account_id=$4`
	result, err := pa.db.Exec(updateSql,
		updateAccount.Company,
		updateAccount.WeekStart,
		updateAccount.AccountTimezone,
		updateAccount.AccountId)

	if err != nil {
		return err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if count == 0 {
		return errors.New("no account values updated")
	}

	return nil
}

func (pa *ProfileData) AddUser(accountId int, profileId int, email string, role AuthorizationRole, status ProfileAccountStatus) error {
	// Make sure we're not trying to add an email that already exists in this account
	checkIfEmailAlreadyInAccountSql := `
        SELECT count(1)
		FROM profile p, 
			 profile_account pa
		WHERE p.email = $1		
		 AND pa.account_id = $2 
		 AND p.profile_id = pa.profile_id`

	var emailExistsInAccount int
	err := pa.db.Get(&emailExistsInAccount, checkIfEmailAlreadyInAccountSql, email, accountId)
	if err != nil {
		return err
	}

	if emailExistsInAccount > 0 {
		return errors.New(EmailExistsInAccount)
	}

	profileAccountSql := `
		INSERT INTO profile_account (profile_id, account_id, role, profile_account_status)
		VALUES ($1, $2, $3, $4)`

	result, err := pa.db.Exec(profileAccountSql, profileId, accountId, role, status)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return database.NoRowAffectedError
	}

	return nil
}

func (pa *ProfileData) RemoveUser(accountId int, userId int) error {
	deleteProfileAccountSql := `DELETE FROM profile_account WHERE account_id=$1 AND profile_id=$2`

	result, err := pa.db.Exec(deleteProfileAccountSql, accountId, userId)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return database.NoRowAffectedError
	}

	return nil
}

func (pa *ProfileData) CloseAccount(accountId int, reason string) error {
	updateAccountSql := `UPDATE account SET account_status=$1, close_reason=$2 WHERE account_id=$3`

	result, err := pa.db.Exec(updateAccountSql, AccountArchived, reason, accountId)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return database.NoRowAffectedError
	}

	return nil
}
