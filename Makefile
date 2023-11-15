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

.PHONY: test
test:
	go test -v ./...

##################################################################################
# binaries to build

ARM_BIN_NAME       	:= reorderpath.Linux.arm71
DARWIN_BIN_NAME    	:= reorderpath.Darwin.x86_64
DARWIN_ARM_BIN_NAME := reorderpath.Darwin.arm64
LINUX_X86_BIN_NAME 	:= reorderpath.Linux.x86_64

ARM_BIN_PATH       	:= $(BIN)/$(ARM_BIN_NAME)
DARWIN_BIN_PATH    	:= $(BIN)/$(DARWIN_BIN_NAME)
LINUX_X86_BIN_PATH 	:= $(BIN)/$(LINUX_X86_BIN_NAME)
DARWIN_ARM_BIN_PATH := $(BIN)/$(DARWIN_ARM_BIN_NAME)

$(ARM_BIN_PATH): $(BIN)
	env GOOS=linux GOARCH=arm GOARM=7 go build -o $(ARM_BIN_PATH)

$(LINUX_X86_BIN_PATH): $(BIN)
	env GOOS=linux GOARCH=amd64 go build -o $(LINUX_X86_BIN_PATH)

$(DARWIN_BIN_PATH): $(BIN)
	env GOOS=darwin GOARCH=amd64 go build -o $(DARWIN_BIN_PATH)

$(DARWIN_ARM_BIN_PATH): $(BIN)
	env GOOS=darwin GOARCH=arm64 go build -o $(DARWIN_ARM_BIN_PATH)


BINARIES := $(ARM_BIN_PATH) $(LINUX_X86_BIN_PATH) $(DARWIN_BIN_PATH) $(DARWIN_ARM_BIN_PATH)

VERSION ?= $(shell git describe --tags --always --dirty --match=v* 2> /dev/null || echo v0)

.PHONY: clean
clean:
	rm -f $(BINARIES)

.PHONY: all
all: fmt lint | $(BIN) $(BINARIES)
