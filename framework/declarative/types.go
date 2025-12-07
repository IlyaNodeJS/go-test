package declarative

import "time"

// DeclarativeTest models YAML/JSON driven integration flows.
type DeclarativeTest struct {
	Name               string              `json:"name"`
	Description        string              `json:"description"`
	Action             Action              `json:"action"`
	ResponseAssertions *ResponseAssertions `json:"responseAssertions"`
	Assertions         []Assertion         `json:"assertions"`
	DelayAfter         time.Duration       `json:"delayAfter"`
}

// Action describes the HTTP call to perform.
type Action struct {
	Service  string            `json:"service"`
	Endpoint string            `json:"endpoint"`
	Method   string            `json:"method"`
	Body     map[string]any    `json:"body"`
	Headers  map[string]string `json:"headers"`
	Extract  map[string]string `json:"extract"`
}

// ResponseAssertions holds HTTP validations.
type ResponseAssertions struct {
	Status int             `json:"status"`
	Body   *BodyAssertions `json:"body"`
}

// BodyAssertions supports contains checks.
type BodyAssertions struct {
	Contains map[string]any `json:"contains"`
}

// Assertion defines a DB validation.
type Assertion struct {
	Database     string         `json:"database"`
	DatabaseName string         `json:"databaseName"`
	Collection   string         `json:"collection"`
	Schema       string         `json:"schema"`
	Table        string         `json:"table"`
	Query        map[string]any `json:"query"`
	Expected     ExpectedResult `json:"expected"`
}

type ExpectedResult struct {
	Count    *int           `json:"count"`
	Contains map[string]any `json:"contains"`
}
