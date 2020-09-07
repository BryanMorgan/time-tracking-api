package reporting

import (
	"log"
	"testing"
	"time"

	"github.com/bryanmorgan/time-tracking-api/config"
)

func TestAdjustForWeekStart(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		today         string
		weekStart     int
		finalFromDate string
		finalToDate   string
	}{
		{"Sun 0", "2019-01-06", 0, "2019-01-06", "2019-01-12"},
		{"Sun 1", "2019-01-06", 1, "2018-12-31", "2019-01-06"},
		{"Sun 2", "2019-01-06", 2, "2019-01-01", "2019-01-07"},
		{"Sun 3", "2019-01-06", 3, "2019-01-02", "2019-01-08"},
		{"Sun 4", "2019-01-06", 4, "2019-01-03", "2019-01-09"},
		{"Sun 5", "2019-01-06", 5, "2019-01-04", "2019-01-10"},
		{"Sun 6", "2019-01-06", 6, "2019-01-05", "2019-01-11"},
		{"Sat 0", "2019-01-12", 0, "2019-01-06", "2019-01-12"},
		{"Sat 1", "2019-01-12", 1, "2019-01-07", "2019-01-13"},
		{"Sat 2", "2019-01-12", 2, "2019-01-08", "2019-01-14"},
		{"Sat 3", "2019-01-12", 3, "2019-01-09", "2019-01-15"},
		{"Sat 4", "2019-01-12", 4, "2019-01-10", "2019-01-16"},
		{"Sat 5", "2019-01-12", 5, "2019-01-11", "2019-01-17"},
		{"Sat 6", "2019-01-12", 6, "2019-01-12", "2019-01-18"},
		{"Mon 0", "2019-01-07", 0, "2019-01-06", "2019-01-12"},
		{"Mon 1", "2019-01-07", 1, "2019-01-07", "2019-01-13"},
		{"Mon 2", "2019-01-07", 2, "2019-01-01", "2019-01-07"},
		{"Mon 3", "2019-01-07", 3, "2019-01-02", "2019-01-08"},
		{"Mon 4", "2019-01-07", 4, "2019-01-03", "2019-01-09"},
		{"Mon 5", "2019-01-07", 5, "2019-01-04", "2019-01-10"},
		{"Mon 6", "2019-01-07", 6, "2019-01-05", "2019-01-11"},
	}

	fromDate := getDate("2019-01-06") // Sunday
	toDate := getDate("2019-01-12")   // Saturday

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			from, to := AdjustForWeekStart(fromDate, toDate, testCase.weekStart, getDate(testCase.today))
			if getDate(testCase.finalFromDate) != from || getDate(testCase.finalToDate) != to {
				t.Errorf("Date adjustment incorrect. Got [%s] [%s] Expected: [%s] [%s]", from, to, testCase.finalFromDate, testCase.finalToDate)
			}
		})
	}
}

func getDate(dateString string) time.Time {
	date, err := time.Parse(config.ISOShortDateFormat, dateString)
	if err != nil {
		log.Panicf("Failed to parse date: " + dateString)
	}
	return date
}
