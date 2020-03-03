package client

import (
	"database/sql"
)

const (
	ClientNameMinLength = 1
	ClientNameMaxLength = 64

	ProjectNameMinLength = 1
	ProjectNameMaxLength = 128
)

type Client struct {
	ClientId     int            `json:"-" db:"client_id"`
	AccountId    int            `json:"-" db:"account_id"`
	ClientName   string         `json:"-" db:"client_name"`
	Address      sql.NullString `json:"-"`
	ClientActive bool           `json:"-" db:"client_active"`
}
