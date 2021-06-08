// +build integration

package integration_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"github.com/bryanmorgan/time-tracking-api/config"
	"github.com/go-chi/chi"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/bryanmorgan/time-tracking-api/app"
	"github.com/bryanmorgan/time-tracking-api/profile"
)

var db *sqlx.DB
var router *chi.Mux

var testPasswordEncrypted string

type jsonResult struct {
	Status string
	Code   string
	Data   json.RawMessage
}

const (
	TestEmail               = "Unit.Test@example.com"
	TestPassword            = "=G)A-B8XAD=rP~Kp$&8n$/dmeGpipY,L3]C!6x9"
	TestFirstName           = "John"
	TestLastName            = "Smith"
	TestCompany             = "ACME Unit Testing Inc."
	TestCompany2            = "Example Unit Testing LLC"
	TestHost                = "test.example.com"
	TestSessionType         = "web"
	TestToken               = "8NsbqMkL5aCYH9mPed8zk6kyQKEbsdXso5qckr9DSEdHoaqkf585kHD4EhbKMdSAqFiMbLbjryJAYELrMgEcXtr8HeCHTFSXpptRPtA7nFsigcodQzTtbHCL"
	TestForgotPasswordToken = "9gPebFemAb3egNsjkPKti8eodCYa4nyNpMgEcXtr8HeCHTFSXpptRPtA7nFsigcodQzTtbHCLJ3oEjafq8odBtJr7dQKHEqtrrTePbxmPAxoXdnLrkGXg4xq"
	TestClientName          = "Build Something Amazing, Inc."
	TestProjectName         = "Launch App"
	TestClientAddress       = "1200 Main Street \nSan Diego, CA 92121"
)


func TestMain(m *testing.M) {
	os.Setenv("GO_ENV", "test")

	viper.AddConfigPath("../config")
	server := app.NewApp()
	router = server.Router
	db = server.DB

	setupVariables()
	code := m.Run()

	db.Close()
	os.Exit(code)
}

func AddAuthorizationHeaders(r *http.Request) {
	AddRequestHeaders(r)
	r.Header.Add("Authorization", "Bearer "+TestToken)
}

func AddRequestHeaders(r *http.Request) {
	r.Header.Add("Content-Type", "application/json")
	r.Host = TestHost
}

func setupVariables() {
	encryptedPassword, err := bcrypt.GenerateFromPassword([]byte(TestPassword), 8) // Use 8 for integration test speed
	if err != nil {
		log.Fatalf("Could not encrypt test password: %v", err.Error())
	}
	testPasswordEncrypted = string(encryptedPassword)
}

func encodeJson(t *testing.T, input *map[string]interface{}) *bytes.Buffer {
	t.Helper()
	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(input)
	if err != nil {
		t.Fatalf("Could not encode to JSON: %s", err)
	}

	return body
}

func tearDownEnvironment() {
	db.Close()
}

// Returns the profileId, accountId of a new test account
func createDefaultUnitTestAccount() (int, int) {
	return createUnitTestAccount(TestEmail, TestFirstName, TestLastName, TestCompany, TestCompany2, profile.Admin)
}

