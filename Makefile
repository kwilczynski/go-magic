SHELL := /bin/bash
VERBOSE := $(or $(VERBOSE),$(V))

.SUFFIXES:

.PHONY: \
	help \
	default \
	clean \
	tools \
	test \
	lint \
	coverage \
	env \
	build \
	doc \
	version

ifneq ($(VERBOSE), 1)
.SILENT:
endif

default: all

all: lint build

help: ## Show this help screen.
	@echo 'Usage: make <OPTIONS> ... <TARGETS>'
	@echo ''
	@echo 'Available targets are:'
	@echo ''
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN { FS = ":.*?## " }; { printf "%-30s %s\n", $$1, $$2 }'
	@echo ''
	@echo 'Targets run by default are lint and build.'
	@echo ''

print-%:
	@echo $* = $($*)

clean: ## Remove binaries, artifacts and releases.
	go clean -i ./...
	rm -f $(CURDIR)/coverage.*

tools: ## Install tools needed by the project.
	GO111MODULE=off go get github.com/alecthomas/gometalinter
	GO111MODULE=off go get github.com/axw/gocov/gocov
	GO111MODULE=off go get github.com/matm/gocov-html
	GO111MODULE=off GOPATH=$(shell go env GOPATH) gometalinter --install

test: ## Run unit tests.
	go test -v ./...

lint: ## Run lint tests suite.
	$(eval QUIET := $(shell test "$(MAKECMDGOALS)" == "lint" || echo 1))
	gometalinter ./... $(shell test -z "$(QUIET)" || echo '&>/dev/null'); \
	if (( $$? > 0 )); then \
		if [[ -n "$(QUIET)" ]]; then \
			echo "Found number of issues when running lint tests suite. Run 'make lint' to check directly."; \
		else \
			test -z "$(VERBOSE)" || exit $$?; \
		fi; \
	fi

coverage: ## Report code tests coverage.
	gocov test ./... > $(CURDIR)/coverage.out 2>/dev/null
	gocov report $(CURDIR)/coverage.out
	if [[ -z "$$CI" ]]; then \
		gocov-html $(CURDIR)/coverage.out > $(CURDIR)/coverage.html; \
	  	if which open &>/dev/null; then \
	    		open $(CURDIR)/coverage.html; \
		fi; \
	fi

env: ## Display Go environment.
	@go env

build: ## Build project for current platform.
	go build ./...

doc: ## Start Go documentation server on port 8080.
	godoc -http=:8080 -index

version: ## Display Go version.
	@go version
