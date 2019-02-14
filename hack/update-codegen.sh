#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

# code-gen does not work with go modules yet :(
GO111MODULE_OLD=${GO111MODULE:-off}
GO111MODULE=off

SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
CODEGEN_BASE=${GOPATH}/src/k8s.io
CODEGEN_PKG=${CODEGEN_BASE}/code-generator

if [ ! -d ${CODEGEN_PKG} ]; then
    echo "code-generator does not exist, cloaning it into \$GOPATH"
    mkdir -p ${CODEGEN_BASE}
    pushd ${CODEGEN_BASE}
    git clone git@github.com:kubernetes/code-generator.git
    popd
fi



"${CODEGEN_PKG}"/generate-groups.sh "deepcopy,client" \
  github.com/projectriff/cnab-k8s-installer-base/pkg/client github.com/projectriff/cnab-k8s-installer-base/pkg/apis \
  kab:v1alpha1 \
  --go-header-file "${SCRIPT_ROOT}"/hack/copyright.go.txt

GO111MODULE=${GO111MODULE_OLD}
