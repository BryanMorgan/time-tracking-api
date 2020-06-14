package timesheet

import (
	"github.com/bryanmorgan/time-tracking-api/config"
	"github.com/bryanmorgan/time-tracking-api/logger"
	"strconv"
	"time"
)

type TimeEntry struct {
	Day         time.Time `json:"-"`
	Hours       float64   `json:"-"`
	AccountId   int       `json:"-" db:"account_id"`
	ProfileId   int       `json:"-" db:"profile_id"`
	ProjectId   int       `json:"-" db:"project_id"`
	TaskId      int       `json:"-" db:"task_id"`
	ClientName  string    `json:"-" db:"client_name"`
	ProjectName string    `json:"-" db:"project_name"`
	TaskName    string    `json:"-" db:"task_name"`
}

type Timesheet struct {
	Days      [7]TimeEntry `json:"-"`
	StartDate time.Time    `json:"-"`
	Status    string       `json:"-"`
}

// Returns the 6 day week start and end range based on the current date/time and timezone
func getCurrentWeekRange(timezone string, weekdayStart time.Weekday) (time.Time, time.Time, error) {
	location, err := time.LoadLocation(timezone)
	var start time.Time
	if err != nil {
		logger.Log.Error("Could not load timezone: " + timezone)
		start = time.Now()
	} else {
		// Find the start date based on the timezone
		start = time.Now().In(location)
	}

	start = getWeekdayStartDate(start, weekdayStart)
	end := start.AddDate(0, 0, 6)
	return start, end, err
}

// Returns the 6 day week start and end range, given an input date and the weekday to start with
func getWeekRangeFromDate(dateString string, weekdayStart time.Weekday) (time.Time, time.Time, error) {
	var start time.Time
	date, err := time.Parse(config.ISOShortDateFormat, dateString)
	if err != nil {
		logger.Log.Error("Could not parse date: " + dateString)
		start = time.Now()
	} else {
		start = date
	}

	start = getWeekdayStartDate(start, weekdayStart)
	end := start.AddDate(0, 0, 6)
	return start, end, err
}

// Find first weekday based on the start date by walking backwards 1 day until we hit the weekday start
func getWeekdayStartDate(start time.Time, weekdayStart time.Weekday) time.Time {
	for ; start.Weekday() != weekdayStart; start = start.AddDate(0, 0, -1) {
	}
	return start
}

func getWeekdayStart(weekStart int) time.Weekday {
	if weekStart >= 0 && weekStart <= 6 {
		return time.Weekday(weekStart)
	}

	logger.Log.Warn("Invalid week start conversion using value: " + strconv.Itoa(weekStart))
	return time.Monday
}
