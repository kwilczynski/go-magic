SHELL := /bin/bash

.PHONY: \
	help \
	default \
	clean \
	tools \
	test \
	coverage \
	vet \
	errors \
	static \
	lint \
	imports \
	fmt \
	env \
	build \
	doc \
	version

all: imports fmt lint vet errors static build

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
	@echo '    static             Run staticcheck.'
	@echo '    lint               Run golint.'
	@echo '    imports            Run goimports.'
	@echo '    fmt                Run go fmt.'
	@echo '    env                Display Go environment.'
	@echo '    build              Build project for current platform.'
	@echo '    doc                Start Go documentation server on port 8080.'
	@echo '    version            Display Go version.'
	@echo ''
	@echo 'Targets run by default are: imports, fmt, lint, vet, errors and build.'
	@echo ''

print-%:
	@echo $* = $($*)

clean:
	go clean -i ./...

tools:
	go get github.com/axw/gocov/gocov
	go get github.com/golang/lint/golint
	go get github.com/kisielk/errcheck
	go get github.com/matm/gocov-html
	go get golang.org/x/tools/cmd/goimports
	go get honnef.co/go/staticcheck/cmd/staticcheck

test:
	go test -v ./...

coverage:
	gocov test ./... > $(CURDIR)/coverage.out 2>/dev/null
	gocov report $(CURDIR)/coverage.out
	if test -z "$$CI"; then \
	  gocov-html $(CURDIR)/coverage.out > $(CURDIR)/coverage.html; \
	  if which open &>/dev/null; then \
	    open $(CURDIR)/coverage.html; \
	  fi; \
	fi

vet:
	go vet -v ./...

errors:
	errcheck -ignoretests -blank ./...

static:
	staticcheck ./...

lint:
	golint ./...

imports:
	goimports -l -w .

fmt:
	go fmt ./...

env:
	@go env

build:
	go build -v ./...

doc:
	godoc -http=:8080 -index

version:
	@go version
