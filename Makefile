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

TESTS=./...
COVERAGE_FILE=coverage.out

BUILD_DATE=`date`
BUILD_COMMIT=`git rev-parse HEAD`

.PHONY: all test build coverage clean toc staticcheck

all: test build

build:
		$(GOBUILD) -v ./selfupdate
		$(GOBUILD) -v ./cmd/go-get-release
		$(GOBUILD) -v ./cmd/detect-latest-release

test:
		$(GOTEST) -race -v $(TESTS)

coverage:
		$(GOTEST) -coverprofile=$(COVERAGE_FILE) $(TESTS)
		$(GOTOOL) cover -html=$(COVERAGE_FILE)

clean:
		$(GOCLEAN)

toc:
	go install github.com/ekalinin/github-markdown-toc.go
	go mod tidy
	cat README.md | github-markdown-toc.go --hide-footer

staticcheck:
	go get -u honnef.co/go/tools/cmd/staticcheck
	go mod tidy
	go run honnef.co/go/tools/cmd/staticcheck ./...
