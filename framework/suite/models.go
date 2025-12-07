package suite

import (
	"time"

	"github.com/example/go-test-framework/framework/declarative"
	"github.com/example/go-test-framework/framework/env"
)

// ExecutionType controls sequential vs parallel behavior.
type ExecutionType string

const (
	ExecutionTypeSequential ExecutionType = "sequential"
	ExecutionTypeParallel   ExecutionType = "parallel"
)

// TestDefinition represents a classic Go/integration test entry point.
type TestDefinition struct {
	Service string `json:"service"`
	Type    string `json:"type"`
}

// TestSuite describes infra requirements, tests, retries, etc.
type TestSuite struct {
	ID               string                        `json:"id"`
	Name             string                        `json:"name"`
	Services         []string                      `json:"services"`
	Dependencies     []string                      `json:"dependencies"`
	ExecutionType    ExecutionType                 `json:"executionType"`
	Timeout          time.Duration                 `json:"timeout"`
	Retries          int                           `json:"retries"`
	Tests            []TestDefinition              `json:"tests"`
	DeclarativeTests []declarative.DeclarativeTest `json:"declarativeTests"`
	Environment      env.EnvironmentConfig         `json:"environment"`
	Config           map[string]any                `json:"config"`
}
