package profile

// Returns true if the ProfileStatus is new or valid
func IsProfileStatusValid(status ProfileStatus) bool {
	if status == ProfileNew || status == ProfileValid {
		return true
	}

	return false
}

func IsRole(role string) bool {
	authorizationRole := AuthorizationRole(role)
	switch authorizationRole {
	case Admin:
		return true
	case Owner:
		return true
	case Reporting:
		return true
	case User:
		return true
	default:
		return false
	}
}

// Returns true if the AuthorizationRole has admin privileges
func IsAdmin(role AuthorizationRole) bool {
	if role == Admin || role == Owner {
		return true
	}

	return false
}

// Returns true if the AccountStatus is new or valid
func IsAccountStatusValid(status AccountStatus) bool {
	if status == AccountNew || status == AccountValid {
		return true
	}

	return false
}

// Returns true if the ProfileAccountStatus is valid
func IsProfileAccountStatusValid(status ProfileAccountStatus) bool {
	return status == ProfileAccountValid
}

// Returns true if the start of week is between 0 and 6 (Sunday to Saturday)
func IsWeekStart(weekStart int) bool {
	return weekStart >= 0 && weekStart < 7
}
