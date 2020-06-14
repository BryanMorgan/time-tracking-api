package timesheet

import (
	"strings"
	"testing"
	"time"

	"github.com/bryanmorgan/time-tracking-api/config"
)

// Test different timezones and weekday starts with the current date to make sure we get the correct week start/end dates
func TestCurrentWeekStartEnd(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		timezone     string
		startWeekday time.Weekday
		errorPrefix  string
	}{
		{"Valid Timezone New York (Monday)", "America/New_York", time.Monday, ""},
		{"Valid Timezone Africa (Sunday)", "Africa/Monrovia", time.Sunday, ""},
		{"Valid Timezone Africa (Saturday)", "Africa/Monrovia", time.Saturday, ""},
		{"Valid Timezone Africa (Monday)", "Africa/Monrovia", time.Monday, ""},
		{"Valid Timezone (Pacific)", "America/Tijuana", time.Monday, ""},
		{"Invalid Timezone ", "TwilightZone", time.Monday, "unknown time zone TwilightZone"},
		{"Missing Timezone (use UTC)", "", time.Monday, ""},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			start, end, err := getCurrentWeekRange(testCase.timezone, testCase.startWeekday)

			if err != nil {
				t.Logf("Error: %s", err)
				if testCase.errorPrefix == "" {
					t.Errorf("Unexpected error: [%s]", err)
				} else if !strings.Contains(err.Error(), testCase.errorPrefix) {
					t.Errorf("Unexpected error: [%s] wanted error to contain: [%s]", err, testCase.errorPrefix)
				}
			}

			t.Logf("Compare start [%s] (%s) to end [%s] (%s)", start.Format(config.ISOShortDateFormat), start.Weekday(), end.Format(config.ISOShortDateFormat), end.Weekday())

			if start.IsZero() || end.IsZero() {
				t.Errorf("Got invalid start date: [%s] or end date: [%s]", start.Format(config.ISOShortDateFormat), end.Format(config.ISOShortDateFormat))
			}

			if start.Weekday() != testCase.startWeekday {
				t.Errorf("Wrong start weekday: [%s] wanted: [%s]", start.Weekday(), testCase.startWeekday)
			}

			if end.Before(start) {
				t.Errorf("End date: [%s] is before start: [%s]", end.Format(config.ISOShortDateFormat), start.Format(config.ISOShortDateFormat))
			}

			if end.AddDate(0, 0, -6).Weekday() != start.Weekday() {
				t.Errorf("End date: [%s] is not 6 days after start date: [%s]", end.Format(config.ISOShortDateFormat), start.Format(config.ISOShortDateFormat))
			}
		})
	}
}

// Test different date strings and weekday starts to make sure we get the correct start/end date ranges
func TestWeekStartEnd(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		date         string
		startWeekday time.Weekday
		startDate    string
		endDate      string
		errorPrefix  string
	}{
		{"Valid - Friday Nov 17, 2017 - Monday", "2017-11-17", time.Monday, "2017-11-13", "2017-11-19", ""},
		{"Valid - Friday Nov 17, 2017 - Sunday", "2017-11-17", time.Sunday, "2017-11-12", "2017-11-18", ""},
		{"Valid - Friday Nov 17, 2017 - Saturday", "2017-11-17", time.Saturday, "2017-11-11", "2017-11-17", ""},
		{"Valid - Sunday Nov 19, 2017 - Monday", "2017-11-19", time.Monday, "2017-11-13", "2017-11-19", ""},
		{"Valid - Sunday Nov 19, 2017 - Sunday", "2017-11-19", time.Sunday, "2017-11-19", "2017-11-25", ""},
		{"Valid - Sunday Nov 19, 2017 - Saturday", "2017-11-19", time.Saturday, "2017-11-18", "2017-11-24", ""},
		{"Invalid Date", "2017-00-01", time.Monday, "", "", "parsing time"},
		{"Invalid Date (day)", "2017-02-31", time.Monday, "", "", "parsing time"},
		{"Missing Date", "", time.Monday, "", "", "parsing time"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			start, end, err := getWeekRangeFromDate(testCase.date, testCase.startWeekday)

			if err != nil {
				t.Logf("Error: %s", err)
				if testCase.errorPrefix == "" {
					t.Errorf("Unexpected error: [%s]", err)
				} else if !strings.Contains(err.Error(), testCase.errorPrefix) {
					t.Errorf("Unexpected error: [%s] wanted error to contain: [%s]", err, testCase.errorPrefix)
				}
			} else {
				if testCase.errorPrefix != "" {
					t.Errorf("Expected error but got no errors. Expected: [%s]", testCase.errorPrefix)
				}

				expectedStart, err := time.Parse(config.ISOShortDateFormat, testCase.startDate)
				if err != nil {
					t.Errorf("Could not parse expected output start date: [%s]", testCase.startDate)
				}
				expectedEnd, err := time.Parse(config.ISOShortDateFormat, testCase.endDate)
				if err != nil {
					t.Errorf("Could not parse expected output end date: [%s]", testCase.endDate)
				}

				if expectedStart.Format(config.ISOShortDateFormat) != start.Format(config.ISOShortDateFormat) {
					t.Errorf("Invalid start date: [%s] wanted: [%s]", start.Format(config.ISOShortDateFormat), expectedStart.Format(config.ISOShortDateFormat))
				}

				if expectedEnd.Format(config.ISOShortDateFormat) != end.Format(config.ISOShortDateFormat) {
					t.Errorf("Invalid end date: [%s] wanted: [%s]", end.Format(config.ISOShortDateFormat), expectedEnd.Format(config.ISOShortDateFormat))
				}
			}

			t.Logf("Using [%s] compare [%s] (%s) to end [%s] (%s)", testCase.date, start.Format(config.ISOShortDateFormat), start.Weekday(), end.Format(config.ISOShortDateFormat), end.Weekday())

			if start.IsZero() || end.IsZero() {
				t.Errorf("Got invalid start date: [%s] or end date: [%s]", start.Format(config.ISOShortDateFormat), end.Format(config.ISOShortDateFormat))
			}

			if start.Weekday() != testCase.startWeekday {
				t.Errorf("Wrong start weekday: [%s] wanted: [%s]", start.Weekday(), testCase.startWeekday)
			}

			if end.Before(start) {
				t.Errorf("End date: [%s] is before start: [%s]", end.Format(config.ISOShortDateFormat), start.Format(config.ISOShortDateFormat))
			}

			if end.AddDate(0, 0, -6).Weekday() != start.Weekday() {
				t.Errorf("End date: [%s] is not 6 days after start date: [%s]", end.Format(config.ISOShortDateFormat), start.Format(config.ISOShortDateFormat))
			}
		})
	}
}
