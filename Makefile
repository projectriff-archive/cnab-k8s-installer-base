.PHONY: build clean test all

OUTPUT = ./cnab/app/run
GO_SOURCES = $(shell find . -type f -name '*.go')
VERSION ?= $(shell cat VERSION)

GOBIN ?= $(shell go env GOPATH)/bin

all: build test

build: kab

test:
	GO111MODULE=on go test ./...

check-mockery:
	@which mockery > /dev/null || (echo mockery not found: issue \"GO111MODULE=off go get -u  github.com/vektra/mockery/.../\" && false)

check-jq:
	@which jq > /dev/null || (echo jq not found: please install jq, eg \"brew install jq\" && false)

gen-mocks: check-mockery check-jq
	GO111MODULE=on mockery -output pkg/core/kustomize/mocks                -outpkg mockkustomize       -dir pkg/kustomize                                                                                               -name Kustomizer
	GO111MODULE=on mockery -output pkg/kab/vendor_mocks/mockextensions     -outpkg mockextensions      -dir $(call source_of,k8s.io/apiextensions-apiserver)/pkg/client/clientset/clientset                             -name Interface
	GO111MODULE=on mockery -output pkg/kab/vendor_mocks/mockextensions     -outpkg mockextensions      -dir $(call source_of,k8s.io/apiextensions-apiserver)/pkg/client/clientset/clientset/typed/apiextensions/v1beta1 -name ApiextensionsV1beta1Interface
	GO111MODULE=on mockery -output pkg/kab/vendor_mocks/mockextensions     -outpkg mockextensions      -dir $(call source_of,k8s.io/apiextensions-apiserver)/pkg/client/clientset/clientset/typed/apiextensions/v1beta1 -name CustomResourceDefinitionInterface

install: build
	cp $(OUTPUT) $(GOBIN)

kab: $(GO_SOURCES) VERSION
	GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(OUTPUT) -v
