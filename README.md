# Go Test Framework

This repository contains a self-contained Go testing framework that discovers microservices under `./work`, loads registered suites, and executes both classic integration tests and declarative HTTP/database flows.

## Layout

- `framework/` – core SDK with suite models, runner, http/db clients, declarative executor, env/config helpers, logging and variable substitution utilities.
- `suites/` – suite definitions that auto-register via `init` (see `suites/deposit_suite.go`).
- `cmd/runner/` – CLI entry wiring the loader, HTTP resolver, and runner.
- `work/` – placeholder microservices;

## Features

- Suite loader scans `/work`, injects service lists and exposes environment config.
- Runner (`framework/runner/runner.go`) supports sequential/parallel execution, retries, global suite timeouts, shared variable context, and structured logs.
- Declarative executor performs HTTP actions, extracts variables (`${var}`), delays, and asserts against Postgres + Mongo via lightweight clients.
- HTTP client builds URLs from service names, handles JSON payloads, validates responses.
- DB helpers implement simple query builders plus `count`/`contains` validations.

## Running

1. Adjust `cmd/runner/main.go` with real service base URLs and database DSNs if available.
2. Execute `go run ./cmd/runner` to load suites and run tests.
3. Once the Go build cache is writable in your environment, `go test ./...` will also exercise the modules.

This skeleton focuses on the framework; no microservice business logic is included. Expand the suite set and wire real dependencies to tailor it to your environment.
