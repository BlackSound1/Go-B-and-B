SHELL := /bin/bash
.ONESHELL:
.DEFAULT_GOAL := help


.PHONY: test
test:   ## Run tests
	@go test -v ./cmd/web ./internal/handlers/


.PHONY: cov-report
cov-report:  ## Create full coverage report and open in browser
	@go test -coverprofile=coverage.out ./cmd/web/ ./internal/handlers/ && go tool cover -html=coverage.out


.PHONY: cov
cov:  ## Create short coverage report
	@go test -cover ./cmd/web/ ./internal/handlers/


.PHONY: run
run:   ## Run web server
	@go run ./cmd/web


.PHONY: help
help:   ## Show this help
	@echo -e "\nCommands:\n"
	@egrep '^[a-zA-Z_-]+:.*?## .*' Makefile | sort |
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