func createUnitTestAccount(email string, firstName string, lastName string, company string, company2, role profile.AuthorizationRole) (int, int) {
	// Convert emails to lower case to simulate what the real services do
	email = strings.ToLower(email)

	tx, err := db.Begin()
	if err != nil {
		log.Panicf("Transaction failed in create unit test account [%s]", err.Error())
		return 0, 0
	}

	// Create main account
	var accountId int
	accountSql := `INSERT INTO account (company) VALUES ($1) RETURNING account_id`
	err = tx.QueryRow(accountSql, company).Scan(&accountId)
	if err != nil {
		tx.Rollback()
		log.Panicf("Transaction rolled back in create unit test account [%s]", err.Error())
		return 0, 0
	}

	// Create a 2nd account for this profile
	var accountId2 int
	err = tx.QueryRow(accountSql, company2).Scan(&accountId2)
	if err != nil {
		tx.Rollback()
		log.Panicf("Transaction rolled back in create unit test account2 [%s]", err.Error())
		return 0, 0
	}

	var userId int
	profileSql := `INSERT INTO profile (email, password, first_name, last_name, profile_status, forgot_password_token, forgot_password_expiration) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING profile_id`
	err = tx.QueryRow(profileSql, email, testPasswordEncrypted, firstName, lastName, profile.ProfileValid, TestForgotPasswordToken, "2050-12-31").Scan(&userId)
	if err != nil {
		tx.Rollback()
		log.Panicf("Transaction rolled back in create unit test account [%s]", err.Error())
		return 0, 0
	}

	profileAccountSql := `INSERT INTO profile_account (profile_id, account_id, role, profile_account_status) VALUES ($1, $2, $3, $4)`
	_, err = tx.Exec(profileAccountSql, userId, accountId, role, profile.ProfileAccountValid)
	if err != nil {
		tx.Rollback()
		log.Panicf("Transaction rolled back in create unit test account [%s]", err.Error())
		return 0, 0
	}

	// Associate the profile to a second account
	_, err = tx.Exec(profileAccountSql, userId, accountId2, role, profile.ProfileNew)
	if err != nil {
		tx.Rollback()
		log.Panicf("Transaction rolled back in create unit test account2 [%s]", err.Error())
		return 0, 0
	}

	sessionSql := `INSERT INTO session (profile_id, account_id, token, token_expiration, type) VALUES ($1, $2, $3, $4, $5)`
	_, err = tx.Exec(sessionSql, userId, accountId, TestToken, time.Now().Add(time.Minute*time.Duration(5)), TestSessionType)
	if err != nil {
		tx.Rollback()
		log.Panicf("Transaction rolled back in test account [%s]", err.Error())
		return 0, 0
	}

	if err = tx.Commit(); err != nil {
		log.Panicf("Transaction commit failed in create unit test account [%s]", err.Error())
		return 0, 0
	}

	return userId, accountId
}

func deleteDefaultUnitTestAccount() {
	deleteUnitTestProfileByEmail(TestEmail)
}

func deleteUnitTestProfileByEmail(email string) {
	var userId int
	emails := []string{email, strings.ToLower(email)}
	for _, emailToDelete := range emails {
		err := db.Get(&userId, "SELECT profile_id FROM profile WHERE email = $1", emailToDelete)
		if err == sql.ErrNoRows {
			continue
		}

		if err != nil {
			log.Panicf("Failed to get unit test profile for email [%s]. Err: [%s]", emailToDelete, err.Error())
			return
		}

		_, err = db.Exec("DELETE FROM profile WHERE profile_id = $1", userId)
		if err != nil {
			log.Panicf("Failed to delete unit test profile [%s]", err.Error())
			return
		}

		_, err = db.Exec("DELETE FROM profile_account WHERE profile_id = $1", userId)
		if err != nil {
			log.Panicf("Failed to delete unit test profile account [%s]", err.Error())
			return
		}

		_, err = db.Exec("DELETE FROM session WHERE profile_id = $1", userId)
		if err != nil {
			log.Panicf("Failed to delete unit test session [%s]", err.Error())
			return
		}

		_, err = db.Exec("DELETE FROM login_attempts WHERE email  = $1", emailToDelete)
		if err != nil {
			log.Panicf("Failed to delete unit test login attempts [%s]", err.Error())
			return
		}
	}
}

func createTestClient(accountId int, name string, address string) int {
	tx, err := db.Begin()
	if err != nil {
		log.Panicf("Transaction failed in create test client [%s]", err)
		return 0
	}

	// Create test client
	var clientId int
	accountSql := `INSERT INTO client (account_id, client_name, address) VALUES ($1, $2, $3) RETURNING client_id`
	err = tx.QueryRow(accountSql, accountId, name, address).Scan(&clientId)
	if err != nil {
		tx.Rollback()
		log.Panicf("Transaction rolled back in create test client [%s]", err)
		return 0
	}
	if err = tx.Commit(); err != nil {
		log.Panicf("Transaction commit failed in create test client [%s]", err)
		return 0
	}

	return clientId
}

