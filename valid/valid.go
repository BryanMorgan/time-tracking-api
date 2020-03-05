package valid

import (
	"database/sql"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/lib/pq"
)

const (
	EmailPattern = "^[a-zA-Z0-9_.+-]+@[a-zA-Z0-9-]+\\.[a-zA-Z0-9-.]+$"
)

var (
	EmailRegEx = regexp.MustCompile(EmailPattern)
)

func IsNull(s string) bool {
	return len(s) == 0
}

func IsNullString(s sql.NullString) bool {
	return !s.Valid || IsNull(s.String)
}

func IsEmail(s string) bool {
	return EmailRegEx.MatchString(s)
}

func IsInt(s string) (int, bool) {
	value, err := strconv.Atoi(s)
	if err != nil {
		return 0, false
	}
	return value, true
}

func IsIntBetween(s string, min int, max int) (int, bool) {
	value, err := strconv.Atoi(s)
	if err == nil {
		if value >= min && value <= max {
			return value, true
		}
	}
	return 0, false
}

func IsLength(s string, min int, max int) bool {
	length := utf8.RuneCountInString(s)
	if length >= min && length <= max {
		return true
	}
	return false
}

func IsTimezone(zone string) bool {
	return !IsNull(zone) && strings.Contains(zone, "/")
}

func ToNullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func ToNullTime(t time.Time) pq.NullTime {
	return pq.NullTime{Time: t, Valid: !t.IsZero()}
}

func ToNullFloat64(f float64) sql.NullFloat64 {
	return sql.NullFloat64{Float64: f, Valid: f != 0.0}
}
