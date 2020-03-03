package account

import (
	"time"
)

type AccountStatus string

const (
	AccountNew       AccountStatus = "new"
	AccountValid     AccountStatus = "valid"
	AccountArchived  AccountStatus = "archived"
	AccountSuspended AccountStatus = "suspended"
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
