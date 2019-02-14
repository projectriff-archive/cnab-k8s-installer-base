.PHONY: build clean test all

OUTPUT = ./cnab/app/run
GO_SOURCES = $(shell find . -type f -name '*.go')
VERSION ?= $(shell cat VERSION)

GOBIN ?= $(shell go env GOPATH)/bin

all: build test

build: kab

test:
	GO111MODULE=on go test ./...

install: build
	cp $(OUTPUT) $(GOBIN)

kab: $(GO_SOURCES) VERSION
	GO111MODULE=on go build -o $(OUTPUT) -v
