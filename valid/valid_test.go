package valid

import (
	"database/sql"
	"testing"
)

func TestEmail(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		email  string
		result bool
	}{
		{"Email valid", "a@a.com", true},
		{"Minimal Email", "a@a.a", true},
		{"Number Suffix", "1@1.1", true},
		{"Empty", "", false},
		{"Too Short Prefix", "@a.com", false},
		{"Too Short Suffix", "a@.", false},
		{"Small Suffix", "a@a", false},
		{"Underscore Prefix", "_a@a", false},
		{"Underscore Suffix", "a@_", false},
		{"Underscore Dot Suffix", "a@._", false},
		{"Underscore Dot Letter Suffix", "a@a._", false},
		{"Number Prefix", "1@a.org", true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if IsEmail(testCase.email) != testCase.result {
				t.Errorf("Expected %v got %v", testCase.result, !testCase.result)
			}
		})
	}
}

func TestLength(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		testString string
		minLength  int
		maxLength  int
		result     bool
	}{
		{"ProfileValid", "a", 1, 1, true},
		{"Empty", "", 1, 1, false},
		{"Too Short", "a", 20, 1, false},
		{"Too Long", "aa", 1, 1, false},
		{"No Minimum", "a", 0, 1, true},
		{"Empty No Minimum", "", 0, 1, true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if IsLength(testCase.testString, testCase.minLength, testCase.maxLength) != testCase.result {
				t.Errorf("Expected %v got %v", testCase.result, !testCase.result)
			}
		})
	}
}

func TestIntBetween(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		testString string
		minValue   int
		maxValue   int
		value      int
		result     bool
	}{
		{"ProfileValid", "1", 1, 1, 1, true},
		{"Empty Number", "", 0, 1, 0, false},
		{"Invalid Number", "a", 1, 1, 1, false},
		{"ProfileValid 123", "123", 1, 200, 123, true},
		{"ProfileValid Min", "1", 1, 200, 1, true},
		{"ProfileValid Max", "200", 1, 200, 200, true},
		{"Invalid Min", "0", 1, 200, 0, false},
		{"Invalid Max", "201", 1, 200, 0, false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			value, ok := IsIntBetween(testCase.testString, testCase.minValue, testCase.maxValue)
			if ok != testCase.result {
				t.Errorf("Expected %v got %v", testCase.result, ok)
			}

			if ok {
				if value != testCase.value {
					t.Errorf("Value not correct. Expected %d got %d", testCase.value, value)
				}
			}
		})
	}
}

func TestInt(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		testString string
		value      int
		result     bool
	}{
		{"ProfileValid", "1", 1, true},
		{"Alpha", "a", 0, false},
		{"Empty", "", 0, false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			value, ok := IsInt(testCase.testString)
			if ok != testCase.result {
				t.Errorf("Expected %v got %v", testCase.result, ok)
			}

			if ok {
				if value != testCase.value {
					t.Errorf("Value not correct. Expected %d got %d", testCase.value, value)
				}
			}
		})
	}
}

func TestNull(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		testString string
		result     bool
	}{
		{"Null", "", true},
		{"Not Null - Empty Space", " ", false},
		{"Not Null - Empty Single", "_", false},
		{"Not Null - Empty Larger", "123", false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if IsNull(testCase.testString) != testCase.result {
				t.Errorf("Expected %v got %v", testCase.result, !testCase.result)
			}
		})
	}
}

func TestNullString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		testString sql.NullString
		result     bool
	}{
		{"Null", sql.NullString{String: ""}, true},
		{"Empty String", ToNullString(""), true},
		{"Not Null - Empty Space", ToNullString(" "), false},
		{"Not Null - Empty Single", ToNullString("_"), false},
		{"Not Null - Empty Larger", ToNullString("123"), false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if IsNullString(testCase.testString) != testCase.result {
				t.Errorf("Expected %v got %v", testCase.result, !testCase.result)
			}
		})
	}
}

func TestIsTimeZone(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		timeZone string
		result   bool
	}{
		{"Valid: New York", "America/New_York", true},
		{"Valid: Africa", "Africa/Monrovia", true},
		{"Invalid: TwilightZone", "TwilightZone", false},
		{"Invalid: empty", "", false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if IsTimezone(testCase.timeZone) != testCase.result {
				t.Errorf("Expected %v got %v", testCase.result, !testCase.result)
			}
		})
	}
}
