# Miya Engine Makefile

.PHONY: test test-verbose test-coverage test-race bench fmt vet clean help

# Default target
all: fmt vet test

# Run all tests (excluding examples)
test:
	go test $$(go list ./... | grep -v /examples/)

# Run tests with verbose output
test-verbose:
	go test -v $$(go list ./... | grep -v /examples/)

# Run tests with coverage
test-coverage:
	go test -cover $$(go list ./... | grep -v /examples/)

# Generate coverage report
coverage-report:
	go test -coverprofile=coverage.out $$(go list ./... | grep -v /examples/)
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run tests with race detector
test-race:
	go test -race $$(go list ./... | grep -v /examples/)

# Run specific package tests
test-filters:
	go test -v ./filters

test-parser:
	go test -v ./parser

test-runtime:
	go test -v ./runtime

test-lexer:
	go test -v ./lexer

test-loader:
	go test -v ./loader

test-unit:
	go test -v ./tests/unit

# Run benchmarks
bench:
	go test -bench=. $$(go list ./... | grep -v /examples/)

# Run standalone Miya benchmark
bench-miya:
	go run benchmarks/benchmark_miya.go

# Run Python Jinja2 benchmark (requires python3 and jinja2)
bench-python:
	python3 benchmarks/benchmark_python.py

# Run both for comparison
bench-compare: bench-miya bench-python

bench-filters:
	go test -bench=. ./filters

bench-parser:
	go test -bench=. ./parser

# Format code
fmt:
	gofmt -w .

# Run static analysis
vet:
	go vet $$(go list ./... | grep -v /examples/)

# Clean build artifacts and test cache
clean:
	go clean -testcache
	rm -f coverage.out coverage.html

# Run examples (tutorial)
run-tutorial:
	@echo "Running tutorial examples..."
	@cd examples/tutorial && go run step1_hello_world.go
	@cd examples/tutorial && go run step2_variables_filters.go
	@cd examples/tutorial && go run step3_control_flow.go
	@cd examples/tutorial && go run step4_inheritance.go
	@cd examples/tutorial && go run step5_macros.go
	@cd examples/tutorial && go run step6_filesystem.go

# Run showcase demo
run-showcase:
	@cd examples/showcase && go run demo.go

# Help
help:
	@echo "Miya Engine - Available targets:"
	@echo ""
	@echo "  make test            Run all tests"
	@echo "  make test-verbose    Run tests with verbose output"
	@echo "  make test-coverage   Run tests with coverage summary"
	@echo "  make coverage-report Generate HTML coverage report"
	@echo "  make test-race       Run tests with race detector"
	@echo "  make bench           Run all Go test benchmarks"
	@echo "  make bench-miya      Run standalone Miya benchmark"
	@echo "  make bench-python    Run Python Jinja2 benchmark"
	@echo "  make bench-compare   Run Miya vs Python comparison"
	@echo "  make fmt             Format code with gofmt"
	@echo "  make vet             Run static analysis"
	@echo "  make clean           Clean test cache and coverage files"
	@echo "  make run-tutorial    Run all tutorial examples"
	@echo "  make run-showcase    Run showcase demo"
	@echo ""
	@echo "  make test-filters    Run filter tests only"
	@echo "  make test-parser     Run parser tests only"
	@echo "  make test-runtime    Run runtime tests only"
	@echo "  make test-lexer      Run lexer tests only"
	@echo "  make test-loader     Run loader tests only"
	@echo "  make test-unit       Run unit tests only"
