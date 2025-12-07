package main

import (
	"context"
	"log"
	"path/filepath"

	_ "github.com/example/go-test-framework/suites"

	"github.com/example/go-test-framework/framework/declarative"
	"github.com/example/go-test-framework/framework/executor"
	httpclient "github.com/example/go-test-framework/framework/http"
	"github.com/example/go-test-framework/framework/runner"
	"github.com/example/go-test-framework/framework/suite"
	"github.com/example/go-test-framework/framework/utils"
)

func main() {
	ctx := context.Background()
	workDir := filepath.Join(".", "work")
	loader := suite.NewLoader(workDir)
	suites := loader.LoadTestSuites()

	resolver := httpclient.StaticResolver{
		"bonus-service":    "http://localhost:8081",
		"payments-service": "http://localhost:8082",
	}

	declExec := &declarative.Executor{
		HTTP:   httpclient.New(resolver),
		Logger: utils.NewLogger(),
		// Postgres and Mongo clients should be wired using env.EnvironmentConfig
		// details coming from suites when targeting a real environment.
	}

	run := runner.New(executor.BuildExecutor(), declExec)
	if err := run.RunAll(ctx, suites); err != nil {
		log.Fatalf("suite execution failed: %v", err)
	}
}
