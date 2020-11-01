# Time Tracking API

<img align="right" width="128" src="https://user-images.githubusercontent.com/479339/74610686-49244e80-50aa-11ea-8a3d-dd4a11856d6c.png">

![build](https://github.com/BryanMorgan/time-tracking-api/workflows/build/badge.svg?branch=main&event=push)
[![Go Report Card](https://goreportcard.com/badge/github.com/BryanMorgan/time-tracking-api)](https://goreportcard.com/report/github.com/BryanMorgan/time-tracking-api)

Go API for tracking time.

Manages time entries for tasks that are associated with projects.

See the React [Time Tracking App](https://github.com/BryanMorgan/time-tracking-app) for a reference UI.

# Setup

## Database
Ensure you have PostgreSQL 12 or higher installed and running

Create a `timetracker` database and role using the bootstrap SQL in:

```./database/bootstrap.sql```

The create the schema in the `timetracker` database using:

```./database/schema-1.sql```

# Running
To run the API server you can run the Makefile target:

```make run```

which will start the dev server on the port configured in `config/dev.yml`

# Testing

## Unit Tests
Unit tests can be run using:

```make unit_test```

## Integration Tests
Integration tests are managed under the `integration_test` root folder and can be run using:

```make int_test```

## Postman Tests
Additional functional tests are available using he [Postman](https://www.postman.com/) tool.
These tests require the [newman](https://github.com/postmanlabs/newman) Postman command-line runner. Install using:

```npm install -g newman```

Also relies on the `database/bootstrap.sql` data to be present. To run the Postman tests locally, first start the web server:

```make run```

then run the Postman tests:

```make postman```

# API

### Authentication Token
Most of the REST endpoints require an authentication token which can be supplied as a custom header:

```Authorization: Bearer {token}```

Endpoints that do not require a token are noted below.

### Authentication

| Method | Path | Request | Response | Notes |
|--------|------|---------|----------|-------|
| POST | /api/auth/login | [LoginRequest](https://github.com/BryanMorgan/time-tracking-api/blob/34d9b71d7ce096280cb15f1e3be25c616e5044ad/profile/handler.go#L65) | [AuthResponse](https://github.com/BryanMorgan/time-tracking-api/blob/34d9b71d7ce096280cb15f1e3be25c616e5044ad/profile/handler.go#L79) | Does not require an authentication token |
| POST | /api/auth/token |  |  [AuthResponse](https://github.com/BryanMorgan/time-tracking-api/blob/34d9b71d7ce096280cb15f1e3be25c616e5044ad/profile/handler.go#L79) |  |
| POST | /api/auth/logout |  | `{}` | |
| POST | /api/auth/forgot | [EmailRequest](https://github.com/BryanMorgan/time-tracking-api/blob/34d9b71d7ce096280cb15f1e3be25c616e5044ad/profile/handler.go#L61) | `{}` | Does not require an authentication token. Sends a forgot password validation email. |

### Profile

| Method | Path | Request | Response | Notes |
|--------|------|---------|----------|-------|
| GET | /api/profile/ |  | [ProfileResponse](https://github.com/BryanMorgan/time-tracking-api/blob/34d9b71d7ce096280cb15f1e3be25c616e5044ad/profile/handler.go#L39) | |
| PUT | /api/profile/ | [ProfileRequest](https://github.com/BryanMorgan/time-tracking-api/blob/34d9b71d7ce096280cb15f1e3be25c616e5044ad/profile/handler.go#L25) | [ProfileResponse](https://github.com/BryanMorgan/time-tracking-api/blob/34d9b71d7ce096280cb15f1e3be25c616e5044ad/profile/handler.go#L39) | |
| PUT | /api/profile/password | [PasswordChangeRequest](https://github.com/BryanMorgan/time-tracking-api/blob/34d9b71d7ce096280cb15f1e3be25c616e5044ad/profile/handler.go#L33) | `{}` | |

### Account

| Method | Path | Request | Response | Notes |
|--------|------|---------|----------|-------|
| POST | /api/account |  [AccountRequest](https://github.com/BryanMorgan/time-tracking-api/blob/9b6d78799f7738a41bf955004fa6a0b8e5311da5/profile/handler.go#L53) | [ProfileResponse](https://github.com/BryanMorgan/time-tracking-api/blob/34d9b71d7ce096280cb15f1e3be25c616e5044ad/profile/handler.go#L39) | Does not require an authentication token |
| PUT | /api/account |  [AccountUpdateRequest](https://github.com/BryanMorgan/time-tracking-api/blob/c9d110f52882ede1544121abf9762bcc6451492c/profile/handler.go#L71) | [ProfileResponse](https://github.com/BryanMorgan/time-tracking-api/blob/34d9b71d7ce096280cb15f1e3be25c616e5044ad/profile/handler.go#L39) | |
| GET | /api/account |   | [AccountResponse](https://github.com/BryanMorgan/time-tracking-api/blob/c9d110f52882ede1544121abf9762bcc6451492c/profile/handler.go#L63) | |
| GET | /api/account/users |   | [][ProfileResponse](https://github.com/BryanMorgan/time-tracking-api/blob/c9d110f52882ede1544121abf9762bcc6451492c/profile/handler.go#L39) | |
| POST | /api/account/user |  [AddUserRequest](https://github.com/BryanMorgan/time-tracking-api/blob/c9d110f52882ede1544121abf9762bcc6451492c/profile/handler.go#L78) | [ProfileResponse](https://github.com/BryanMorgan/time-tracking-api/blob/34d9b71d7ce096280cb15f1e3be25c616e5044ad/profile/handler.go#L39) | |

### Client

| Method | Path | Request | Response | Notes |
|--------|------|---------|----------|-------|
| GET | /api/client/{client_id} | string | [ClientResponse](https://github.com/BryanMorgan/time-tracking-api/blob/c9d110f52882ede1544121abf9762bcc6451492c/client/handler.go#L25) | |
| GET | /api/client/all |   | [][ClientResponse](https://github.com/BryanMorgan/time-tracking-api/blob/c9d110f52882ede1544121abf9762bcc6451492c/client/handler.go#L25) | |
| GET | /api/client/archived|  |  [][ClientResponse](https://github.com/BryanMorgan/time-tracking-api/blob/c9d110f52882ede1544121abf9762bcc6451492c/client/handler.go#L25) | |
| POST | /api/client/ |  [ClientRequest](https://github.com/BryanMorgan/time-tracking-api/blob/c9d110f52882ede1544121abf9762bcc6451492c/client/handler.go#L19) | [ClientResponse](https://github.com/BryanMorgan/time-tracking-api/blob/c9d110f52882ede1544121abf9762bcc6451492c/client/handler.go#L25) | |
| PUT | /api/client/ |  [ClientRequest](https://github.com/BryanMorgan/time-tracking-api/blob/c9d110f52882ede1544121abf9762bcc6451492c/client/handler.go#L19) | `{}` | |
| DELETE | /api/client/ | [ClientRequest](https://github.com/BryanMorgan/time-tracking-api/blob/c9d110f52882ede1544121abf9762bcc6451492c/client/handler.go#L19) | `{}` | |
| PUT | /api/client/archive |  [ClientRequest](https://github.com/BryanMorgan/time-tracking-api/blob/c9d110f52882ede1544121abf9762bcc6451492c/client/handler.go#L19) | `{}` | |
| PUT | /api/client/restore |  [ClientRequest](https://github.com/BryanMorgan/time-tracking-api/blob/c9d110f52882ede1544121abf9762bcc6451492c/client/handler.go#L19) | `{}` | |

### Project

| Method | Path | Request | Response | Notes |
|--------|------|---------|----------|-------|
| GET | /api/project/{project_id} |  string  | [ProjectResponse](https://github.com/BryanMorgan/time-tracking-api/blob/c9d110f52882ede1544121abf9762bcc6451492c/client/handler.go#L62) | |
| GET | /api/project/all |   | [][ProjectResponse](https://github.com/BryanMorgan/time-tracking-api/blob/c9d110f52882ede1544121abf9762bcc6451492c/client/handler.go#L62) | |
| GET | /api/project/archived |   | [][ProjectResponse](https://github.com/BryanMorgan/time-tracking-api/blob/c9d110f52882ede1544121abf9762bcc6451492c/client/handler.go#L62) | |
| POST | /api/project/ |  [ProjectContainerRequest](https://github.com/BryanMorgan/time-tracking-api/blob/c9d110f52882ede1544121abf9762bcc6451492c/client/handler.go#L41) | [ProjectResponse](https://github.com/BryanMorgan/time-tracking-api/blob/c9d110f52882ede1544121abf9762bcc6451492c/client/handler.go#L62) | |
| PUT | /api/project/ |  [ProjectContainerRequest](https://github.com/BryanMorgan/time-tracking-api/blob/c9d110f52882ede1544121abf9762bcc6451492c/client/handler.go#L41) | `{}` | |
| DELETE | /api/project/ | [ProjectIdRequest](https://github.com/BryanMorgan/time-tracking-api/blob/c9d110f52882ede1544121abf9762bcc6451492c/client/handler.go#L31) | `{}` | |
| PUT | /api/project/archive | [ProjectIdRequest](https://github.com/BryanMorgan/time-tracking-api/blob/c9d110f52882ede1544121abf9762bcc6451492c/client/handler.go#L31) | `{}` | |
| PUT | /api/project/restore | [ProjectIdRequest](https://github.com/BryanMorgan/time-tracking-api/blob/c9d110f52882ede1544121abf9762bcc6451492c/client/handler.go#L31) | `{}` | |
| POST | /api/project/copy/last/week | [StartAndEndDateRequest](https://github.com/BryanMorgan/time-tracking-api/blob/c9d110f52882ede1544121abf9762bcc6451492c/client/handler.go#L72) | `{}` or [TimeRangeResponse](https://github.com/BryanMorgan/time-tracking-api/blob/main/timesheet/handler.go#L51) | Empty if no records the prior week|

### Time

| Method | Path | Request | Response | Notes |
|--------|------|---------|----------|-------|
| GET | /api/time/week |   | [TimeRangeResponse](https://github.com/BryanMorgan/time-tracking-api/blob/main/timesheet/handler.go#L51) | |
| GET | /api/time/week/{startDate} | string | [TimeRangeResponse](https://github.com/BryanMorgan/time-tracking-api/blob/main/timesheet/handler.go#L51) | Date must be in the `ISOShortDateFormat` (e.g. "2006-01-02") |
| PUT | /api/time/ | [TimeEntryRangeRequest](https://github.com/BryanMorgan/time-tracking-api/blob/main/timesheet/handler.go#L23) | `{}` | |
| POST | /api/time/project/week |  [ProjectWeekRequest](https://github.com/BryanMorgan/time-tracking-api/blob/main/timesheet/handler.go#L27) | `{}` | |
| DELETE | /api/time/project/week |  [ProjectDeleteRequest](https://github.com/BryanMorgan/time-tracking-api/blob/main/timesheet/handler.go#L34) | `{}` | |

### Task

| Method | Path | Request | Response | Notes |
|--------|------|---------|----------|-------|
| GET | /api/task/{taskId} | string | [TaskResponse](https://github.com/BryanMorgan/time-tracking-api/blob/c9d110f52882ede1544121abf9762bcc6451492c/task/handler.go#L23) | |
| GET | /api/task/all |  | [][TaskResponse](https://github.com/BryanMorgan/time-tracking-api/blob/c9d110f52882ede1544121abf9762bcc6451492c/task/handler.go#L23) | |]
| GET | /api/task/archived |  | [][TaskResponse](https://github.com/BryanMorgan/time-tracking-api/blob/c9d110f52882ede1544121abf9762bcc6451492c/task/handler.go#L23) | |]
| POST | /api/task/ | [TaskRequest](https://github.com/BryanMorgan/time-tracking-api/blob/c9d110f52882ede1544121abf9762bcc6451492c/task/handler.go#L23) | [TaskResponse](https://github.com/BryanMorgan/time-tracking-api/blob/c9d110f52882ede1544121abf9762bcc6451492c/task/handler.go#L23) | |
| PUT | /api/task/ |  [TaskRequest](https://github.com/BryanMorgan/time-tracking-api/blob/c9d110f52882ede1544121abf9762bcc6451492c/task/handler.go#L23) | `{}` | |
| PUT | /api/task/archive | [TaskRequest](https://github.com/BryanMorgan/time-tracking-api/blob/c9d110f52882ede1544121abf9762bcc6451492c/task/handler.go#L23) | `{}` | |
| PUT | /api/task/restore | [TaskRequest](https://github.com/BryanMorgan/time-tracking-api/blob/c9d110f52882ede1544121abf9762bcc6451492c/task/handler.go#L23) | `{}` | |
| DELETE | /api/task/ | [TaskRequest](https://github.com/BryanMorgan/time-tracking-api/blob/c9d110f52882ede1544121abf9762bcc6451492c/task/handler.go#L23) | `{}` | |

### Report

| Method | Path | Request | Response | Notes |
|--------|------|---------|----------|-------|
| GET | /api/report/time/client | query parameters: `from`, `to`, `page` | [][ClientReportResponse](https://github.com/BryanMorgan/time-tracking-api/blob/c9d110f52882ede1544121abf9762bcc6451492c/reporting/handler.go#L18) | `from` and `to` are date strings in the `ISOShortDateFormat` format |
| GET | /api/report/time/project | query parameters: `from`, `to`, `page` | [][ProjectReportResponse](https://github.com/BryanMorgan/time-tracking-api/blob/c9d110f52882ede1544121abf9762bcc6451492c/reporting/handler.go#L26) | `from` and `to` are date strings in the `ISOShortDateFormat` format |
| GET | /api/report/time/task | query parameters: `from`, `to`, `page` | [][TaskReportResponse](https://github.com/BryanMorgan/time-tracking-api/blob/c9d110f52882ede1544121abf9762bcc6451492c/reporting/handler.go#L35) | `from` and `to` are date strings in the `ISOShortDateFormat` format |
| GET | /api/report/time/person | query parameters: `from`, `to`, `page`| [][PersonReportResponse](https://github.com/BryanMorgan/time-tracking-api/blob/c9d110f52882ede1544121abf9762bcc6451492c/reporting/handler.go#L45) | `from` and `to` are date strings in the `ISOShortDateFormat` format |
| GET | /api/report/time/export/client | query parameters: `from`, `to` | CSV file with content type `text/csv` | `from` and `to` are date strings in the `ISOShortDateFormat` format |
| GET | /api/report/time/export/project | query parameters: `from`, `to` | CSV file with content type `text/csv` | `from` and `to` are date strings in the `ISOShortDateFormat` format |
| GET | /api/report/time/export/task | query parameters: `from`, `to` | CSV file with content type `text/csv` | `from` and `to` are date strings in the `ISOShortDateFormat` format |
| GET | /api/report/time/export/person | query parameters: `from`, `to` | CSV file with content type `text/csv` | `from` and `to` are date strings in the `ISOShortDateFormat` format|

### Ping

| Method | Path | Request | Response | Notes |
|--------|------|---------|----------|-------|
| GET | /_ping |   | Text `ok` with content type `text/plain` | |


## Errors
If an API fails, the HTTP status code will reflect the type of error response. Common error status codes are:

| Status Code | Error |
|-------------|-------|
| 400 | `http.StatusBadRequest` |
| 401 | `http.StatusUnauthorized` |
| 404 | `http.StatusNotFound` |
| 405 | `http.StatusMethodNotAllowed` |
| 500 | `http.StatusInternalServerError` |

Errors will produce a JSON response that contains a `status` field set to `error`.
Error details will be included as part of the serialized [Error](https://github.com/BryanMorgan/time-tracking-api/blob/34d9b71d7ce096280cb15f1e3be25c616e5044ad/api/error.go#L62) struct.

For example the error response below shows the JSON response for a 400 error when an `email` value is invalid:
```json
{
  "status": "error",
  "error": "invalid input",
  "message": "Invalid email",
  "code": "InvalidEmail",
  "detail": {"field": "email" }
}
```

