package executor

import (
	"context"
	"fmt"
	"time"

	"github.com/example/go-test-framework/framework/suite"
	"github.com/example/go-test-framework/framework/utils"
)

// TestExecutor is a placeholder for future service-specific logic. For now it
// logs that a test would run.
type TestExecutor struct {
	Logger *utils.StructuredLogger
}

func (te *TestExecutor) Run(ctx context.Context, definition suite.TestDefinition) error {
	if te.Logger == nil {
		te.Logger = utils.NewLogger()
	}
	te.Logger.Info("running test", map[string]any{"service": definition.Service, "type": definition.Type})
	select {
	case <-time.After(100 * time.Millisecond):
	case <-ctx.Done():
		return ctx.Err()
	}
	te.Logger.Info("finished test", map[string]any{"service": definition.Service})
	return nil
}

// BuildExecutor is a helper that could wire dependencies based on suite config.
func BuildExecutor() *TestExecutor {
	return &TestExecutor{Logger: utils.NewLogger()}
}
