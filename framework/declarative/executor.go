package declarative

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/example/go-test-framework/framework/db/mongo"
	"github.com/example/go-test-framework/framework/db/postgres"
	httpclient "github.com/example/go-test-framework/framework/http"
	"github.com/example/go-test-framework/framework/utils"
)

// Executor orchestrates declarative tests end-to-end.
type Executor struct {
	HTTP     *httpclient.Client
	Postgres *postgres.Client
	Mongo    *mongo.Client
	Logger   *utils.StructuredLogger
}

func (e *Executor) Run(ctx context.Context, test DeclarativeTest, execCtx *utils.ExecutionContext) error {
	if e.HTTP == nil {
		return errors.New("http client must be configured")
	}
	if e.Logger == nil {
		e.Logger = utils.NewLogger()
	}

	log := e.Logger.With("test", test.Name)
	log.Info("starting declarative test", map[string]any{"description": test.Description})

	if err := e.executeAction(ctx, test, execCtx, log); err != nil {
		log.Error("action failed", map[string]any{"error": err.Error()})
		return err
	}

	if test.DelayAfter > 0 {
		log.Info("sleeping after action", map[string]any{"delay": test.DelayAfter.String()})
		select {
		case <-time.After(test.DelayAfter):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	for _, assertion := range test.Assertions {
		if err := e.executeAssertion(ctx, assertion, execCtx, log); err != nil {
			return err
		}
	}

	log.Info("declarative test finished", nil)
	return nil
}

func (e *Executor) executeAction(ctx context.Context, test DeclarativeTest, execCtx *utils.ExecutionContext, log *utils.StructuredLogger) error {
	requestBody := map[string]any{}
	if test.Action.Body != nil {
		requestBody = utils.CloneMap(test.Action.Body)
		if substituted, ok := utils.Substitute(requestBody, execCtx.Snapshot()).(map[string]any); ok {
			requestBody = substituted
		}
	}
	headers := map[string]string{}
	for k, v := range test.Action.Headers {
		headers[k] = v
	}

	resp, err := e.HTTP.Do(ctx, httpclient.Request{
		Service:  test.Action.Service,
		Method:   strings.ToUpper(test.Action.Method),
		Endpoint: test.Action.Endpoint,
		Body:     requestBody,
		Headers:  headers,
	})
	if err != nil {
		return err
	}

	if test.ResponseAssertions != nil {
		if err := validateResponse(resp, test.ResponseAssertions); err != nil {
			return err
		}
	}

	for varName, path := range test.Action.Extract {
		if value, ok := resp.Body[path]; ok {
			execCtx.Set(varName, fmt.Sprint(value))
			log.Info("extracted variable", map[string]any{"key": varName, "value": value})
		}
	}
	return nil
}

func validateResponse(resp *httpclient.Response, expectations *ResponseAssertions) error {
	if resp.StatusCode != expectations.Status {
		return fmt.Errorf("unexpected status: got %d want %d", resp.StatusCode, expectations.Status)
	}
	if expectations.Body != nil && len(expectations.Body.Contains) > 0 {
		for key, val := range expectations.Body.Contains {
			if actual, ok := resp.Body[key]; !ok || fmt.Sprint(actual) != fmt.Sprint(val) {
				return fmt.Errorf("response body missing %s=%v", key, val)
			}
		}
	}
	return nil
}

func (e *Executor) executeAssertion(ctx context.Context, assertion Assertion, execCtx *utils.ExecutionContext, log *utils.StructuredLogger) error {
	query := map[string]any{}
	if assertion.Query != nil {
		substituted := utils.Substitute(assertion.Query, execCtx.Snapshot())
		if q, ok := substituted.(map[string]any); ok {
			query = q
		}
	}

	switch strings.ToLower(assertion.Database) {
	case "mongodb", "mongo":
		if e.Mongo == nil {
			return errors.New("mongo client is not configured")
		}
		dbName := assertion.DatabaseName
		if dbName == "" {
			dbName = assertion.Schema
		}
		if dbName == "" {
			return errors.New("mongo databaseName is required")
		}
		if assertion.Expected.Count != nil {
			if err := e.Mongo.ValidateCount(ctx, dbName, assertion.Collection, query, int64(*assertion.Expected.Count)); err != nil {
				return err
			}
		}
		if len(assertion.Expected.Contains) > 0 {
			if err := e.Mongo.ValidateContains(ctx, dbName, assertion.Collection, query, assertion.Expected.Contains); err != nil {
				return err
			}
		}
	case "postgres", "postgresql":
		if e.Postgres == nil {
			return errors.New("postgres client is not configured")
		}
		if assertion.Expected.Count != nil {
			if err := e.Postgres.ValidateCount(ctx, assertion.Schema, assertion.Table, query, *assertion.Expected.Count); err != nil {
				return err
			}
		}
		if len(assertion.Expected.Contains) > 0 {
			if err := e.Postgres.ValidateContains(ctx, assertion.Schema, assertion.Table, query, assertion.Expected.Contains); err != nil {
				return err
			}
		}
	default:
		log.Info("skipping unsupported assertion target", map[string]any{"database": assertion.Database})
	}
	return nil
}
