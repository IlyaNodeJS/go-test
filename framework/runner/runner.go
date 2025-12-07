package runner

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/example/go-test-framework/framework/declarative"
	"github.com/example/go-test-framework/framework/executor"
	"github.com/example/go-test-framework/framework/suite"
	"github.com/example/go-test-framework/framework/utils"
)

// Runner coordinates suite execution.
type Runner struct {
	TestExecutor        *executor.TestExecutor
	DeclarativeExecutor *declarative.Executor
	Logger              *utils.StructuredLogger
}

func New(testExec *executor.TestExecutor, decl *declarative.Executor) *Runner {
	logger := utils.NewLogger()
	if testExec != nil && testExec.Logger == nil {
		testExec.Logger = logger
	}
	if decl != nil && decl.Logger == nil {
		decl.Logger = logger
	}
	return &Runner{TestExecutor: testExec, DeclarativeExecutor: decl, Logger: logger}
}

func (r *Runner) RunAll(ctx context.Context, suites []*suite.TestSuite) error {
	for _, ts := range suites {
		if err := r.RunSuite(ctx, ts); err != nil {
			return err
		}
	}
	return nil
}

func (r *Runner) RunSuite(ctx context.Context, ts *suite.TestSuite) error {
	if r.Logger == nil {
		r.Logger = utils.NewLogger()
	}
	log := r.Logger.With("suite", ts.ID)
	log.Info("starting suite", map[string]any{"services": ts.Services})

	var cancel context.CancelFunc
	if ts.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, ts.Timeout)
		defer cancel()
	}

	execCtx := utils.NewExecutionContext()

	run := func(def suite.TestDefinition) error {
		return r.runWithRetry(ctx, ts.Retries, func(attempt int) error {
			log.Info("running test", map[string]any{"service": def.Service, "attempt": attempt + 1})
			return r.TestExecutor.Run(ctx, def)
		})
	}

	runDeclarative := func(def declarative.DeclarativeTest) error {
		return r.runWithRetry(ctx, ts.Retries, func(attempt int) error {
			log.Info("running declarative test", map[string]any{"name": def.Name, "attempt": attempt + 1})
			return r.DeclarativeExecutor.Run(ctx, def, execCtx)
		})
	}

	switch ts.ExecutionType {
	case suite.ExecutionTypeParallel:
		return r.runParallel(len(ts.Tests)+len(ts.DeclarativeTests), func(idx int) error {
			if idx < len(ts.Tests) {
				return run(ts.Tests[idx])
			}
			return runDeclarative(ts.DeclarativeTests[idx-len(ts.Tests)])
		})
	default:
		for _, testDef := range ts.Tests {
			if err := run(testDef); err != nil {
				return err
			}
		}
		for _, decl := range ts.DeclarativeTests {
			if err := runDeclarative(decl); err != nil {
				return err
			}
		}
		log.Info("suite finished", nil)
		return nil
	}
}

func (r *Runner) runWithRetry(ctx context.Context, retries int, fn func(attempt int) error) error {
	var lastErr error
	attempts := retries + 1
	if attempts < 1 {
		attempts = 1
	}
	for attempt := 0; attempt < attempts; attempt++ {
		if err := fn(attempt); err != nil {
			lastErr = err
			select {
			case <-time.After(time.Duration(attempt+1) * 500 * time.Millisecond):
			case <-ctx.Done():
				return ctx.Err()
			}
			continue
		}
		return nil
	}
	if lastErr == nil {
		lastErr = errors.New("unknown retry failure")
	}
	return lastErr
}

func (r *Runner) runParallel(total int, fn func(idx int) error) error {
	var wg sync.WaitGroup
	errCh := make(chan error, total)
	for i := 0; i < total; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			if err := fn(idx); err != nil {
				errCh <- err
			}
		}(i)
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			return err
		}
	}
	return nil
}
