.PHONY: build clean test all

OUTPUT = ./cnab/app/run
GO_SOURCES = $(shell find . -type f -name '*.go')
VERSION ?= $(shell cat VERSION)

GOBIN ?= $(shell go env GOPATH)/bin

all: build test

build: kab

test:
	GO111MODULE=on go test ./... -race -coverprofile=coverage.txt -covermode=atomic

check-mockery:
	@which mockery > /dev/null || (echo mockery not found: issue \"GO111MODULE=off go get -u  github.com/vektra/mockery/.../\" && false)

check-jq:
	@which jq > /dev/null || (echo jq not found: please install jq, eg \"brew install jq\" && false)

gen-mocks: check-mockery check-jq
	GO111MODULE=on mockery -output pkg/kustomize/mocks    -outpkg mockkustomize   -dir pkg/kustomize                                                                                               -name Kustomizer
	GO111MODULE=on mockery -output pkg/registry/mocks     -outpkg mockregistry    -dir pkg/registry                                                                                                -name Client
	GO111MODULE=on mockery -output pkg/kab/vendor_mocks/ext   -outpkg vendor_mocks_ext    -dir $(call source_of,k8s.io/apiextensions-apiserver)/pkg/client/clientset/clientset                             -name Interface
	GO111MODULE=on mockery -output pkg/kab/vendor_mocks/ext   -outpkg vendor_mocks_ext    -dir $(call source_of,k8s.io/apiextensions-apiserver)/pkg/client/clientset/clientset/typed/apiextensions/v1beta1 -name ApiextensionsV1beta1Interface
	GO111MODULE=on mockery -output pkg/kab/vendor_mocks/ext   -outpkg vendor_mocks_ext    -dir $(call source_of,k8s.io/apiextensions-apiserver)/pkg/client/clientset/clientset/typed/apiextensions/v1beta1 -name CustomResourceDefinitionInterface
	GO111MODULE=on mockery -output pkg/kab/vendor_mocks   -outpkg vendor_mocks    -dir $(call source_of,k8s.io/client-go)/kubernetes/typed/core/v1                                                 -name CoreV1Interface
	GO111MODULE=on mockery -output pkg/kab/vendor_mocks   -outpkg vendor_mocks    -dir $(call source_of,k8s.io/client-go)/kubernetes/typed/core/v1                                                 -name NodeInterface
	GO111MODULE=on mockery -output pkg/kab/vendor_mocks   -outpkg vendor_mocks    -dir $(call source_of,k8s.io/client-go)/kubernetes                                                               -name Interface

install: build
	cp $(OUTPUT) $(GOBIN)

kab: $(GO_SOURCES) VERSION
	GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(OUTPUT) -v

define source_of
	$(shell GO111MODULE=on go mod download -json | jq -r 'select(.Path == "$(1)").Dir' | tr '\\' '/'  2> /dev/null)
endef
