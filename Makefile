# note: call scripts from /scripts
# stolen from https://vincent.bernat.ch/en/blog/2019-makefile-build-golang

.DEFAULT_GOAL := all

BIN = $(CURDIR)/bin
$(BIN):
	@mkdir -p $@
$(BIN)/%: | $(BIN)
	@tmp=$$(mktemp -d); \
		env GO111MODULE=off GOPATH=$$tmp GOBIN=$(BIN) go get $(PACKAGE) \
		|| ret=$$?; \
		rm -rf $$tmp ; exit $$ret

$(BIN)/golint: PACKAGE=golang.org/x/lint/golint

GOLINT = $(BIN)/golint
ifdef (ENFORCE_LINT)
lint: | $(GOLINT)
	$(GOLINT) -set_exit_status ./...
else
lint:
endif

.PHONY: fmt
fmt:
	gofmt -w -s .

#############################################################################
# test

TIMEOUT  = 20
PKGS     = $(or $(PKG),$(shell env GO111MODULE=on $(GO) list ./...))
TESTPKGS = $(shell env GO111MODULE=on $(GO) list -f \
		'{{ if or .TestGoFiles .XTestGoFiles }}{{ .ImportPath }}{{ end }}' \
		$(PKGS))

TEST_TARGETS := test-default test-bench test-short test-verbose test-race
test-bench:   ARGS=-run=__absolutelynothing__ -bench=.
test-short:   ARGS=-short
test-verbose: ARGS=-v
test-race:    ARGS=-race
$(TEST_TARGETS): test
check test tests: fmt lint
	go test -timeout $(TIMEOUT)s $(ARGS) $(TESTPKGS)

#############################################################################
# coverage

COVERAGE_MODE    = atomic
COVERAGE_PROFILE = $(COVERAGE_DIR)/profile.out
COVERAGE_XML     = $(COVERAGE_DIR)/coverage.xml
COVERAGE_HTML    = $(COVERAGE_DIR)/index.html
test-coverage-tools: | $(GOCOVMERGE) $(GOCOV) $(GOCOVXML) # ❶
test-coverage: COVERAGE_DIR := $(CURDIR)/test/coverage.$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
test-coverage: fmt lint test-coverage-tools
	@mkdir -p $(COVERAGE_DIR)/coverage
	@for pkg in $(TESTPKGS); do \ # ❷
		go test -coverpkg=$$(
				go list -f '{{ join .Deps "\n" }}' $$pkg | \
				grep '^$(MODULE)/' | \
				tr '\n' ','
			)$$pkg \
			-covermode=$(COVERAGE_MODE) \
			-coverprofile="$(COVERAGE_DIR)/coverage/`echo $$pkg | tr "/" "-"`.cover" $$pkg ;\
	done
	@$(GOCOVMERGE) $(COVERAGE_DIR)/coverage/*.cover > $(COVERAGE_PROFILE)
	@go tool cover -html=$(COVERAGE_PROFILE) -o $(COVERAGE_HTML)
	@$(GOCOV) convert $(COVERAGE_PROFILE) | $(GOCOVXML) > $(COVERAGE_XML)



VERSION ?= $(shell git describe --tags --always --dirty --match=v* 2> /dev/null || echo v0)
all: fmt lint | $(BIN)
	go build -o $(BIN)/reorderpath main.go
