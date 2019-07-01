.PHONY: build clean test all kab

OUTPUT = ./cnab/app/run
GO_SOURCES = $(shell find . -type f -name '*.go')
VERSION ?= $(shell cat VERSION)

GOBIN ?= $(shell go env GOPATH)/bin

all: build test verify-goimports

build: kab

test:
	GO111MODULE=on go test ./... -race -coverprofile=coverage.txt -covermode=atomic

check-goimports:
	@which goimports > /dev/null || (echo goimports not found: issue \"go get golang.org/x/tools/cmd/goimports\" && false)

goimports: check-goimports
	@goimports -w pkg main.go

verify-goimports: check-goimports
	@goimports -l pkg main.go | (! grep .) || (echo above files are not formatted correctly. please run \"make goimports\" && false)

kab: $(GO_SOURCES) VERSION
	GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -o $(OUTPUT) -v
