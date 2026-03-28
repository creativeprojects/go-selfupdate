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
GOBIN=$(shell $(GOCMD) env GOBIN)

ifeq ($(GOBIN),)
	GOBIN := $(GOPATH)/bin
endif

TESTS=. ./update
COVERAGE_FILE=coverage.txt

BUILD_DATE=`date`
BUILD_COMMIT=`git rev-parse HEAD`

README=README.md
TOC_START=<\!--ts-->
TOC_END=<\!--te-->
TOC_PATH=toc.md

.PHONY: all test build coverage full-coverage clean toc

all: test build

verify: ## Verify go installation
ifeq ($(GOPATH),)
	@echo "GOPATH not found, please check your go installation"
	exit 1
endif

$(GOBIN)/eget: verify
	@echo "[*] $@"
	GOBIN="$(GOBIN)" $(GOCMD) install -v github.com/zyedidia/eget@v1.3.4

$(GOBIN)/golangci-lint-v2: verify $(GOBIN)/eget
	@echo "[*] $@"
	"$(GOBIN)/eget" golangci/golangci-lint --tag v2.11.4 --asset=tar.gz --upgrade-only --to '$(GOBIN)/golangci-lint-v2'

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
		rm detect-latest-release go-get-release coverage.txt
		$(GOCLEAN)

toc:
	@echo "[*] $@"
	$(GOINSTALL) github.com/ekalinin/github-markdown-toc.go/cmd/gh-md-toc@latest
	cat ${README} | gh-md-toc --hide-footer > ${TOC_PATH}
	sed -i ".1" "/${TOC_START}/,/${TOC_END}/{//!d;}" "${README}"
	sed -i ".2" "/${TOC_START}/r ${TOC_PATH}" "${README}"
	rm ${README}.1 ${README}.2 ${TOC_PATH}

.PHONY: lint
lint: $(GOBIN)/golangci-lint-v2
	@echo "[*] $@"
	GOOS=darwin $(GOBIN)/golangci-lint-v2 run
	GOOS=linux $(GOBIN)/golangci-lint-v2 run
	GOOS=windows $(GOBIN)/golangci-lint-v2 run

.PHONY: fix
fix: $(GOBIN)/golangci-lint-v2
	@echo "[*] $@"
	$(GOCMD) mod tidy
	$(GOCMD) fix ./...
	GOOS=darwin $(GOBIN)/golangci-lint-v2 run --fix
	GOOS=linux $(GOBIN)/golangci-lint-v2 run --fix
	GOOS=windows $(GOBIN)/golangci-lint-v2 run --fix
