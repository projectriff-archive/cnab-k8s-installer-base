.PHONY: build clean test all

OUTPUT = ./kab
GO_SOURCES = $(shell find . -type f -name '*.go')
VERSION ?= $(shell cat VERSION)

GOBIN ?= $(shell go env GOPATH)/bin

all: build test

build: $(OUTPUT)

test:
	GO111MODULE=on go test ./...

install: build
	cp $(OUTPUT) $(GOBIN)

$(OUTPUT): $(GO_SOURCES) VERSION
	GO111MODULE=on go build -o $(OUTPUT) -v
