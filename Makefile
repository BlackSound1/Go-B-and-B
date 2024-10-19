SHELL := /bin/bash
.ONESHELL:
.DEFAULT_GOAL := help


.PHONY: test
test:   ## Run tests
	@cd ./cmd/web && go test -v && cd ../../


.PHONY: cov-report
cov-report:  ## Create full coverage report and open in browser
	@cd ./cmd/web && go test -coverprofile=coverage.out && go tool cover -html=coverage.out && cd ../../


.PHONY: cov
cov:  ## Create short coverage report
	@cd ./cmd/web && go test -cover && cd ../../


.PHONY: run
run:   ## Run web server
	@go run ./cmd/web


.PHONY: help
help:   ## Show this help
	@echo -e "\nCommands:\n"
	@egrep '^[a-zA-Z_-]+:.*?## .*' Makefile | sort |
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
