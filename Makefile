PROJECT_NAME := "karten"
PKG := "github.com/egregors/$(PROJECT_NAME)"
PKG_LIST := $(shell go list ${PKG}/... | grep -v /vendor/)
GO_FILES := $(shell find . -name '*.go' | grep -v /vendor/ | grep -v _test.go)

.PHONY: all test lint race bench run update-go-deps help

all: run

build: lint test
	@go build -o karten

install: build  ## Build executable and put it into $GOPATH
	cp -f ./karten ${GOPATH}/bin/

run:  ## Run the learn mode
	@go run main.go

add:  ## Run the add mode
	@go run main.go add

lint:  ## Lint the files
	@golangci-lint run --config .golangci.yml ./...

test:  ## Run unittests
	@go test -short ${PKG_LIST}

race:  ## Run data race detector
	@go test -race -short ${PKG_LIST}

bench:  ## Run benchmarks
	@go test -bench=.

update-go-deps:  ## Updating Go dependencies
	@echo ">> updating Go dependencies"
	@for m in $$(go list -mod=readonly -m -f '{{ if and (not .Indirect) (not .Main)}}{{.Path}}{{end}}' all); do \
		go get $$m; \
	done
	go mod tidy
ifneq (,$(wildcard vendor))
	go mod vendor
endif

## Help

help:  ## Show help message
	@IFS=$$'\n' ; \
	help_lines=(`fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//' | sed -e 's/##/:/'`); \
	printf "%s\n\n" "Usage: make [task]"; \
	printf "%-20s %s\n" "task" "help" ; \
	printf "%-20s %s\n" "------" "----" ; \
	for help_line in $${help_lines[@]}; do \
		IFS=$$':' ; \
		help_split=($$help_line) ; \
		help_command=`echo $${help_split[0]} | sed -e 's/^ *//' -e 's/ *$$//'` ; \
		help_info=`echo $${help_split[2]} | sed -e 's/^ *//' -e 's/ *$$//'` ; \
		printf '\033[36m'; \
		printf "%-20s %s" $$help_command ; \
		printf '\033[0m'; \
		printf "%s\n" $$help_info; \
	done
