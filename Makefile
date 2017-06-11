SHELL := /bin/bash

FILES ?= $(shell find . -type f -name '*.go')

.SUFFIXES:

.PHONY: \
	help \
	default \
	clean \
	tools \
	test \
	coverage \
	vet \
	errors \
	assignments \
	static \
	lint \
	imports \
	fmt \
	env \
	build \
	doc \
	version

all: imports fmt lint vet errors assignments static build

help:
	@echo 'Usage: make <OPTIONS> ... <TARGETS>'
	@echo ''
	@echo 'Available targets are:'
	@echo ''
	@echo '    help               Show this help screen.'
	@echo '    clean              Remove binaries, artifacts and releases.'
	@echo '    tools              Install tools needed by the project.'
	@echo '    test               Run unit tests.'
	@echo '    coverage           Report code tests coverage.'
	@echo '    vet                Run go vet.'
	@echo '    errors             Run errcheck.'
	@echo '    assignments        Run ineffassign.'
	@echo '    static             Run staticcheck.'
	@echo '    lint               Run golint.'
	@echo '    imports            Run goimports.'
	@echo '    fmt                Run go fmt.'
	@echo '    env                Display Go environment.'
	@echo '    build              Build project for current platform.'
	@echo '    doc                Start Go documentation server on port 8080.'
	@echo '    version            Display Go version.'
	@echo ''
	@echo 'Targets run by default are: imports, fmt, lint, vet, errors, assignments and build.'
	@echo ''

print-%:
	@echo $* = $($*)

clean:
	go clean -i ./...
	rm -f $(CURDIR)/coverage.*

tools:
	go get github.com/axw/gocov/gocov
	go get github.com/golang/lint/golint
	go get github.com/gordonklaus/ineffassign
	go get github.com/kisielk/errcheck
	go get github.com/matm/gocov-html
	go get golang.org/x/tools/cmd/goimports
	go get honnef.co/go/tools/cmd/staticcheck

test:
	go test -v ./...

coverage:
	gocov test ./... > $(CURDIR)/coverage.out 2>/dev/null
	gocov report $(CURDIR)/coverage.out
	if [[ -z "$$CI" ]]; then \
		gocov-html $(CURDIR)/coverage.out > $(CURDIR)/coverage.html; \
	  	if which open &>/dev/null; then \
	    		open $(CURDIR)/coverage.html; \
		fi; \
	fi

vet:
	$(eval QUIET := $(shell test "$(MAKECMDGOALS)" == "vet" || echo 1))
	@go vet -v ./... $(shell test -z "$(QUIET)" || echo '&>/dev/null'); \
	if (( $$? > 0 )); then \
		if [[ -n "$(QUIET)" ]]; then \
			echo "go vet found number of issues. Run 'make vet' to check directly."; \
		else \
			exit $$?; \
		fi; \
	fi

errors:
	$(eval QUIET := $(shell test "$(MAKECMDGOALS)" == "errors" || echo 1))
	@errcheck -ignoretests -blank ./... $(shell test -z "$(QUIET)" || echo '&>/dev/null'); \
	if (( $$? > 0 )); then \
		if [[ -n "$(QUIET)" ]]; then \
			echo "errcheck found number of issues. Run 'make errors' to check directly."; \
		else \
			exit $$?; \
		fi; \
	fi

assignments:
	$(eval QUIET := $(shell test "$(MAKECMDGOALS)" == "assignments" || echo 1))
	@ineffassign . $(shell test -z "$(QUIET)" || echo '&>/dev/null'); \
	if (( $$? > 0 )); then \
		if [[ -n "$(QUIET)" ]]; then \
			echo "ineffassign found number of issues. Run 'make assignments' to check directly."; \
		else \
			exit $$?; \
		fi; \
	fi

static:
	$(eval QUIET := $(shell test "$(MAKECMDGOALS)" == "static" || echo 1))
	@staticcheck ./... $(shell test -z "$(QUIET)" || echo '&>/dev/null'); \
	if (( $$? > 0 )); then \
		if [[ -n "$(QUIET)" ]]; then \
			echo "staticcheck found number of issues. Run 'make static' to check directly."; \
		else \
			exit $$?; \
		fi; \
	fi

lint:
	$(eval QUIET := $(shell test "$(MAKECMDGOALS)" == "lint" || echo 1))
	@golint ./... $(shell test -z "$(QUIET)" || echo '&>/dev/null'); \
	if (( $$? > 0 )); then \
		if [[ -n "$(QUIET)" ]]; then \
			echo "golint found number of issues. Run 'make lint' to check directly."; \
		else \
			exit $$?; \
		fi; \
	fi

imports:
	$(eval QUIET := $(shell test "$(MAKECMDGOALS)" == "imports" || echo 1))
	@goimports -l $(FILES) $(shell test -z "$(QUIET)" || echo '&>/dev/null'); \
	if (( $$? > 0 )); then \
		if [[ -n "$(QUIET)" ]]; then \
			echo "goimports found number of issues. Run 'make imports' to check directly."; \
		else \
			exit $$?; \
		fi; \
	fi

fmt:
	$(eval QUIET := $(shell test "$(MAKECMDGOALS)" == "fmt" || echo 1))
	@gofmt -l $(FILES) $(shell test -z "$(QUIET)" || echo '&>/dev/null'); \
	if (( $$? > 0 )); then \
		if [[ -n "$(QUIET)" ]]; then \
			echo "gofmt found number of issues. Run 'make fmt' to check directly."; \
		else \
			exit $$?; \
		fi; \
	fi

env:
	@go env

build:
	go build ./...

doc:
	godoc -http=:8080 -index

version:
	@go version
