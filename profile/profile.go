package profile

import (
	"database/sql"
	"time"

	"github.com/lib/pq"
)

type ProfileStatus string
type ProfileAccountStatus string
type AuthorizationRole string
type AccountStatus string

const (
	ProfileNew         ProfileStatus = "new"
	ProfileNotVerified ProfileStatus = "not-verified"
	ProfileValid       ProfileStatus = "valid"
)

const (
	ProfileAccountValid   ProfileAccountStatus = "valid"
	ProfileAccountInvalid ProfileAccountStatus = "invalid"
)

// Account Status
const (
	AccountNew       AccountStatus = "new"
	AccountValid     AccountStatus = "valid"
	AccountArchived  AccountStatus = "archived"
)

const (
	Owner     AuthorizationRole = "owner"
	Admin     AuthorizationRole = "admin"
	Reporting AuthorizationRole = "reporting"
	User      AuthorizationRole = "user"
)

// Profile Field Lengths
const (
	EmailMinLength       = 5
	EmailMaxLength       = 254
	PasswordMinLength    = 8
	PasswordMaxLength    = 64
	NameMinLength        = 1
	NameMaxLength        = 64
	CompanyNameMinLength = 1
	CompanyNameMaxLength = 64
)

type Account struct {
	AccountId       int           `json:"-" db:"account_id"`
	Company         string        `json:"-"`
	AccountStatus   AccountStatus `json:"-" db:"account_status"`
	WeekStart       int           `json:"-" db:"week_start"`
	AccountTimezone string        `json:"-" db:"account_timezone"`
	Created         time.Time     `json:"-"`
	Updated         time.Time     `json:"-"`
	CloseReason     string        `json:"-" db:"close_reason"`
}

type Session struct {
	Token           sql.NullString `json:"-"`
	TokenExpiration pq.NullTime    `json:"-" db:"token_expiration"`
	SessionType     string         `json:"-" db:"type"`
}

type Profile struct {
	Account
	Session
	ProfileId                int                  `json:"-" db:"profile_id"`
	FirstName                string               `json:"-" db:"first_name"`
	LastName                 string               `json:"-" db:"last_name"`
	Email                    string               `json:"-"`
	Password                 string               `json:"-"`
	Phone                    sql.NullString       `json:"-"`
	ProfileStatus            ProfileStatus        `json:"-" db:"profile_status"`
	Created                  pq.NullTime          `json:"-" db:"created"`
	Updated                  time.Time            `json:"-" db:"updated"`
	LockedUntil              pq.NullTime          `json:"-" db:"locked_until"`
	Role                     AuthorizationRole    `json:"-"`
	ProfileAccountStatus     ProfileAccountStatus `json:"-" db:"profile_account_status"`
	Timezone                 string               `json:"-"`
	ForgotPasswordToken      pq.NullTime          `json:"-" db:"forgot_password_token"`
	ForgotPasswordExpiration pq.NullTime          `json:"-" db:"forgot_password_expiration"`
}

type ForgotPassword struct {
	ProfileId                int            `json:"-" db:"profile_id"`
	ForgotPasswordToken      sql.NullString `json:"-" db:"forgot_password_token"`
	ForgotPasswordExpiration pq.NullTime    `json:"-" db:"forgot_password_expiration"`
}

const (
	MissingIpAddress = "0.0.0.0"
)
