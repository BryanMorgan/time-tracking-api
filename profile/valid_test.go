package profile

import (
	"testing"

	"github.com/bryanmorgan/time-tracking-api/api"
)

func TestProfileStatusValid(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		status string
		valid  bool
	}{
		{"New", string(ProfileNew), true},
		{"Valid", string(ProfileValid), true},
		{"Inactive", string(api.ProfileInactive), false},
		{"Unknown", "unknown", false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.valid != IsProfileStatusValid(ProfileStatus(testCase.status)) {
				t.Errorf("Profile status check failed: [%s] wanted: [%v]", testCase.status, testCase.valid)
			}
		})
	}
}

func TestRoleValidation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		role  string
		valid bool
	}{
		{"ProfileValid (Owner)", string(Owner), true},
		{"ProfileValid (Admin)", string(Admin), true},
		{"ProfileValid (User)", string(User), true},
		{"ProfileValid (Reporting)", string(Reporting), true},
		{"Invalid (Tester)", "tester", false},
		{"Invalid Empty", "", false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.valid != IsRole(testCase.role) {
				t.Errorf("Role note valid. Expected %t for role '%s'", testCase.valid, testCase.role)
			}
		})
	}
}

func TestAdminRole(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name  string
		role  AuthorizationRole
		valid bool
	}{
		{"ProfileValid (Owner)", Owner, true},
		{"ProfileValid (Admin)", Admin, true},
		{"Invalid (User)", User, false},
		{"Invalid (Tester)", "tester", false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.valid != IsAdmin(testCase.role) {
				t.Errorf("Not admin: [%s] wanted: [%v]", testCase.role, testCase.valid)
			}
		})
	}
}

func TestAccountStatusValid(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		status string
		valid  bool
	}{
		{"New", string(AccountNew), true},
		{"Valid", string(AccountValid), true},
		{"Inactive", string(api.AccountInactive), false},
		{"Unknown", "unknown", false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.valid != IsAccountStatusValid(AccountStatus(testCase.status)) {
				t.Errorf("Account status check failed: [%s] wanted: [%v]", testCase.status, testCase.valid)
			}
		})
	}
}
func TestProfileAccountStatusValid(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		status string
		valid  bool
	}{
		{"Valid", string(ProfileAccountValid), true},
		{"Invalid", string(ProfileAccountInvalid), false},
		{"Unknown", "unknown", false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if testCase.valid != IsProfileAccountStatusValid(ProfileAccountStatus(testCase.status)) {
				t.Errorf("Profile account status check failed: [%s] wanted: [%v]", testCase.status, testCase.valid)
			}
		})
	}
}
