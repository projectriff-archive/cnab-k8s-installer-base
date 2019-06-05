# Developer's Guide

## Prerequisites

You will need
* go 1.14 or later
* make
* docker

## Getting the code

Clone this repository
```$bash
$ go get -d github.com/projectriff/cnab-k8s-installer-base/...
$ cd $(go env GOPATH)/src/github.com/projectriff/cnab-k8s-installer-base
```

## Building locally with duffle

1. build the run tool
    ```bash
    $ make build
    ```
1. build the docker image
    ```bash
    $ docker build -t projectriff/cnab-k8s-installer-base:edge cnab
    ```
1. Building the cnab bundle with duffle now should use this docker image. example: building [cnab-riff](https://github.com/projectriff/cnab-riff) now with `duffle build .` 

## Building locally without duffle

1. Build the binary for your platform (change the Makefile to remove GOOS=linux etc.
1. Set an env var MANIFEST_FILE to point to the manifest file you want
1. Set an env var CNAB_ACTION to install or uninstall
1. You can also set the LOG_LEVEL to debug for a detailed output
1. run ./cnab/app/run
