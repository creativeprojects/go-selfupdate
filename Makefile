# 
# Makefile for go-selfupdate
# 
GOCMD=go
GOBUILD=$(GOCMD) build
GOINSTALL=$(GOCMD) install
GORUN=$(GOCMD) run
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOTOOL=$(GOCMD) tool
GOGET=$(GOCMD) get
GOPATH?=`$(GOCMD) env GOPATH`

TESTS=. ./update
COVERAGE_FILE=coverage.out

BUILD_DATE=`date`
BUILD_COMMIT=`git rev-parse HEAD`

.PHONY: all test build coverage full-coverage clean toc staticcheck

all: test build

build:
		$(GOBUILD) -v ./...

test:
		$(GOTEST) -race -v $(TESTS)

coverage:
		$(GOTEST) -short -coverprofile=$(COVERAGE_FILE) $(TESTS)
		$(GOTOOL) cover -html=$(COVERAGE_FILE)

full-coverage:
		$(GOTEST) -coverprofile=$(COVERAGE_FILE) $(TESTS)
		$(GOTOOL) cover -html=$(COVERAGE_FILE)

clean:
		rm detect-latest-release go-get-release coverage.out
		$(GOCLEAN)

toc:
	go install github.com/ekalinin/github-markdown-toc.go
	go mod tidy
	cat README.md | github-markdown-toc.go --hide-footer

staticcheck:
	go get -u honnef.co/go/tools/cmd/staticcheck
	go mod tidy
	go run honnef.co/go/tools/cmd/staticcheck ./...