func deleteTestClient(clientId int) {
	_, err := db.Exec("DELETE FROM client WHERE client_id = $1", clientId)
	if err != nil {
		log.Panicf("Failed to delete client by id [%d]: [%s]", clientId, err)
		return
	}
}

func createTestProject(accountId int, clientId int, name string) int {
	tx, err := db.Begin()
	if err != nil {
		log.Panicf("Transaction failed in create test project [%s]", err)
		return 0
	}

	// Create test project
	var projectId int
	accountSql := `INSERT INTO project (account_id, client_id, project_name) VALUES ($1, $2, $3) RETURNING project_id`
	err = tx.QueryRow(accountSql, accountId, clientId, name).Scan(&projectId)
	if err != nil {
		tx.Rollback()
		log.Panicf("Transaction rolled back in create test project [%s]", err)
		return 0
	}
	if err = tx.Commit(); err != nil {
		log.Panicf("Transaction commit failed in create test project [%s]", err)
		return 0
	}

	return projectId
}

func createTestTimeEntries(startString string, days int, accountId int, profileId int, projectId int, taskId int) {
	tx, err := db.Begin()
	if err != nil {
		log.Panicf("Transaction failed: [%s]", err)
		return
	}

	currentDate, err := time.Parse(config.ISOShortDateFormat, startString)
	startDate := currentDate
	if err != nil {
		tx.Rollback()
		log.Panicf("Invalid start date: [%s]: error: [%s]", startString, err)
		return
	}
	for i := 0; i < days; i++ {
		startDate = currentDate.AddDate(0, 0, i)
		sql := `INSERT INTO time (account_id, profile_id, project_id, task_id, day, hours) VALUES ($1, $2, $3, $4, $5, $6)`
		result, err := tx.Exec(sql, accountId, profileId, projectId, taskId, startDate.Format(config.ISOShortDateFormat), rand.Float64()*8)
		if err != nil {
			tx.Rollback()
			log.Panicf("Transaction rolled back: [%s]", err)
			return
		}

		count, err := result.RowsAffected()
		if err != nil {
			tx.Rollback()
			log.Panicf("Transaction rolled back: [%s]", err)
			return
		}

		if count == 0 {
			tx.Rollback()
			log.Panicf("Transaction rolled back: no rows inserted")
			return
		}
	}

	if err = tx.Commit(); err != nil {
		log.Panicf("Transaction commit failed: [%s]", err)
		return
	}

	return
}

func createTestTask(accountId int) int {
	tx, err := db.Begin()
	if err != nil {
		log.Panicf("Transaction failed: [%s]", err)
		return 0
	}

	// Create test task
	var taskId int
	sql := `INSERT INTO task (account_id, task_name) VALUES ($1, $2) RETURNING task_id`
	err = tx.QueryRow(sql, accountId, "Test Task").Scan(&taskId)
	if err != nil {
		tx.Rollback()
		log.Panicf("Transaction rolled back: [%s]", err)
		return 0
	}

	if err = tx.Commit(); err != nil {
		log.Panicf("Transaction commit failed: [%s]", err)
		return 0
	}

	return taskId
}

func deleteTestTimeEntries(accountId int, profileId int, projectId int) {
	_, err := db.Exec("DELETE FROM time WHERE account_id=$1 AND profile_id=$2 AND project_id=$3", accountId, profileId, projectId)
	if err != nil {
		log.Panicf("Failed to delete time for entry: [%d] [%d] [%d]: [%s]", accountId, profileId, projectId, err)
		return
	}
}

func deleteTestProject(projectId int) {
	_, err := db.Exec("DELETE FROM project WHERE project_id = $1", projectId)
	if err != nil {
		log.Panicf("Failed to delete client by id [%d]: [%s]", projectId, err)
		return
	}
}

func deleteTestTask(taskId int, accountId int) {
	_, err := db.Exec("DELETE FROM task WHERE task_id=$1 AND account_id=$2", taskId, accountId)
	if err != nil {
		log.Panicf("Failed to delete client by id [%d]: [%s]", taskId, err)
		return
	}
}


